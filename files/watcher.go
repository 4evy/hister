package files

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/asciimoo/hister/config"
)

const debounceTime = 200 * time.Millisecond

// Watcher observes filtered directory trees and explicit files. Watches are
// installed before NewWatcher returns, so callers can scan without a gap
// between their initial import and subsequent notifications.
type Watcher struct {
	watcher *fsnotify.Watcher
	dirs    []*config.Directory
	paths   map[string]bool
	strict  bool
}

func NewWatcher(dirs []*config.Directory, paths []string) (*Watcher, error) {
	return newWatcher(dirs, paths, true)
}

func newWatcher(dirs []*config.Directory, paths []string, strict bool) (*Watcher, error) {
	native, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{watcher: native, paths: make(map[string]bool), strict: strict}
	for _, dir := range dirs {
		if dir == nil {
			continue
		}
		copy := *dir
		copy.Path, err = filepath.Abs(ExpandHome(dir.Path))
		if err != nil {
			_ = w.Close()
			return nil, err
		}
		w.dirs = append(w.dirs, &copy)
	}
	for _, path := range paths {
		path, err = filepath.Abs(ExpandHome(path))
		if err != nil {
			_ = w.Close()
			return nil, err
		}
		w.paths[path] = true
	}
	if err := w.scan(nil); err != nil {
		_ = w.Close()
		return nil, err
	}
	return w, nil
}

func (w *Watcher) Close() error { return w.watcher.Close() }

// The server keeps watching accessible directories when another directory
// fails. CLI setup reports incomplete coverage as an error instead.
func (w *Watcher) report(err error) error {
	if err == nil || w.strict {
		return err
	}
	log.Warn().Err(err).Msg("File watcher could not observe a path")
	return nil
}

func (w *Watcher) matches(path string) bool {
	if w.paths[path] {
		return true
	}
	for _, dir := range w.dirs {
		if DirectoryMatchesPath(dir, path) {
			return true
		}
	}
	return false
}

func (w *Watcher) matchesDirectory(path string) bool {
	for _, dir := range w.dirs {
		if !HasPathPrefix(path, dir.Path) {
			continue
		}
		if path == dir.Path {
			return true
		}
		rel, err := filepath.Rel(dir.Path, path)
		if err != nil {
			continue
		}
		excluded := false
		for part := range strings.SplitSeq(rel, string(filepath.Separator)) {
			if shouldSkipDir(part, dir.Excludes, dir.IncludeHidden) {
				excluded = true
				break
			}
		}
		if !excluded {
			return true
		}
	}
	return false
}

func (w *Watcher) scanDirectory(root string, callback func(string)) error {
	// WalkDir does not follow symbolic links, so register the root directly.
	if err := w.report(w.watcher.Add(root)); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return w.report(err)
		}
		if entry.IsDir() {
			if !w.matchesDirectory(path) {
				return filepath.SkipDir
			}
			if path == root {
				return nil
			}
			// Register before visiting children so writes during a scan are queued.
			return w.report(w.watcher.Add(path))
		}
		if callback != nil && w.matches(path) {
			callback(path)
		}
		return nil
	})
}

func (w *Watcher) scan(callback func(string)) error {
	for _, dir := range w.dirs {
		// Observe the parent as well to discover a replaced directory root.
		if err := w.watcher.Add(filepath.Dir(dir.Path)); err != nil {
			if err := w.report(fmt.Errorf("watch parent of %s: %w", dir.Path, err)); err != nil {
				return err
			}
		}
		if err := w.scanDirectory(dir.Path, callback); err != nil {
			return fmt.Errorf("watch directory %s: %w", dir.Path, err)
		}
	}
	for path := range w.paths {
		// Watching the parent survives editors saving via rename over the file.
		if err := w.watcher.Add(filepath.Dir(path)); err != nil {
			return fmt.Errorf("watch parent of %s: %w", path, err)
		}
		if callback != nil {
			callback(path)
		}
	}
	return nil
}

// Run calls callbacks serially. Callbacks should queue work and return promptly.
// Create and write events share a debounce window. No callbacks run after Run
// returns. Removal callbacks use the first matching directory's deletion policy
// and filters.
func (w *Watcher) Run(ctx context.Context, onChange, onRemove func(string)) error {
	pending := make(map[string]time.Time)
	ticker := time.NewTicker(debounceTime / 2)
	defer ticker.Stop()
	schedule := func(path string) { pending[path] = time.Now().Add(debounceTime) }
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			for path, due := range pending {
				if !now.Before(due) {
					delete(pending, path)
					onChange(path)
				}
			}
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				for path := range pending {
					if HasPathPrefix(path, event.Name) {
						delete(pending, path)
					}
				}
				// Descendant watches may survive a directory move on some systems.
				// Drop their old names before a replacement tree is registered.
				for _, path := range w.watcher.WatchList() {
					if HasPathPrefix(path, event.Name) {
						_ = w.watcher.Remove(path)
					}
				}
				dir := FindMatchingDir(w.dirs, event.Name)
				if onRemove != nil && dir != nil && dir.DeleteOnRemove && DirectoryMatchesPath(dir, event.Name) {
					onRemove(event.Name)
				}
			}
			if event.Has(fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() && w.matchesDirectory(event.Name) {
					if err := w.scanDirectory(event.Name, schedule); err != nil && !errors.Is(err, fs.ErrNotExist) {
						return err
					}
					continue
				}
			}
			if (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) && w.matches(event.Name) {
				schedule(event.Name)
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			if !errors.Is(err, fsnotify.ErrEventOverflow) {
				if err := w.report(err); err != nil {
					return err
				}
				continue
			}
			if err := w.scan(schedule); err != nil {
				return err
			}
		}
	}
}

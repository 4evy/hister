//go:build linux || darwin

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const serviceMarker = "Managed by nogon service install"

type userService struct {
	path, name, domain string
	run                serviceRunner
}

type serviceJob struct {
	state, path string // Empty state means the manager does not know this job.
}

func addPlatformServiceCommands(_ *cobra.Command) {}

func manageService(cmd *cobra.Command, platform string, run serviceRunner) error {
	s, err := newUserService(platform, run)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	action := cmd.Name()
	var content string
	if action == "install" {
		content, err = s.definition()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	// Serialize changes made by simultaneous CLI invocations. A directory lock
	// needs no auxiliary file and the kernel releases it when the process exits.
	exists := false
	lock, err := os.Open(dir)
	if err == nil {
		defer func() { _ = lock.Close() }()
		if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
			return fmt.Errorf("lock user service directory: %w", err)
		}
		exists, err = s.ownedDefinition()
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if exists && action == "install" {
		return fmt.Errorf("service already installed at %s", s.path)
	}
	ctx := cmd.Context()
	job, err := s.query(ctx, s.name)
	if err != nil {
		return err
	}
	if job.state != "" && (!exists || !s.matchesPath(job.path)) {
		return fmt.Errorf("%s is managed outside nogon service", s.name)
	}
	if !exists {
		if action == "install" || action == "status" {
			if err := s.checkOtherServices(ctx); err != nil {
				return err
			}
		}
		if action == "status" {
			return reportServiceStatus(cmd, "not installed")
		}
		if action == "uninstall" {
			return nil
		}
		if action != "install" {
			return errors.New("user service is not installed")
		}
	}
	if action == "status" {
		state := job.state
		if state == "" {
			state = "stopped"
		}
		return reportServiceStatus(cmd, state)
	}
	if action == "install" {
		if err := publishServiceDefinition(s.path, content); err != nil {
			return err
		}
	}
	if s.domain == "" {
		err = s.systemdControl(ctx, action)
	} else {
		err = s.launchdControl(ctx, action, job)
	}
	if err != nil {
		if action == "install" {
			return fmt.Errorf("saved %s but activation failed: %w. Run nogon service uninstall before retrying installation", s.path, err)
		}
		return err
	}
	if action == "uninstall" {
		if err := os.Remove(s.path); err != nil {
			return err
		}
		if s.domain == "" {
			_, err = s.run(ctx, "systemctl", "--user", "daemon-reload")
			if err != nil {
				return err
			}
		}
	}
	switch action {
	case "install":
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed user service at %s\n", s.path)
	case "uninstall":
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed user service at %s\n", s.path)
	}
	return err
}

func newUserService(platform string, run serviceRunner) (userService, error) {
	s := userService{run: run}
	var dir string
	if platform == "linux" {
		base := os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return s, err
			}
			base = filepath.Join(home, ".config")
		}
		if !filepath.IsAbs(base) {
			return s, errors.New("systemd user configuration directory must be absolute")
		}
		dir = filepath.Join(base, "systemd", "user")
		s.name = "nogon.service"
		s.path = filepath.Join(dir, s.name)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return s, err
		}
		dir = filepath.Join(home, "Library", "LaunchAgents")
		s.name = launchdServiceName
		s.domain = fmt.Sprintf("gui/%d", os.Getuid())
		s.path = filepath.Join(dir, s.name+".plist")
	}
	return s, nil
}

func (s userService) definition() (string, error) {
	def, err := serviceSettings("")
	if err != nil {
		return "", err
	}
	for _, path := range []string{def.Binary, def.Config, def.WorkingDir, def.Data} {
		if !utf8.ValidString(path) || strings.ContainsFunc(path, unicode.IsControl) || strings.ContainsAny(path, "\ufffe\uffff") {
			return "", errors.New("service paths must contain printable UTF-8 characters")
		}
	}
	if s.domain == "" {
		return systemdDefinition(def)
	}
	return launchdDefinition(def), nil
}

func (s userService) matchesPath(path string) bool {
	if s.path == path {
		return true
	}
	if filepath.Base(s.path) != filepath.Base(path) {
		return false
	}
	// The manager can report the canonical path of a symlinked parent directory.
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	ours, ownErr := filepath.EvalSymlinks(filepath.Dir(s.path))
	return err == nil && ownErr == nil && parent == ours
}

func (s userService) ownedDefinition() (bool, error) {
	st, err := os.Lstat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if st.Mode().IsRegular() {
		b, err := os.ReadFile(s.path)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(string(b), "# "+serviceMarker+"\n") || strings.HasPrefix(string(b), launchdHeader) {
			return true, nil
		}
	}
	return false, fmt.Errorf("%s is managed outside nogon service", s.path)
}

func publishServiceDefinition(path, content string) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".nogon-service-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_, writeErr := f.WriteString(content)
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return err
	}
	// Publish the complete definition atomically, only if the name is free.
	return os.Link(f.Name(), path)
}

func (s userService) query(ctx context.Context, name string) (serviceJob, error) {
	if s.domain == "" {
		return s.systemdQuery(ctx, name)
	}
	return s.launchdQuery(ctx, name)
}

func (s userService) checkOtherServices(ctx context.Context) error {
	names := []string{"hister.service", "sh.brew.hister.service"}
	if s.domain != "" {
		names = []string{"org.nix-community.home.hister", "org.nixos.hister", "sh.brew.hister"}
	}
	for _, name := range names {
		paths := []string{filepath.Join(filepath.Dir(s.path), name)}
		if s.domain != "" {
			paths = []string{filepath.Join(filepath.Dir(s.path), name+".plist"), filepath.Join("/Library/LaunchAgents", name+".plist")}
		}
		for _, path := range paths {
			if _, err := os.Lstat(path); err == nil {
				return fmt.Errorf("%s is managed outside nogon service", path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		job, err := s.query(ctx, name)
		if err != nil {
			return err
		}
		if job.state != "" {
			return fmt.Errorf("%s is managed outside nogon service", name)
		}
	}
	return nil
}

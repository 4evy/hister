package files

import (
	"path/filepath"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/testutil"
)

func TestHasPathPrefix(t *testing.T) {
	root := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	dir := filepath.Join(root, "docs")
	tests := []struct {
		name     string
		filePath string
		dirPath  string
		want     bool
	}{
		{name: "same directory", filePath: dir, dirPath: dir, want: true},
		{name: "child", filePath: filepath.Join(dir, "note.txt"), dirPath: dir, want: true},
		{name: "nested child", filePath: filepath.Join(dir, "notes", "note.txt"), dirPath: dir, want: true},
		{name: "sibling prefix", filePath: filepath.Join(root, "docs-other", "note.txt"), dirPath: dir},
		{name: "parent", filePath: root, dirPath: dir},
		{name: "root", filePath: root, dirPath: root, want: true},
		{name: "root child", filePath: filepath.Join(root, "note.txt"), dirPath: root, want: true},
		{name: "root descendant", filePath: filepath.Join(dir, "note.txt"), dirPath: root, want: true},
		{name: "relative path outside root", filePath: filepath.Join("docs", "note.txt"), dirPath: root},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPathPrefix(tt.filePath, tt.dirPath); got != tt.want {
				t.Fatalf("HasPathPrefix(%q, %q) = %t, want %t", tt.filePath, tt.dirPath, got, tt.want)
			}
		})
	}
}

func TestFindMatchingDirAtFilesystemRoot(t *testing.T) {
	root := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	dir := &config.Directory{
		Path:      root,
		Label:     "notes",
		Filetypes: []string{"txt", "md", "docx", "pdf", "htm", "html"},
	}
	for _, ext := range dir.Filetypes {
		t.Run(ext, func(t *testing.T) {
			path := filepath.Join(root, "docs", "note."+ext)
			if got := FindMatchingDir([]*config.Directory{dir}, path); got != dir {
				t.Fatalf("FindMatchingDir(%q) = %v, want root directory config", path, got)
			}
			if !DirectoryMatchesPath(dir, path) {
				t.Fatalf("DirectoryMatchesPath(%q) = false, want true", path)
			}
		})
	}
}

func TestDirectoryMatchesPath(t *testing.T) {
	root := t.TempDir()
	dir := &config.Directory{
		Path:      root,
		Filetypes: []string{"txt"},
		Excludes:  []string{"excluded*"},
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "matching file", path: filepath.Join(root, "notes", "note.txt"), want: true},
		{name: "wrong file type", path: filepath.Join(root, "notes", "note.go")},
		{name: "excluded file", path: filepath.Join(root, "notes", "excluded-note.txt")},
		{name: "excluded parent", path: filepath.Join(root, "excluded-dir", "note.txt")},
		{name: "hidden parent", path: filepath.Join(root, ".cache", "note.txt")},
		{name: "dependency parent", path: filepath.Join(root, "node_modules", "note.txt")},
		{name: "outside root", path: filepath.Join(t.TempDir(), "note.txt")},
		{name: "directory root", path: root},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirectoryMatchesPath(dir, tt.path); got != tt.want {
				t.Fatalf("DirectoryMatchesPath(%q) = %t, want %t", tt.path, got, tt.want)
			}
		})
	}

	dir.IncludeHidden = true
	if !DirectoryMatchesPath(dir, filepath.Join(root, ".cache", "note.txt")) {
		t.Fatal("hidden parent should match when include_hidden is enabled")
	}
	if !DirectoryMatchesPath(dir, filepath.Join(root, "node_modules", "note.txt")) {
		t.Fatal("dependency parent should match when include_hidden is enabled")
	}
}

func TestFindDirUser(t *testing.T) {
	testutil.InitModel(t)

	u1 := testutil.CreateUser(t, "alice")

	dirs := []*config.Directory{
		{Path: "/home/alice/docs", User: "alice"},
		{Path: "/home/charlie/docs", User: "charlie"},
		{Path: "/shared/docs", User: ""},
		{Path: "/tmp", User: ""},
	}

	tests := []struct {
		name    string
		path    string
		wantID  uint
		wantErr bool
	}{
		{
			name:    "file in alice's directory",
			path:    "/home/alice/docs/file.txt",
			wantID:  u1.ID,
			wantErr: false,
		},
		{
			name:    "file in nested alice directory",
			path:    "/home/alice/docs/subdir/file.txt",
			wantID:  u1.ID,
			wantErr: false,
		},
		{
			name:    "file in shared global directory",
			path:    "/shared/docs/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "file in tmp global directory",
			path:    "/tmp/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "file not in any directory",
			path:    "/other/path/file.txt",
			wantID:  0,
			wantErr: false,
		},
		{
			name:    "non-existent user in config",
			path:    "/home/charlie/docs/file.txt",
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := FindDirUser(dirs, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("FindDirUser(%q) expected error, got nil", tt.path)
				}
				return
			}
			if err != nil {
				t.Errorf("FindDirUser(%q) unexpected error: %v", tt.path, err)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("FindDirUser(%q) = %d, want %d", tt.path, gotID, tt.wantID)
			}
		})
	}
}

func TestFindDirUserWithHomeExpansion(t *testing.T) {
	testutil.InitModel(t)

	home := ExpandHome("~/")
	if home == "~/" {
		t.Skip("could not expand home directory")
	}

	testutil.CreateUser(t, "homeuser")

	dirs := []*config.Directory{
		{Path: "~/project", User: "homeuser"},
	}

	tests := []struct {
		name        string
		path        string
		wantNonZero bool
	}{
		{
			name:        "tilde path matches",
			path:        filepath.Join(home, "project", "file.txt"),
			wantNonZero: true,
		},
		{
			name:        "tilde path does not match",
			path:        filepath.Join(home, "other", "file.txt"),
			wantNonZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := FindDirUser(dirs, tt.path)
			if err != nil {
				t.Errorf("FindDirUser(%q) unexpected error: %v", tt.path, err)
				return
			}
			if tt.wantNonZero && gotID == 0 {
				t.Errorf("FindDirUser(%q) = 0, expected non-zero user ID", tt.path)
			}
			if !tt.wantNonZero && gotID != 0 {
				t.Errorf("FindDirUser(%q) = %d, expected 0", tt.path, gotID)
			}
		})
	}
}

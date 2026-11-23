//go:build linux || darwin

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestServiceInstallRejectsTransientSettings(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		env  string
	}{
		{"token flag", []string{"--token", "private-value"}, ""},
		{"Nogon environment", nil, "NOGON__APP__ACCESS_TOKEN"},
		{"legacy environment", nil, "HISTER__APP__ACCESS_TOKEN"},
		{"empty override", nil, "NOGON__APP__LOG_FILE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			data := filepath.Join(dir, "data")
			t.Setenv("NOGON_DATA_DIR", data)
			t.Setenv("XDG_CONFIG_HOME", dir)
			t.Setenv("HOME", dir)
			t.Setenv("PATH", dir)
			if tc.env != "" {
				value := "private-value"
				if tc.name == "empty override" {
					value = ""
				}
				t.Setenv(tc.env, value)
			}
			want := tc.env
			if want == "" {
				want = "--token"
			}
			out, stderr, err := executeInspectionWithStderr(t, append([]string{"service", "install"}, tc.args...)...)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("expected error naming %s, got %v", want, err)
			}
			if strings.Contains(out+stderr+err.Error(), "private-value") {
				t.Fatal("printed a credential")
			}
			if _, err := os.Stat(data); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("install initialized data: %v", err)
			}
		})
	}
}

func TestUserServiceLifecycle(t *testing.T) {
	for _, platform := range []string{"linux", "darwin"} {
		t.Run(platform, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", home)
			work := filepath.Join(home, "relative paths")
			if err := os.Mkdir(work, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Chdir(work)
			cfgPath := filepath.Join(work, "main.yml")
			data := filepath.Join(home, "data")
			t.Setenv("NOGON_CONFIG", cfgPath)
			t.Setenv("NOGON_DATA_DIR", data)
			if err := os.WriteFile(cfgPath, []byte("app:\n  access_token: private-value\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(home, "systemd/user/nogon.service")
			if platform == "darwin" {
				path = filepath.Join(home, "Library/LaunchAgents/eu.nogon.server.plist")
			}

			loaded, running := false, false
			enables := 0
			failStart := false
			run := func(_ context.Context, name string, args ...string) (string, error) {
				last := args[len(args)-1]
				switch name {
				case "systemctl":
					if !slices.Contains(args, "--user") {
						t.Fatal("contacted the system manager")
					}
					switch {
					case slices.Contains(args, "show"):
						if last != "nogon.service" || !loaded {
							return "LoadState=not-found\n", nil
						}
						state, pid := "inactive", 0
						if running {
							state, pid = "active", 123
						}
						return fmt.Sprintf("LoadState=loaded\nFragmentPath=%s\nActiveState=%s\nMainPID=%d\n", path, state, pid), nil
					case slices.Contains(args, "enable"):
						enables++
						loaded, running = true, slices.Contains(args, "--now")
						if failStart {
							running = false
							return "", errors.New("start failed")
						}
					case slices.Contains(args, "disable"):
						loaded, running = false, false
					case slices.Contains(args, "start"), slices.Contains(args, "restart"):
						running = true
					case slices.Contains(args, "stop"):
						running = false
					case slices.Contains(args, "daemon-reload"):
					default:
						t.Fatalf("unexpected systemctl args %v", args)
					}
				case "launchctl":
					switch args[0] {
					case "print":
						if !strings.HasSuffix(last, "/eu.nogon.server") || !loaded {
							return "Could not find service", errors.New("no service")
						}
						return fmt.Sprintf("job = {\n\tpath = %s\n\tstate = running\n\tpid = 123\n}\n", path), nil
					case "bootstrap", "kickstart":
						if failStart {
							return "", errors.New("start failed")
						}
						loaded, running = true, true
					case "bootout":
						loaded, running = false, false
					default:
						t.Fatalf("unexpected launchctl args %v", args)
					}
				default:
					t.Fatalf("unexpected manager %q", name)
				}
				return "", nil
			}
			execute := func(action string) (string, error) {
				var out bytes.Buffer
				root := &cobra.Command{Use: "nogon", SilenceErrors: true, SilenceUsage: true}
				root.AddCommand(newServiceCommand(platform, run))
				root.SetArgs([]string{"service", action})
				root.SetOut(&out)
				root.SetErr(&out)
				err := root.Execute()
				return out.String(), err
			}
			if out, err := execute("status"); ExitCode(err) != 4 || !strings.Contains(out, "not installed") {
				t.Fatalf("missing service: %q, %v", out, err)
			}
			if err := os.MkdirAll(data, 0o700); err != nil {
				t.Fatal(err)
			}
			indexFile := filepath.Join(data, "index.db")
			if err := os.WriteFile(indexFile, []byte("existing index"), 0o600); err != nil {
				t.Fatal(err)
			}

			if _, err := execute("install"); err != nil {
				t.Fatal(err)
			}
			if !running {
				t.Fatal("install did not start service")
			}
			definition, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{work, cfgPath, data, "listen"} {
				if !strings.Contains(string(definition), want) {
					t.Errorf("definition missing %q:\n%s", want, definition)
				}
			}
			if strings.Contains(string(definition), "private-value") {
				t.Fatal("persisted a credential in the service definition")
			}
			if _, err := execute("install"); err == nil {
				t.Fatal("replaced an existing definition")
			}
			if err := os.Remove(cfgPath); err != nil {
				t.Fatal(err)
			}
			for _, action := range []string{"stop", "start", "restart", "status", "uninstall"} {
				out, err := execute(action)
				if err != nil {
					t.Fatalf("%s: %v", action, err)
				}
				if action == "status" && !strings.Contains(out, "running") {
					t.Errorf("status: %q", out)
				}
				if running != (action != "stop" && action != "uninstall") {
					t.Fatalf("running=%v after %s", running, action)
				}
			}
			if platform == "linux" && enables != 1 {
				t.Fatalf("expected only install to enable login startup, got %d enables", enables)
			}
			if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("service definition remains: %v", err)
			}
			if b, err := os.ReadFile(indexFile); err != nil || string(b) != "existing index" {
				t.Fatalf("uninstall changed indexed data: %v", err)
			}
			if err := os.WriteFile(cfgPath, []byte("app: {}"), 0o600); err != nil {
				t.Fatal(err)
			}
			failStart = true
			if _, err := execute("install"); err == nil || !strings.Contains(err.Error(), "start failed") {
				t.Fatalf("activation failure: %v", err)
			}
			if _, err := execute("uninstall"); err != nil {
				t.Fatalf("cleanup after activation failure: %v", err)
			}
			failStart = false
			if _, err := execute("install"); err != nil {
				t.Fatalf("retry after activation failure: %v", err)
			}
		})
	}
}

func TestNativeServiceDefinitions(t *testing.T) {
	dir := t.TempDir()
	def := serviceDefinition{Binary: filepath.Join(dir, `no$gon% 1`), WorkingDir: dir, Config: filepath.Join(dir, `$main & config%.yml`), Data: filepath.Join(dir, "Tom's <data>")}
	if err := os.WriteFile(def.Binary, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Run("systemd", func(t *testing.T) {
		analyze, err := exec.LookPath("systemd-analyze")
		if err != nil {
			t.Skip("systemd-analyze not installed")
		}
		unit, err := systemdDefinition(def)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "nogon.service")
		if err := os.WriteFile(path, []byte(unit), 0o600); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.CommandContext(t.Context(), analyze, "verify", path).CombinedOutput(); err != nil {
			t.Fatalf("invalid systemd unit: %v\n%s", err, out)
		}
	})
	t.Run("launchd", func(t *testing.T) {
		path := filepath.Join(dir, "eu.nogon.server.plist")
		if err := os.WriteFile(path, []byte(launchdDefinition(def)), 0o600); err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS == "darwin" {
			if out, err := exec.CommandContext(t.Context(), "/usr/bin/plutil", "-lint", path).CombinedOutput(); err != nil {
				t.Fatalf("invalid property list: %v\n%s", err, out)
			}
		}
		python, err := exec.LookPath("python3")
		if err != nil {
			t.Skip("python3 not installed")
		}
		const check = "import plistlib, sys\nwith open(sys.argv[1], 'rb') as f: job = plistlib.load(f)\nassert job['ProgramArguments'] == [sys.argv[2], 'listen', '--config', sys.argv[3]]\nassert job['WorkingDirectory'] == sys.argv[4]\nassert job['EnvironmentVariables']['NOGON_DATA_DIR'] == sys.argv[5]"
		if out, err := exec.CommandContext(t.Context(), python, "-c", check, path, def.Binary, def.Config, def.WorkingDir, def.Data).CombinedOutput(); err != nil {
			t.Fatalf("decoded launchd configuration: %v\n%s", err, out)
		}
	})
}

func TestServiceStatusChecksTheManagedJob(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "config")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", link)
	file := filepath.Join(dir, "systemd/user/nogon.service")
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("# "+serviceMarker+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, fragment, active string
		pid, code              int
	}{
		{"running through symlinked config directory", file, "active", 12, 0},
		{"waiting between restarts", file, "activating", 0, 3},
		{"other manager owns loaded unit", filepath.Join(dir, "external.service"), "active", 12, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			group := newServiceCommand("linux", func(_ context.Context, _ string, args ...string) (string, error) {
				if !slices.Contains(args, "show") {
					t.Fatalf("status modified a service: %v", args)
				}
				return fmt.Sprintf("LoadState=loaded\nActiveState=%s\nMainPID=%d\nFragmentPath=%s\n", tc.active, tc.pid, tc.fragment), nil
			})
			group.SetOut(&out)
			group.SetErr(&out)
			group.SetArgs([]string{"status"})
			if err := group.Execute(); ExitCode(err) != tc.code {
				t.Fatalf("exit %d, want %d: %v (%s)", ExitCode(err), tc.code, err, &out)
			}
			if tc.code == 3 && strings.Contains(out.String(), "running") {
				t.Fatalf("waiting job reported as running: %s", &out)
			}
		})
	}
}

func TestServiceRefusesForeignDefinitions(t *testing.T) {
	for _, tc := range []struct{ platform, path string }{
		{"linux", "systemd/user/nogon.service"},
		{"linux", "systemd/user/hister.service"},
		{"darwin", "Library/LaunchAgents/eu.nogon.server.plist"},
		{"darwin", "Library/LaunchAgents/org.nix-community.home.hister.plist"},
		{"darwin", "Library/LaunchAgents/sh.brew.hister.plist"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			t.Setenv("HOME", dir)
			t.Setenv("XDG_CONFIG_HOME", dir)
			t.Setenv("NOGON_DATA_DIR", filepath.Join(dir, "data"))
			path := filepath.Join(dir, tc.path)
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(dir, "foreign"), path); err != nil {
				t.Fatal(err)
			}
			group := newServiceCommand(tc.platform, func(_ context.Context, manager string, args ...string) (string, error) {
				if slices.Contains(args, "show") {
					return "LoadState=not-found\n", nil
				}
				if args[0] == "print" {
					return "Could not find service", errors.New("absent")
				}
				t.Fatalf("modified foreign service via %s %v", manager, args)
				return "", nil
			})
			group.SetArgs([]string{"install"})
			if err := group.Execute(); err == nil || !strings.Contains(err.Error(), "managed outside") {
				t.Fatalf("foreign definition: %v", err)
			}
			if _, err := os.Readlink(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

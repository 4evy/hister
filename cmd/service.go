// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type serviceRunner func(context.Context, string, ...string) (string, error)

func nativeServiceCommand(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.Env = append(c.Environ(), "LC_ALL=C")
	b, err := c.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(b)))
	}
	return string(b), nil
}

func newServiceCommand(platform string, run serviceRunner) *cobra.Command {
	group := &cobra.Command{
		Use: "service", Short: "Manage your background Nogon server",
		PersistentPreRun: skipRuntimeInitialization,
	}
	setCommandScope(group, executionScopeLocal)
	group.Annotations[applicableFlagsAnnotation] = ""
	installHelp := "Install and start Nogon for the current user"
	if platform == "windows" {
		installHelp = `Install and start the Windows service as NT SERVICE\Nogon`
	}
	for _, action := range []struct{ name, help string }{
		{"install", installHelp},
		{"uninstall", "Stop Nogon and remove its service definition"},
		{"start", "Start the installed service"},
		{"stop", "Stop the installed service"},
		{"restart", "Restart the installed service"},
		{"status", "Show the state of the installed service"},
	} {
		command := &cobra.Command{
			Use: action.name, Short: action.help, Args: cobra.NoArgs, SilenceUsage: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				for _, name := range scopedGlobalFlagNames {
					if name == "config" && cmd.Name() == "install" {
						continue
					}
					if cmd.Flags().Changed(name) {
						return fmt.Errorf("configure server settings in the configuration file because --%s does not apply to service %s", name, cmd.Name())
					}
				}
				return manageService(cmd, platform, run)
			},
		}
		setCommandScope(command, executionScopeLocal)
		command.Annotations[applicableFlagsAnnotation] = ""
		if action.name == "install" {
			command.Annotations[applicableFlagsAnnotation] = "config"
		}
		group.AddCommand(command)
	}
	addPlatformServiceCommands(group)
	configureScopeGroups(group)
	return group
}

type serviceDefinition struct {
	Binary, Config, Data, WorkingDir string
}

func serviceSettings(machineData string) (serviceDefinition, error) {
	var def serviceDefinition
	c, name, err := inspectConfig()
	if err != nil {
		return def, err
	}
	var transient []string
	for path, source := range c.Sources() {
		if source.Kind == "environment" && path != "app.directory" {
			transient = append(transient, source.Name)
		}
	}
	if len(transient) != 0 {
		slices.Sort(transient)
		return def, fmt.Errorf("put settings from %s in the configuration file before installing the service", strings.Join(slices.Compact(transient), ", "))
	}
	if machineData != "" {
		for _, dir := range c.Indexer.Directories {
			if strings.HasPrefix(dir.Path, "~/") {
				return def, fmt.Errorf("set an absolute path for watched directory %q before installing under a different account", dir.Path)
			}
		}
	}
	def.Config = name
	def.WorkingDir, err = os.Getwd()
	if err != nil {
		return def, err
	}
	def.Data, err = filepath.Abs(c.App.Directory)
	if err != nil {
		return def, err
	}
	if machineData != "" && c.Source("app.directory").Kind == "default" {
		populated, err := c.HasStoredData(def.Data)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return def, fmt.Errorf("inspect data directory: %w", err)
		}
		if !populated {
			def.Data = machineData
		}
	}
	def.Binary, err = os.Executable()
	if err != nil {
		return def, err
	}
	// Keep a stable invocation symlink across package upgrades when it still
	// points at the running executable.
	if invoked, err := exec.LookPath(os.Args[0]); err == nil {
		a, errA := os.Stat(invoked)
		b, errB := os.Stat(def.Binary)
		if errA == nil && errB == nil && os.SameFile(a, b) {
			def.Binary, err = filepath.Abs(invoked)
			if err != nil {
				return def, err
			}
		}
	}
	return def, nil
}

type serviceStatusCode int

func (c serviceStatusCode) Error() string { return fmt.Sprintf("service status code %d", c) }

func reportServiceStatus(cmd *cobra.Command, state string, details ...string) error {
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), state); err != nil {
		return err
	}
	for _, line := range details {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
			return err
		}
	}
	if state == "running" {
		return nil
	}
	cmd.SilenceErrors = true // The state has already been written to stdout.
	if state == "not installed" {
		return serviceStatusCode(4)
	}
	return serviceStatusCode(3)
}

//go:build !linux && !darwin && !windows

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func addPlatformServiceCommands(_ *cobra.Command) {}

func manageService(_ *cobra.Command, _ string, _ serviceRunner) error {
	return errors.New("service management requires Windows, systemd on Linux or launchd on macOS")
}

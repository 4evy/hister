//go:build !linux && !darwin

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func manageUserService(_ *cobra.Command, _ string, _ serviceRunner) error {
	return errors.New("user services require systemd on Linux or launchd on macOS")
}

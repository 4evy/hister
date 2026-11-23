//go:build linux || darwin

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func systemdDefinition(d serviceDefinition) (string, error) {
	// WorkingDirectory is a single literal path rather than an argv token.
	// A trailing backslash continues the unit line and trailing whitespace is stripped.
	if strings.HasSuffix(d.WorkingDir, "\\") || strings.HasSuffix(d.WorkingDir, " ") {
		return "", errors.New("systemd requires an installation working directory whose name does not end in a space or backslash")
	}
	literal := func(s string) string { return strings.ReplaceAll(s, "%", "%%") }
	// The ':' executable prefix disables environment substitution in all arguments.
	command := strconv.Quote(":"+literal(d.Binary)) + " listen"
	if d.Config != "" {
		command += " --config " + strconv.Quote(literal(d.Config))
	}
	return fmt.Sprintf(`# %s
[Unit]
Description=Nogon search server

[Service]
Type=exec
ExecStart=%s
WorkingDirectory=%s
Environment=%s
Restart=on-failure
UMask=0077

[Install]
WantedBy=default.target
`, serviceMarker, command, literal(d.WorkingDir), strconv.Quote("NOGON_DATA_DIR="+literal(d.Data))), nil
}

func (s userService) systemdQuery(ctx context.Context, name string) (serviceJob, error) {
	out, err := s.run(ctx, "systemctl", "--user", "show", "--property=LoadState,ActiveState,MainPID,FragmentPath", name)
	values := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			values[key] = value
		}
	}
	if values["LoadState"] == "not-found" && (values["ActiveState"] == "" || values["ActiveState"] == "inactive") {
		return serviceJob{}, nil
	}
	if err != nil {
		return serviceJob{}, err
	}
	if values["LoadState"] == "" {
		return serviceJob{}, fmt.Errorf("systemctl did not report the load state of %s", name)
	}
	state := values["ActiveState"]
	pid, _ := strconv.Atoi(values["MainPID"])
	switch {
	case (state == "active" || state == "reloading") && pid > 0:
		state = "running"
	case state == "inactive" || state == "active":
		state = "stopped"
	case state == "":
		state = values["LoadState"]
	}
	return serviceJob{state: state, path: values["FragmentPath"]}, nil
}

func (s userService) systemdControl(ctx context.Context, action string) error {
	args := []string{"--user"}
	switch action {
	case "install":
		args = append(args, "enable", "--now", s.path)
	case "uninstall":
		args = append(args, "disable", "--now", s.name)
	default:
		args = append(args, action, s.name)
	}
	_, err := s.run(ctx, "systemctl", args...)
	return err
}

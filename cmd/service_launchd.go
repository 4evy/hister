//go:build linux || darwin

// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

const (
	launchdHeader      = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- " + serviceMarker + " -->\n"
	launchdServiceName = "eu.nogon.server"
)

func launchdDefinition(d serviceDefinition) string {
	xmlString := func(v string) string { return "<string>" + html.EscapeString(v) + "</string>" }
	args := xmlString(d.Binary) + xmlString("listen")
	if d.Config != "" {
		args += xmlString("--config") + xmlString(d.Config)
	}
	return fmt.Sprintf(`%s<plist version="1.0">
<dict>
  <key>Label</key>%s
  <key>ProgramArguments</key><array>%s</array>
  <key>WorkingDirectory</key>%s
  <key>EnvironmentVariables</key><dict><key>NOGON_DATA_DIR</key>%s</dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict>
  <key>Umask</key><integer>63</integer>
</dict>
</plist>
`, launchdHeader, xmlString(launchdServiceName), args, xmlString(d.WorkingDir), xmlString(d.Data))
}

func (s userService) launchdQuery(ctx context.Context, name string) (serviceJob, error) {
	out, err := s.run(ctx, "launchctl", "print", s.domain+"/"+name)
	if err != nil {
		if strings.Contains(out, "Could not find service") {
			return serviceJob{}, nil
		}
		return serviceJob{}, err
	}
	job := serviceJob{}
	var pid int
	for line := range strings.SplitSeq(out, "\n") {
		// Nested resource dictionaries also have a state. Read the job's keys only.
		line, top := strings.CutPrefix(line, "\t")
		if !top {
			continue
		}
		key, value, _ := strings.Cut(line, " = ")
		switch key {
		case "path":
			job.path = value
		case "state":
			job.state = value
		case "pid":
			pid, _ = strconv.Atoi(value)
		}
	}
	if job.path == "" || job.state == "" {
		return serviceJob{}, fmt.Errorf("launchctl did not report the path and state of %s", name)
	}
	if job.state == "running" && pid == 0 {
		job.state = "starting"
	}
	return job, nil
}

func (s userService) launchdControl(ctx context.Context, action string, job serviceJob) error {
	target := s.domain + "/" + s.name
	if job.state != "" && (action == "stop" || action == "uninstall" || action == "restart") {
		if _, err := s.run(ctx, "launchctl", "bootout", target); err != nil {
			return err
		}
		// bootout can return before launchd has finished unloading the job.
		wait, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		for {
			status, err := s.query(wait, s.name)
			if err != nil {
				return err
			}
			if status.state == "" {
				break
			}
			select {
			case <-wait.Done():
				return fmt.Errorf("waiting for %s to stop: %w", s.name, wait.Err())
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	if action == "stop" || action == "uninstall" {
		return nil
	}
	if action == "start" && job.state != "" {
		if job.state == "running" {
			return nil
		}
		_, err := s.run(ctx, "launchctl", "kickstart", target)
		return err
	}
	_, err := s.run(ctx, "launchctl", "bootstrap", s.domain, s.path)
	return err
}

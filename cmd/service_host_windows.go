// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/asciimoo/nogon/config"
)

type windowsServiceHost struct{ serviceDefinition }

func (h windowsServiceHost) Execute(_ []string, requests <-chan svc.ChangeRequest, updates chan<- svc.Status) (bool, uint32) {
	pending := func(state svc.State) svc.Status {
		return svc.Status{State: state, CheckPoint: 1, WaitHint: 30_000}
	}
	status := pending(svc.StartPending)
	updates <- status
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	result := make(chan error, 1)
	go func() { result <- h.serve(ctx, func() { close(ready) }) }()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-result:
			if err != nil {
				if log, openErr := eventlog.Open(windowsServiceName); openErr == nil {
					_ = log.Error(1, h.diagnostic(err))
					_ = log.Close()
				}
				return true, 1
			}
			return false, 0
		case <-ready:
			ready = nil
			if status.State == svc.StartPending {
				status = svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
				updates <- status
			}
		case req := <-requests:
			switch req.Cmd {
			case svc.Interrogate:
				updates <- status
			case svc.Stop, svc.Shutdown:
				status = pending(svc.StopPending)
				updates <- status
				cancel()
			}
		case <-ticker.C:
			if status.State != svc.Running {
				status.CheckPoint++
				updates <- status
			}
		}
	}
}

// The Application log is readable outside the service account. Driver and
// extractor errors can contain connection strings or other configuration values.
func (h windowsServiceHost) diagnostic(err error) string {
	if pathErr, ok := errors.AsType[*os.PathError](err); ok {
		return pathErr.Error()
	}
	if netErr, ok := errors.AsType[*net.OpError](err); ok && netErr.Op == "listen" {
		return netErr.Error()
	}
	if field, ok := errors.AsType[config.FieldError](err); ok {
		return fmt.Sprintf("Check setting %s in configuration %q.", field.Path, h.Config)
	}
	return fmt.Sprintf("Nogon failed to start or serve using data directory %q. Check the database, index, and server settings in configuration %q.", h.Data, h.Config)
}

func (h windowsServiceHost) serve(ctx context.Context, ready func()) error {
	if err := os.Chdir(h.WorkingDir); err != nil {
		return err
	}
	// If a volume is offline or the data was moved, do not create a new index.
	info, err := os.Stat(h.Data)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("expected a data directory at %q", h.Data)
	}
	data := config.Override{
		Path: "app.directory", Value: h.Data,
		Source: config.ValueSource{Kind: "service", Name: "Windows service"},
	}
	cfg, err = config.Load(h.Config, data)
	if err != nil {
		return err
	}
	configureRuntime()
	return serve(ctx, ready, data)
}

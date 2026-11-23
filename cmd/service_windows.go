// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	windowsServiceName        = "Nogon"
	windowsServiceAccount     = `NT SERVICE\Nogon`
	windowsServiceDescription = "Nogon search server. Managed by nogon service install."
)

// These entry points carry the selected paths across two identity changes:
// from the installing user to the administrator, and from SCM to the service.
func addPlatformServiceCommands(group *cobra.Command) {
	for _, action := range []string{"register", "run"} {
		command := &cobra.Command{
			Use: action, Hidden: true, Args: cobra.RangeArgs(2, 3), SilenceUsage: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				for _, path := range args {
					if !filepath.IsAbs(path) {
						return fmt.Errorf("service path must be absolute: %q", path)
					}
				}
				def := serviceDefinition{Data: args[0], WorkingDir: args[1]}
				if len(args) == 3 {
					def.Config = args[2]
				}
				if action == "run" {
					return svc.Run(windowsServiceName, windowsServiceHost{def})
				}
				var err error
				def.Binary, err = os.Executable()
				if err != nil {
					return err
				}
				return installWindowsService(cmd, def)
			},
		}
		setCommandScope(command, executionScopeLocal)
		group.AddCommand(command)
	}
}

func (d serviceDefinition) windowsArgs(action string) []string {
	args := []string{"service", action, d.Data, d.WorkingDir}
	if d.Config != "" {
		args = append(args, d.Config)
	}
	return args
}

func machineDataDirectory() (string, error) {
	root, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, windowsServiceName), nil
}

func manageService(cmd *cobra.Command, _ string, _ serviceRunner) error {
	action := cmd.Name()
	if action == "install" {
		data, err := machineDataDirectory()
		if err != nil {
			return err
		}
		def, err := serviceSettings(data)
		if err != nil {
			return err
		}
		return installWindowsService(cmd, def)
	}

	manager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseServiceHandle(manager) }()
	access := uint32(windows.SERVICE_QUERY_CONFIG | windows.SERVICE_QUERY_STATUS)
	switch action {
	case "start":
		access |= windows.SERVICE_START
	case "stop":
		access |= windows.SERVICE_STOP
	case "restart":
		access |= windows.SERVICE_STOP | windows.SERVICE_START
	case "uninstall":
		access |= windows.SERVICE_STOP | windows.DELETE
	}
	name, _ := windows.UTF16PtrFromString(windowsServiceName)
	handle, err := windows.OpenService(manager, name, access)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		switch action {
		case "status":
			return reportServiceStatus(cmd, "not installed")
		case "uninstall":
			return nil
		}
		return errors.New("the Nogon service is not installed")
	}
	if err != nil {
		return fmt.Errorf("open Nogon service for %s: %w", action, err)
	}
	service := &mgr.Service{Handle: handle, Name: windowsServiceName}
	defer func() { _ = service.Close() }()
	settings, err := service.Config()
	if err != nil {
		return err
	}
	args, err := windows.DecomposeCommandLine(settings.BinaryPathName)
	if err != nil || settings.Description != windowsServiceDescription || len(args) < 5 || len(args) > 6 || args[1] != "service" || args[2] != "run" {
		return errors.New("the Nogon service is managed outside nogon service")
	}

	if action == "status" {
		status, err := service.Query()
		if err != nil {
			return err
		}
		return reportServiceStatus(cmd, windowsStateName(status.State), "Account: "+settings.ServiceStartName)
	}
	if action != "start" {
		if err := setWindowsServiceState(cmd.Context(), service, svc.Stopped); err != nil {
			return err
		}
	}
	switch action {
	case "uninstall":
		return service.Delete()
	case "restart", "start":
		return setWindowsServiceState(cmd.Context(), service, svc.Running)
	}
	return nil
}

func installWindowsService(cmd *cobra.Command, def serviceDefinition) error {
	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CREATE_SERVICE)
	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		// Build single quoted PowerShell literals. Represent typographic quotes
		// by code point so their interpretation cannot depend on the PS version.
		escape := strings.NewReplacer("'", "''", "‘", "'+[char]0x2018+'", "’", "'+[char]0x2019+'", "‚", "'+[char]0x201a+'", "‛", "'+[char]0x201b+'")
		args := append([]string{def.Binary}, def.windowsArgs("register")...)
		for i, arg := range args {
			args[i] = "'" + escape.Replace(arg) + "'"
			if strings.ContainsAny(arg, "‘’‚‛") {
				args[i] = "(" + args[i] + ")"
			}
		}
		return fmt.Errorf("to install with the selected paths, run in an administrator PowerShell window:\n& %s", strings.Join(args, " "))
	}
	if err != nil {
		return err
	}
	manager := &mgr.Mgr{Handle: handle}
	defer func() { _ = manager.Disconnect() }()
	service, err := manager.CreateService(windowsServiceName, def.Binary, mgr.Config{
		DisplayName: windowsServiceName, Description: windowsServiceDescription,
		ServiceStartName: windowsServiceAccount, SidType: windows.SERVICE_SID_TYPE_UNRESTRICTED,
		StartType: mgr.StartDisabled, ErrorControl: mgr.ErrorNormal,
	}, def.windowsArgs("run")...)
	if errors.Is(err, windows.ERROR_SERVICE_EXISTS) {
		return errors.New("a Windows service named Nogon is already installed")
	}
	if err != nil {
		return fmt.Errorf("create Nogon service: %w", err)
	}
	defer func() { _ = service.Close() }()
	if err := prepareWindowsServiceData(def.Data); err != nil {
		return errors.Join(fmt.Errorf("prepare data directory %q: %w", def.Data, err), service.Delete())
	}
	// Make the service startable only after its data directory is ready.
	if err := windows.ChangeServiceConfig(service.Handle, windows.SERVICE_NO_CHANGE, mgr.StartAutomatic, windows.SERVICE_NO_CHANGE, nil, nil, nil, nil, nil, nil, nil); err != nil {
		return errors.Join(err, service.Delete())
	}
	if err := setWindowsServiceState(cmd.Context(), service, svc.Running); err != nil {
		return fmt.Errorf("installed Nogon but startup failed: %w. Verify %s can read the executable, configuration %q and working directory %q, and modify data in %q. After fixing access, run nogon service start", err, windowsServiceAccount, def.Config, def.WorkingDir, def.Data)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed Windows service %s using %s\n", windowsServiceName, def.Data)
	return err
}

// Only create the dedicated machine directory. An existing installation keeps
// its permissions. A standard user can precreate a name under ProgramData.
func prepareWindowsServiceData(path string) error {
	machine, err := machineDataDirectory()
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Clean(path), machine) {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("provide an existing data directory %q and grant %s Modify access", path, windowsServiceAccount)
		}
		return nil
	}
	sid, _, _, err := windows.LookupSID("", windowsServiceAccount)
	if err != nil {
		return err
	}
	// SYSTEM and Administrators have full control. The service can modify
	// data, including files and directories created later, without changing ACLs.
	sd, err := windows.SecurityDescriptorFromString("O:BAD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1301bf;;;" + sid.String() + ")")
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{SecurityDescriptor: sd}
	sa.Length = uint32(unsafe.Sizeof(sa))
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	err = windows.CreateDirectory(name, &sa)
	if !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	// Inspect the directory itself, holding it against replacement while
	// checking its type and security descriptor.
	handle, err := windows.CreateFile(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("expected a local data directory at %q", path)
	}
	existing, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	if !privateMachineDirectory(existing, sid) {
		return fmt.Errorf("check the owner and permissions on %q: the machine directory must be owned by Administrators or SYSTEM and accessible only to them and %s", path, windowsServiceAccount)
	}
	return nil
}

func privateMachineDirectory(sd *windows.SECURITY_DESCRIPTOR, serviceSID *windows.SID) bool {
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return false
	}
	admin := func(sid *windows.SID) bool {
		return sid.IsWellKnown(windows.WinBuiltinAdministratorsSid) || sid.IsWellKnown(windows.WinLocalSystemSid)
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !admin(owner) {
		return false
	}
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil {
		return false
	}
	for i := uint16(0); i < dacl.AceCount; i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, uint32(i), &ace); err != nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return false
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !admin(sid) && !sid.Equals(serviceSID) {
			return false
		}
	}
	return true
}

func windowsStateName(state svc.State) string {
	switch state {
	case svc.Running:
		return "running"
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "starting"
	case svc.StopPending:
		return "stopping"
	case svc.Paused:
		return "paused"
	case svc.PausePending:
		return "pausing"
	case svc.ContinuePending:
		return "resuming"
	default:
		return fmt.Sprintf("unknown (%d)", state)
	}
}

func setWindowsServiceState(ctx context.Context, service *mgr.Service, target svc.State) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	requested := false
	for {
		status, err := service.Query()
		if err != nil {
			return err
		}
		switch status.State {
		case target:
			return nil
		case svc.Stopped:
			if requested {
				return fmt.Errorf("service stopped during startup (Windows error %d, service error %d); see the Application event log", status.Win32ExitCode, status.ServiceSpecificExitCode)
			}
			err = service.Start()
			if err != nil && !errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
				return err
			}
			requested = true
		case svc.Paused, svc.Running:
			if target != svc.Stopped {
				return errors.New("service is paused; resume it in Windows Services")
			}
			if !requested {
				_, err = service.Control(svc.Stop)
				if err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
					return err
				}
				requested = true
			}
		case svc.StartPending:
			if target == svc.Running {
				requested = true
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for %s: %w", windowsStateName(target), ctx.Err())
		case <-ticker.C:
		}
	}
}

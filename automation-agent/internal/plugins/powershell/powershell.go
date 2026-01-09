package powershell

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/automation-platform/agent/internal/probe"
)

// PowerShellPlugin executes PowerShell scripts
type PowerShellPlugin struct{}

// New creates a new PowerShell plugin
func New() *PowerShellPlugin {
	return &PowerShellPlugin{}
}

// Name returns the plugin name
func (p *PowerShellPlugin) Name() string {
	return "powershell"
}

// Execute executes a PowerShell script
func (p *PowerShellPlugin) Execute(ctx context.Context, config map[string]interface{}) (*probe.Result, error) {
	if runtime.GOOS != "windows" {
		return &probe.Result{
			ExitCode: 1,
			Stdout:   "",
			Stderr:   "PowerShell plugin only works on Windows",
			Success:  false,
		}, nil
	}
	
	script, ok := config["script"].(string)
	if !ok {
		return nil, fmt.Errorf("script is required")
	}
	
	// Execute PowerShell script
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}
	
	return &probe.Result{
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
		Success:  exitCode == 0,
	}, nil
}

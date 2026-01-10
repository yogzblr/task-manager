package probe

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// PowerShellTask executes PowerShell scripts on Windows
type PowerShellTask struct {
	Script  string
	Timeout time.Duration
}

// Configure sets up the PowerShell task
func (t *PowerShellTask) Configure(config map[string]interface{}) error {
	// Check if running on Windows
	if runtime.GOOS != "windows" {
		return fmt.Errorf("PowerShell task only works on Windows")
	}
	
	// Script is required
	script, ok := config["script"].(string)
	if !ok || script == "" {
		return fmt.Errorf("script is required")
	}
	t.Script = script
	
	// Timeout (default: 30s)
	if timeoutStr, ok := config["timeout"].(string); ok {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		t.Timeout = duration
	} else {
		t.Timeout = 30 * time.Second
	}
	
	return nil
}

// Execute runs the PowerShell script
func (t *PowerShellTask) Execute(ctx context.Context) (interface{}, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()
	
	// Execute PowerShell script
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", t.Script)
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("PowerShell execution failed: %w", err)
		}
	}
	
	result := map[string]interface{}{
		"output":    string(output),
		"exit_code": exitCode,
	}
	
	if exitCode != 0 {
		return result, fmt.Errorf("PowerShell script exited with code %d", exitCode)
	}
	
	return result, nil
}

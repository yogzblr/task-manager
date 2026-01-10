package probe

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// CommandTask executes local shell commands
type CommandTask struct {
	Command string
	Args    []string
	Timeout time.Duration
	Shell   bool
}

// Configure sets up the command task
func (t *CommandTask) Configure(config map[string]interface{}) error {
	// Command is required
	command, ok := config["command"].(string)
	if !ok || command == "" {
		return fmt.Errorf("command is required")
	}
	t.Command = command
	
	// Args (optional)
	if args, ok := config["args"].([]interface{}); ok {
		t.Args = make([]string, len(args))
		for i, arg := range args {
			if strArg, ok := arg.(string); ok {
				t.Args[i] = strArg
			}
		}
	}
	
	// Shell (default: false)
	if shell, ok := config["shell"].(bool); ok {
		t.Shell = shell
	}
	
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

// Execute runs the command
func (t *CommandTask) Execute(ctx context.Context) (interface{}, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()
	
	var cmd *exec.Cmd
	
	if t.Shell {
		// Execute through shell
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd.exe", "/c", t.Command)
		} else {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", t.Command)
		}
	} else {
		// Execute directly
		cmd = exec.CommandContext(ctx, t.Command, t.Args...)
	}
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}
	
	result := map[string]interface{}{
		"output":    string(output),
		"exit_code": exitCode,
	}
	
	if exitCode != 0 {
		return result, fmt.Errorf("command exited with code %d", exitCode)
	}
	
	return result, nil
}

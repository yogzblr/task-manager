package probe

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// Task represents an executable task
type Task interface {
	Execute(ctx context.Context) (*Result, error)
	Name() string
}

// Result represents task execution result
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Success  bool
}

// Runner executes tasks
type Runner struct {
	tasks []Task
}

// NewRunner creates a new task runner
func NewRunner() *Runner {
	return &Runner{
		tasks: make([]Task, 0),
	}
}

// AddTask adds a task to the runner
func (r *Runner) AddTask(task Task) {
	r.tasks = append(r.tasks, task)
}

// Run executes all tasks sequentially
func (r *Runner) Run(ctx context.Context) ([]*Result, error) {
	results := make([]*Result, 0, len(r.tasks))
	
	for _, task := range r.tasks {
		result, err := task.Execute(ctx)
		if err != nil {
			return results, fmt.Errorf("task %s failed: %w", task.Name(), err)
		}
		results = append(results, result)
		
		if !result.Success {
			return results, fmt.Errorf("task %s returned non-zero exit code: %d", task.Name(), result.ExitCode)
		}
	}
	
	return results, nil
}

// ExecuteCommand executes a shell command
func ExecuteCommand(ctx context.Context, command string, args ...string) (*Result, error) {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		// Windows: use cmd.exe
		cmd = exec.CommandContext(ctx, "cmd.exe", "/c", command)
	} else {
		// Unix: use /bin/sh
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
	}
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}
	
	// For simplicity, we'll treat all output as stdout
	// In production, you'd separate stdout and stderr
	return &Result{
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
		Success:  exitCode == 0,
	}, nil
}

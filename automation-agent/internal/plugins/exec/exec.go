package exec

import (
	"context"
	"fmt"

	"github.com/automation-platform/agent/internal/probe"
)

// ExecPlugin executes shell commands
type ExecPlugin struct{}

// New creates a new exec plugin
func New() *ExecPlugin {
	return &ExecPlugin{}
}

// Name returns the plugin name
func (p *ExecPlugin) Name() string {
	return "exec"
}

// Execute executes a shell command
func (p *ExecPlugin) Execute(ctx context.Context, config map[string]interface{}) (*probe.Result, error) {
	command, ok := config["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}
	
	return probe.ExecuteCommand(ctx, command)
}

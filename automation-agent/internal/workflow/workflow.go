package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/automation-platform/agent/internal/probe"
)

// Workflow represents a workflow definition
type Workflow struct {
	Steps []Step `yaml:"steps" json:"steps"`
}

// Step represents a workflow step
type Step struct {
	Type     string                 `yaml:"type" json:"type"`
	OS       string                 `yaml:"os,omitempty" json:"os,omitempty"`
	Config   map[string]interface{} `yaml:",inline" json:",inline"`
	Artifact *Artifact              `yaml:"artifact,omitempty" json:"artifact,omitempty"`
	Script   string                 `yaml:"script,omitempty" json:"script,omitempty"`
}

// Artifact represents an artifact to download
type Artifact struct {
	URL       string `yaml:"url" json:"url"`
	SHA256    string `yaml:"sha256" json:"sha256"`
	Signature string `yaml:"signature" json:"signature"`
	KeyID     string `yaml:"key_id" json:"key_id"`
}

// Executor executes workflows
type Executor struct {
	registry *probe.PluginRegistry
}

// NewExecutor creates a new workflow executor
func NewExecutor(registry *probe.PluginRegistry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// Execute executes a workflow
func (e *Executor) Execute(ctx context.Context, workflow *Workflow) ([]*probe.Result, error) {
	results := make([]*probe.Result, 0, len(workflow.Steps))
	
	for i, step := range workflow.Steps {
		// Check OS filter
		if step.OS != "" {
			// TODO: Check current OS and skip if doesn't match
		}
		
		// Get plugin
		plugin, err := e.registry.Get(step.Type)
		if err != nil {
			return results, fmt.Errorf("step %d: %w", i, err)
		}
		
		// Build config
		config := make(map[string]interface{})
		for k, v := range step.Config {
			config[k] = v
		}
		
		// Add step-specific fields
		if step.Artifact != nil {
			config["artifact"] = step.Artifact
		}
		if step.Script != "" {
			config["script"] = step.Script
		}
		
		// Execute step
		result, err := plugin.Execute(ctx, config)
		if err != nil {
			return results, fmt.Errorf("step %d failed: %w", i, err)
		}
		
		results = append(results, result)
		
		if !result.Success {
			return results, fmt.Errorf("step %d returned non-zero exit code: %d", i, result.ExitCode)
		}
	}
	
	return results, nil
}

// ParseWorkflow parses a workflow from JSON
func ParseWorkflow(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}
	return &workflow, nil
}

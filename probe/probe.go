package probe

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Probe is the main executor for workflows
type Probe struct {
	tasks map[string]TaskFactory
}

// TaskFactory creates a new task instance
type TaskFactory func() Task

// New creates a new Probe instance with all built-in tasks registered
func New() *Probe {
	p := &Probe{
		tasks: make(map[string]TaskFactory),
	}
	
	// Register built-in tasks
	p.RegisterTask("http", func() Task { return &HTTPTask{} })
	p.RegisterTask("db", func() Task { return &DBTask{} })
	p.RegisterTask("ssh", func() Task { return &SSHTask{} })
	p.RegisterTask("command", func() Task { return &CommandTask{} })
	p.RegisterTask("powershell", func() Task { return &PowerShellTask{} })
	p.RegisterTask("downloadexec", func() Task { return &DownloadExecTask{} })
	
	return p
}

// RegisterTask registers a new task type
func (p *Probe) RegisterTask(taskType string, factory TaskFactory) {
	p.tasks[taskType] = factory
}

// ExecuteYAML parses and executes a YAML workflow
func (p *Probe) ExecuteYAML(ctx context.Context, yamlData []byte) (*WorkflowResult, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(yamlData, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	return p.Execute(ctx, &workflow)
}

// Execute executes a workflow
func (p *Probe) Execute(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Name:    workflow.Name,
		Tasks:   make([]TaskResult, 0, len(workflow.Tasks)),
		Success: true,
	}
	
	for i, taskDef := range workflow.Tasks {
		// Get task factory
		factory, ok := p.tasks[taskDef.Type]
		if !ok {
			return result, fmt.Errorf("unknown task type: %s", taskDef.Type)
		}
		
		// Create task instance
		task := factory()
		
		// Configure task
		if err := task.Configure(taskDef.Config); err != nil {
			return result, fmt.Errorf("task %d (%s): failed to configure: %w", i, taskDef.Name, err)
		}
		
		// Execute task
		taskResult := TaskResult{
			Name: taskDef.Name,
			Type: taskDef.Type,
		}
		
		output, err := task.Execute(ctx)
		if err != nil {
			taskResult.Error = err.Error()
			taskResult.Success = false
			result.Success = false
			result.Tasks = append(result.Tasks, taskResult)
			return result, fmt.Errorf("task %d (%s): %w", i, taskDef.Name, err)
		}
		
		taskResult.Output = output
		taskResult.Success = true
		result.Tasks = append(result.Tasks, taskResult)
	}
	
	return result, nil
}

// Workflow represents a YAML workflow definition
type Workflow struct {
	Name    string           `yaml:"name"`
	Timeout string           `yaml:"timeout,omitempty"`
	Tasks   []TaskDefinition `yaml:"tasks"`
}

// TaskDefinition defines a task in the workflow
type TaskDefinition struct {
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// WorkflowResult contains the results of workflow execution
type WorkflowResult struct {
	Name    string
	Tasks   []TaskResult
	Success bool
}

// TaskResult contains the result of a single task
type TaskResult struct {
	Name    string
	Type    string
	Output  interface{}
	Success bool
	Error   string
}

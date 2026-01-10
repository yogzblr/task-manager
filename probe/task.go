package probe

import "context"

// Task is the interface that all tasks must implement
type Task interface {
	// Configure sets up the task with the given configuration
	Configure(config map[string]interface{}) error
	
	// Execute runs the task and returns the result
	Execute(ctx context.Context) (interface{}, error)
}

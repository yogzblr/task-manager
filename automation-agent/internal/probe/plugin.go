package probe

import (
	"context"
	"fmt"
)

// Plugin represents a task plugin
type Plugin interface {
	Name() string
	Execute(ctx context.Context, config map[string]interface{}) (*Result, error)
}

// PluginRegistry manages plugins
type PluginRegistry struct {
	plugins map[string]Plugin
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin
func (r *PluginRegistry) Register(plugin Plugin) {
	r.plugins[plugin.Name()] = plugin
}

// Get retrieves a plugin by name
func (r *PluginRegistry) Get(name string) (Plugin, error) {
	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return plugin, nil
}

// List returns all registered plugin names
func (r *PluginRegistry) List() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

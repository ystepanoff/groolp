package plugins

import (
	"fmt"
	"sync"

	"github.com/ystepanoff/groolp/core"
)

type PluginRegistry struct {
	plugins []Plugin
	mu      sync.Mutex
}

var Registry *PluginRegistry = &PluginRegistry{
	plugins: make([]Plugin, 0),
}

// RegisterPlugin() adds a new plugin to the registry
func (pr *PluginRegistry) RegisterPlugin(p Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	for _, plugin := range pr.plugins {
		if plugin.GetName() == p.GetName() {
			return fmt.Errorf(
				"plugin with name '%s' is already registered",
				p.GetName(),
			)
		}
	}

	pr.plugins = append(pr.plugins, p)
	return nil
}

// InitPlugins() initialises all registered plugins by registering their tasks
func (pr *PluginRegistry) InitPlugins(tm core.TaskManagerInterface) {
	for _, plugin := range pr.plugins {
		fmt.Printf(
			"Initializing plugin: %s v%s\n",
			plugin.GetName(),
			plugin.GetVersion(),
		)
		if err := plugin.RegisterTasks(tm); err != nil {
			fmt.Printf(
				"Error initializing plugin %s: %v\n",
				plugin.GetName(),
				err,
			)
			continue
		}
		fmt.Printf("Successfully initialised plugin: %s\n", plugin.GetName())
	}
}

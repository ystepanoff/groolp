package plugins

import "github.com/ystepanoff/groolp/core"

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// RegisterTasks() registers the plugin's tasks with TaskManager
	RegisterTasks(tm core.TaskManagerInterface) error

	GetName() string
	GetVersion() string
	GetDescription() string
}

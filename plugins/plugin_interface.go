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

/*
   Each plugin must provide an initialisation function that Groolp can invoke to register the plugin's tasks.
   This is typically done using Go's init() function which runs automatically when the module is imported.
*/

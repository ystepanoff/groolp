package plugins

import "github.com/ystepanoff/groolp/core"

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	RegisterTasks(tm core.TaskManagerInterface) error
}

const PluginSymbol = "Plugin"

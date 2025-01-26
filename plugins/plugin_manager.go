package plugins

import "github.com/ystepanoff/groolp/internal/core"

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	RegisterTasks(tm *core.TaskManager) error
}

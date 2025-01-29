package plugins

import (
	"sync"
)

var (
	pluginRegistry = make([]Plugin, 0)
	registryMutex  sync.Mutex
)

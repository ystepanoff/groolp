package plugins

import (
	"bytes"
	"fmt"
	"os/exec"
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

func InstallPlugin(pluginPath string) error {
	fmt.Printf("Fetching plugin (Go module): %s\n", pluginPath)

	cmdGet := exec.Command("go", "get", pluginPath+"@latest")
	var out bytes.Buffer
	cmdGet.Stdout = &out
	cmdGet.Stderr = &out

	if err := cmdGet.Run(); err != nil {
		return fmt.Errorf("go get failed: %v\nOutput: %s", err, out.String())
	}

	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Stdout = &out
	cmdTidy.Stderr = &out
	if err := cmdTidy.Run(); err != nil {
		return fmt.Errorf(
			"go mod tidy failed: %v\nOutput: %s",
			err,
			out.String(),
		)
	}

	return nil
}

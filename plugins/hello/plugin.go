package hello

import (
	"fmt"

	"github.com/ystepanoff/groolp/core"
	"github.com/ystepanoff/groolp/plugins"
)

type HelloPlugin struct{}

func (p *HelloPlugin) RegisterTasks(tm core.TaskManagerInterface) error {
	task := &core.Task{
		Name:        "hello",
		Description: "Prints Hello from HelloPlugin",
		Action: func() error {
			fmt.Println("Hello from HelloPlugin!")
			return nil
		},
	}
	return tm.Register(task)
}

func (p *HelloPlugin) GetName() string {
	return "HelloPlugin"
}

func (p *HelloPlugin) GetVersion() string {
	return "1.0.0"
}

func (p *HelloPlugin) GetDescription() string {
	return "A plugin that adds a sample task."
}

// init() registers the plugin with Groolp upon import
func init() {
	plugin := &HelloPlugin{}
	if err := plugins.Registry.RegisterPlugin(plugin); err != nil {
		fmt.Printf("Failed to register plugin %s: %v\n", plugin.GetName(), err)
	}
}

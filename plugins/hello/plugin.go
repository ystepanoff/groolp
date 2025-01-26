package hello

import (
	"fmt"

	"github.com/ystepanoff/groolp/internal/core"
	"github.com/ystepanoff/groolp/plugins"
)

type HelloPlugin struct{}

func NewHelloPlugin() plugins.Plugin {
	return &HelloPlugin{}
}

func (p *HelloPlugin) RegisterTasks(tm *core.TaskManager) error {
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

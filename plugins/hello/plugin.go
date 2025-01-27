package hello

import (
	"fmt"

	"github.com/ystepanoff/groolp/core"
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

var Plugin HelloPlugin

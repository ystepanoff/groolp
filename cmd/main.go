/*
Copyright © 2025 Yegor Stepanov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"github.com/ystepanoff/groolp/internal/cli"
	"github.com/ystepanoff/groolp/internal/core"
	"github.com/ystepanoff/groolp/internal/plugins/hello"
)

func main() {
	taskManager := core.NewTaskManager()

	// Register sample built-in tasks or plugins
	taskManager.Register(&core.Task{
		Name:        "clean",
		Description: "Clean the build directory",
		Action: func() error {
			println("Cleaning build directory...")
			// Implement cleaning logic here
			return nil
		},
	})

	taskManager.Register(&core.Task{
		Name:         "build",
		Description:  "Build the project",
		Dependencies: []string{"clean"},
		Action: func() error {
			println("Building the project...")
			// Implement build logic here
			return nil
		},
	})

	taskManager.Register(&core.Task{
		Name:         "deploy",
		Description:  "Deploy the project",
		Dependencies: []string{"build"},
		Action: func() error {
			println("Deploying the project...")
			// Implement deployment logic here
			return nil
		},
	})

	// Register plugins
	helloPlugin := hello.NewHelloPlugin()
	if err := helloPlugin.RegisterTasks(taskManager); err != nil {
		println("Error registering HelloPlugin:", err.Error())
	}

	rootCmd := cli.Initialize(taskManager)
	rootCmd.Execute()
}

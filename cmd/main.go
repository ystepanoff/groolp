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
	"fmt"

	"github.com/ystepanoff/groolp/internal/cli"
	"github.com/ystepanoff/groolp/internal/core"
	"github.com/ystepanoff/groolp/plugins/hello"
)

func main() {
	taskManager := core.NewTaskManager()
	rootCmd := cli.Init(taskManager)

	config, err := cli.InitConfig()
	if err != nil {
		fmt.Println("Error loading config file:", err)
		return
	}

	if err := taskManager.RegisterTasksFromConfig(config); err != nil {
		fmt.Println("Error registering tasks from config:", err)
	}

	// Register plugins
	helloPlugin := hello.NewHelloPlugin()
	if err := helloPlugin.RegisterTasks(taskManager); err != nil {
		fmt.Println("Error registering HelloPlugin:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Failed to run:", err)
	}
}

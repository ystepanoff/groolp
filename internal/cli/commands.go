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
package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/ystepanoff/groolp/internal/core"
	"github.com/ystepanoff/groolp/internal/watcher"
)

var taskManager *core.TaskManager

var (
	watchPaths []string
	watchTask  string
)

// Initialize() initializes the CLI with a TaskManager instance.
func Initialize(tm *core.TaskManager) *cobra.Command {
	taskManager = tm
	rootCmd := &cobra.Command{
		Use:   "groolp",
		Short: "groolp is a Gulp-like task runner built in Go (Groolp = Groovy Gulp)",
	}

	// Run Command
	runCmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run a specified task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskName := args[0]
			if err := taskManager.Run(taskName); err != nil {
				log.Printf("Error running task '%s': %v\n", taskName, err)
			}
		},
	}

	// List Command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasks := taskManager.ListTasks()
			log.Println("Available Tasks:")
			for _, task := range tasks {
				log.Printf("- %s: %s\n", task.Name, task.Description)
			}
		},
	}

	// Watch Command (Placeholder)
	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch files for changes and trigger tasks",
		Run: func(cmd *cobra.Command, args []string) {
			if watchTask == "" {
				fmt.Println("Specify a task to run on changes using --task")
				return
			}
			if len(watchPaths) == 0 {
				fmt.Println("Specify paths to watch using --path")
				return
			}

			w, err := watcher.NewWatcher(tm, watchPaths, watchTask)
			if err != nil {
				log.Printf("Error initializing watcher: %v\n", err)
				return
			}

			w.Start()
		},
	}

	watchCmd.Flags().StringSliceVarP(
		&watchPaths,
		"path", "p", []string{"."},
		"Paths to watch for changes",
	)
	watchCmd.Flags().StringVarP(
		&watchTask,
		"task", "t", "",
		"Task to run on changes",
	)

	rootCmd.AddCommand(runCmd, listCmd, watchCmd)
	return rootCmd
}

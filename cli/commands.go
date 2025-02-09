package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/ystepanoff/groolp/core"
	"github.com/ystepanoff/groolp/scripts"
	"github.com/ystepanoff/groolp/watcher"
)

var taskManager *core.TaskManager

var (
	watchPaths            []string
	watchTask             string
	watchDebounceDuration int64
)

// Init() initialises the CLI with a TaskManager instance.
func Init(tm *core.TaskManager, groolpDir string) *cobra.Command {
	taskManager = tm
	rootCmd := &cobra.Command{
		Use:   "groolp",
		Short: "Groolp is a Gulp-like task runner built in Go (Groolp = Groovy Gulp)",
	}

	// run command
	runCmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run a specified task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskName := args[0]
			if err := taskManager.Run(taskName); err != nil {
				rootCmd.Printf("Error running task '%s': %v\n", taskName, err)
			}
		},
	}

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasks := taskManager.ListTasks()
			rootCmd.Println("Available tasks:")
			for _, task := range tasks {
				rootCmd.Printf("- %s: %s\n", task.Name, task.Description)
			}
		},
	}

	// watch command
	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch files for changes and trigger tasks",
		Run: func(cmd *cobra.Command, args []string) {
			if watchTask == "" {
				rootCmd.Println("Specify a task to run on changes using --task")
				return
			}
			if len(watchPaths) == 0 {
				rootCmd.Println("Specify paths to watch using --path")
				return
			}

			w, err := watcher.NewWatcher(
				tm,
				watchPaths,
				watchTask,
				time.Duration(watchDebounceDuration)*time.Millisecond,
			)
			if err != nil {
				rootCmd.Printf("Error initializing watcher: %v\n", err)
				return
			}

			w.Start()
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if watchDebounceDuration < 500 {
				return fmt.Errorf(
					"invalid value for --debounce: %d; minimum allowed is 500 milliseconds",
					watchDebounceDuration,
				)
			}
			return nil
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
	watchCmd.Flags().Int64VarP(
		&watchDebounceDuration,
		"debounce", "d", 500,
		"Debounce duration in milliseconds (has to be at least 500)",
	)

	scriptCmd := &cobra.Command{
		Use:   "script",
		Short: "Manage Lua user scripts",
	}

	scriptInstallCmd := &cobra.Command{
		Use:   "install [url]",
		Short: "Install a Lua script",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			scriptsDir := filepath.Join(groolpDir, "scripts")
			if err := scripts.LuaInstaller.InstallScript(url, scriptsDir); err != nil {
				rootCmd.Printf("Error installing script: %v\n", err)
				return
			}
			rootCmd.Println("Script installed successfully!")
		},
	}
	scriptCmd.AddCommand(scriptInstallCmd)

	rootCmd.AddCommand(runCmd, listCmd, watchCmd, scriptCmd)
	return rootCmd
}

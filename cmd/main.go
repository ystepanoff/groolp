package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ystepanoff/groolp/cli"
	"github.com/ystepanoff/groolp/core"
	"github.com/ystepanoff/groolp/scripts"
)

const groolpDir = ".groolp"

func main() {
	if err := cli.InitGroolpDirectory(groolpDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing %s: %v\n", groolpDir, err)
		os.Exit(1)
	}

	taskManager := core.NewTaskManager()

	config, err := cli.InitConfig(groolpDir)
	if err != nil {
		fmt.Println("Error loading config file:", err)
		return
	}

	if err := taskManager.RegisterFromConfig(config); err != nil {
		fmt.Println("Error registering tasks from config:", err)
	}

	scriptsDir := filepath.Join(groolpDir, "scripts")
	if err := scripts.LoadScripts(scriptsDir, taskManager); err != nil {
		fmt.Printf("Error loading scripts at startup: %v\n", err)
	}

	rootCmd := cli.Init(taskManager)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Failed to run:", err)
	}
}

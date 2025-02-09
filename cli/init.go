package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ystepanoff/groolp/core"
)

// InitGroolpDirectory() initialises ".groolp" directory if does not exist
func InitGroolpDirectory(groolpDir string) error {
	tasksConfig := filepath.Join(groolpDir, "tasks.yaml")
	scriptsDir := filepath.Join(groolpDir, "scripts")
	sampleScript := filepath.Join(scriptsDir, "hello.lua")

	fi, err := os.Stat(groolpDir)
	if err == nil {
		if fi.IsDir() {
			return nil
		} else {
			return fmt.Errorf("found .groolp file instead of a directory")
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat .groolp: %w", err)
		}
	}

	fmt.Println("Initializing .groolp directory...")
	if err := os.Mkdir(groolpDir, 0755); err != nil {
		return fmt.Errorf("failed to create .groolp dir: %w", err)
	}

	sampleTasks := `# Sample tasks.yaml
tasks:
  # This is a sample YAML-based task definition
  sample-yaml-task:
    description: "A sample task from tasks.yaml"
    action: "echo Hello from tasks.yaml!"
`
	if err := os.WriteFile(tasksConfig, []byte(sampleTasks), 0644); err != nil {
		return fmt.Errorf("failed to write tasks.yaml: %w", err)
	}

	if err := os.Mkdir(scriptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .groolp/scripts dir: %w", err)
	}

	sampleLua := `-- sample.lua
-- Register a sample plugin-based task in Lua

register_task(
  "sample-lua-task",
  "Print a greeting from sample.lua",
  function()
    print("Hello from sample.lua!")
  end
)
`
	if err := os.WriteFile(sampleScript, []byte(sampleLua), 0644); err != nil {
		return fmt.Errorf("failed to write sample.lua: %w", err)
	}

	fmt.Println(
		"Created .groolp directory with sample tasks.yaml and scripts/sample.lua",
	)
	return nil
}

// InitTasksConfig() loading simple tasks from tasks config
func InitTasksConfig(groolpDir string) (*core.TasksConfig, error) {
	tasksConfig := filepath.Join(groolpDir, "tasks.yaml")
	config, err := core.LoadConfig(tasksConfig)
	if err != nil {
		return nil, fmt.Errorf("error loading tasks config: %v", err)
	}
	return config, nil
}

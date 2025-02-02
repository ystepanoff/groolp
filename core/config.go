package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// TasksConfig represents the structure of the tasks configuration file.
type TasksConfig struct {
	Tasks map[string]struct {
		Description  string   `yaml:"description"`
		Dependencies []string `yaml:"dependencies,omitempty"`
		Action       string   `yaml:"action"`
	} `yaml:"tasks"`
}

// LoadConfig loads and parses the configuration file.
func LoadConfig(filename string) (*TasksConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config TasksConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// RegisterTasksFromConfig registers tasks defined in the configuration file.
func (tm *TaskManager) RegisterFromConfig(config *TasksConfig) error {
	for name, taskData := range config.Tasks {
		task := NewTaskFromConfig(
			name,
			taskData.Description,
			taskData.Dependencies,
			taskData.Action,
		)
		if err := tm.Register(task); err != nil {
			return fmt.Errorf("failed to register task '%s': %w", name, err)
		}
	}
	return nil
}

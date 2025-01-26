package core

import (
	"fmt"
	"log"
	"sync"
)

// Task represents a single task with its dependencies and action
type Task struct {
	Name         string
	Description  string
	Dependencies []string
	Action       func() error
}

// TaskManager manages registration and execution of tasks
type TaskManager struct {
	tasks map[string]*Task
	mu    sync.Mutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

// Register() adds a new task to the manager
func (tm *TaskManager) Register(task *Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[task.Name]; exists {
		return fmt.Errorf("task '%s' already exists", task.Name)
	}
	tm.tasks[task.Name] = task

	return nil
}

// Run() executes tasks and its dependencies
func (tm *TaskManager) Run(taskName string) error {
	return tm.run(taskName, nil)
}

func (tm *TaskManager) run(taskName string, visited map[string]bool) error {
	tm.mu.Lock()
	task, exists := tm.tasks[taskName]
	tm.mu.Unlock()

	if !exists {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	if visited[taskName] {
		return fmt.Errorf("circular dependency detected on task '%s'", taskName)
	}
	visited[taskName] = true

	// Make sure dependencies run first
	for _, dep := range task.Dependencies {
		if err := tm.run(dep, visited); err != nil {
			return err
		}
	}

	// Execute the task
	log.Printf("Running task: %s\n", task.Name)
	return task.Action()
}

func (tm *TaskManager) ListTasks() []*Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	taskList := []*Task{}
	for _, task := range tm.tasks {
		taskList = append(taskList, task)
	}

	return taskList
}

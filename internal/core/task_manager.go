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
	task, err := tm.retrieveAndCheck(taskName, nil)
	if err != nil {
		return err
	}

	// Make sure dependencies run first
	for _, dep := range task.Dependencies {
		if err := tm.Run(dep); err != nil {
			return err
		}
	}

	// Execute the task
	log.Printf("Running task: %s\n", task.Name)
	return task.Action()
}

func (tm *TaskManager) retrieveAndCheck(
	taskName string,
	visited map[string]bool,
) (*Task, error) {
	tm.mu.Lock()
	task, exists := tm.tasks[taskName]
	tm.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("task '%s' not found", taskName)
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	if visited[taskName] {
		return nil, fmt.Errorf(
			"circular dependency detected on task '%s'",
			taskName,
		)
	}
	visited[taskName] = true

	for _, dep := range task.Dependencies {
		if _, err := tm.retrieveAndCheck(dep, visited); err != nil {
			return nil, err
		}
	}

	return task, nil
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

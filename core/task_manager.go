package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

// Task represents a single task with its dependencies and action
type Task struct {
	Name         string
	Description  string
	Dependencies []string
	Action       func() error
}

func NewTaskFromConfig(
	name string,
	description string,
	dependencies []string,
	actionCmd string,
) *Task {
	return &Task{
		Name:         name,
		Description:  description,
		Dependencies: dependencies,
		Action: func() error {
			cmd := exec.Command("sh", "-c", actionCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

// TaskManagerInterface defines the methods that TaskManager exposes
type TaskManagerInterface interface {
	Register(task *Task) error
	Run(taskName string) error
	ListTasks() []*Task
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
	recStack map[string]bool,
) (*Task, error) {
	tm.mu.Lock()
	task, exists := tm.tasks[taskName]
	tm.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("task '%s' not found", taskName)
	}

	if recStack == nil {
		recStack = make(map[string]bool)
	}

	if recStack[taskName] {
		return nil, fmt.Errorf(
			"circular dependency detected on task '%s'",
			taskName,
		)
	}

	recStack[taskName] = true

	for _, dep := range task.Dependencies {
		if _, err := tm.retrieveAndCheck(dep, recStack); err != nil {
			return nil, err
		}
	}

	recStack[taskName] = false

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

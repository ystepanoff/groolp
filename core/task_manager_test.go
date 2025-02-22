package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskRegistration(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		Name:        "test-task",
		Description: "A test task",
		Action: func() error {
			return nil
		},
	}

	if err := tm.Register(task); err != nil {
		t.Errorf("Failed tp register task: %v", err)
	}

	// Attempt to register the same task again
	if err := tm.Register(task); err == nil {
		t.Errorf("Expected error when registering duplicate task, got nil")
	}
}

func TestTaskExecution(t *testing.T) {
	tm := NewTaskManager()
	executed := false

	task := &Task{
		Name:        "execute-task",
		Description: "Executes a task",
		Action: func() error {
			executed = true
			return nil
		},
	}

	_ = tm.Register(task)

	if err := tm.Run("execute-task"); err != nil {
		t.Errorf("Failed to run task: %v", err)
	}

	if !executed {
		t.Errorf("Task action was not executed")
	}
}

func TestTaskDependencies(t *testing.T) {
	tm := NewTaskManager()
	executionOrder := []string{}

	taskA := &Task{
		Name:        "taskA",
		Description: "Task A",
		Action: func() error {
			executionOrder = append(executionOrder, "taskA")
			return nil
		},
	}

	taskB := &Task{
		Name:         "taskB",
		Description:  "Task B",
		Dependencies: []string{"taskA"},
		Action: func() error {
			executionOrder = append(executionOrder, "taskB")
			return nil
		},
	}

	_ = tm.Register(taskA)
	_ = tm.Register(taskB)

	if err := tm.Run("taskB"); err != nil {
		t.Errorf("Failed to run task with dependencies: %v", err)
	}

	expectedOrder := []string{"taskA", "taskB"}
	for i, taskName := range expectedOrder {
		if executionOrder[i] != taskName {
			t.Errorf(
				"Expected execution order %v, got %v",
				expectedOrder,
				executionOrder,
			)
			break
		}
	}
}

func TestRetrieveAndCheck(t *testing.T) {
	tm := NewTaskManager()

	taskA := &Task{
		Name:         "taskA",
		Description:  "Task A",
		Dependencies: []string{"taskB", "taskC"},
		Action:       func() error { return nil },
	}
	taskB := &Task{
		Name:         "taskB",
		Description:  "Task B",
		Dependencies: []string{"taskC"},
		Action:       func() error { return nil },
	}
	taskC := &Task{
		Name:         "taskC",
		Description:  "Task C",
		Dependencies: nil,
		Action:       func() error { return nil },
	}
	require.NoError(t, tm.Register(taskA))
	require.NoError(t, tm.Register(taskB))
	require.NoError(t, tm.Register(taskC))

	task, err := tm.retrieveAndCheck("taskA", make(map[string]bool))
	require.NoError(t, err)
	require.Equal(t, "taskA", task.Name)

	taskX := &Task{
		Name:         "taskX",
		Description:  "Task X",
		Dependencies: []string{"taskY"},
		Action:       func() error { return nil },
	}
	taskY := &Task{
		Name:         "taskY",
		Description:  "Task Y",
		Dependencies: []string{"taskZ"},
		Action:       func() error { return nil },
	}
	taskZ := &Task{
		Name:         "taskZ",
		Description:  "Task Z",
		Dependencies: []string{"taskX"},
		Action:       func() error { return nil },
	}
	require.NoError(t, tm.Register(taskX))
	require.NoError(t, tm.Register(taskY))
	require.NoError(t, tm.Register(taskZ))

	_, err = tm.retrieveAndCheck("taskX", make(map[string]bool))
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency detected")
}

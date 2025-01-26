package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ystepanoff/groolp/internal/core"
)

func TestRunCommand(t *testing.T) {
	tm := core.NewTaskManager()

	executed := false
	_ = tm.Register(&core.Task{
		Name:        "test-task",
		Description: "A test task",
		Action: func() error {
			executed = true
			return nil
		},
	})

	rootCmd := Init(tm)
	rootCmd.SetArgs([]string{"run", "test-task"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !executed {
		t.Errorf("Task was not executed")
	}
}

var listTests = [][]string{
	{"taskA"},
	{"taskA", "taskB"},
	{"taskA", "taskB", "taskC"},
	{"taskA", "taskC"},
	{"taskA", "taskA"},
	{"☄⨫➰ⱜ⦟⵬⦃⚗⤼☣", "☄⨫➰ⱜ⦟⵬⦃⚗⤼☣"},
	{"ℰ⤍⎿⍷⃥✦⥈⭎⨑⼵", "♔⫫", "ⲚⅪ⏛♈"},
}

func TestListCommand(t *testing.T) {
	for _, test := range listTests {
		tm := core.NewTaskManager()

		for _, task := range test {
			_ = tm.Register(&core.Task{
				Name:        task,
				Description: task,
			})
		}

		buf := new(bytes.Buffer)
		rootCmd := Init(tm)
		rootCmd.SetOut(buf)
		rootCmd.SetArgs([]string{"list"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedOutput := "Available tasks:\n"
		expectedOutputLines := make([]string, 0)
		seen := make(map[string]bool)
		for _, task := range test {
			if _, exists := seen[task]; !exists {
				expectedOutputLines = append(
					expectedOutputLines,
					fmt.Sprintf("- %s: %s", task, task),
				)
				seen[task] = true
			}
		}
		sort.Strings(expectedOutputLines)
		expectedOutput += strings.Join(expectedOutputLines, "\n")

		actualOutputLines := strings.Split(buf.String(), "\n")
		actualOutput := actualOutputLines[0]
		actualOutputLines = actualOutputLines[1:]
		sort.Strings(actualOutputLines)
		actualOutput += strings.Join(actualOutputLines, "\n")

		if actualOutput != expectedOutput {
			t.Errorf(
				"Expected output '%s', got '%s'",
				expectedOutput,
				buf.String(),
			)
		}
	}
}

func TestWatchCommand(t *testing.T) {
	tm := core.NewTaskManager()

	buf := new(bytes.Buffer)
	rootCmd := Init(tm)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"watch"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedOutput := "Specify a task to run on changes using --task\n"
	if buf.String() != expectedOutput {
		t.Errorf(
			"Expected output '%s', got '%s'",
			expectedOutput,
			buf.String(),
		)
	}
}

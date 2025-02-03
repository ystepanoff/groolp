package cli

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ystepanoff/groolp/core"
	"github.com/ystepanoff/groolp/scripts"
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

type MockInstaller struct {
	errToReturn error
}

func (m *MockInstaller) InstallScript(url, scriptsDir string) error {
	return m.errToReturn
}

func TestRunCommand_UnknownTask(t *testing.T) {
	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	rootCmd.SetArgs([]string{"run", "nonexistent-task"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error running task 'nonexistent-task'") {
		t.Errorf("Expected error about unknown task, got: %s", output)
	}
}

func TestRunCommand_TaskFailure(t *testing.T) {
	tm := core.NewTaskManager()

	_ = tm.Register(&core.Task{
		Name:        "fail-task",
		Description: "Always fails",
		Action: func() error {
			return errors.New("simulated failure")
		},
	})

	rootCmd := Init(tm)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"run", "fail-task"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "simulated failure") {
		t.Errorf("Expected failure message in output, got: %s", output)
	}
}

func TestWatchCommand_NoPathSpecified(t *testing.T) {
	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"watch", "--task", "some-task", "--path", ""})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Specify paths to watch using --path\n"
	if buf.String() != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, buf.String())
	}
}

func TestWatchCommand_InvalidDebounce(t *testing.T) {
	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{
		"watch",
		"--path", ".",
		"--task", "some-task",
		"--debounce", "400",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("Expected an error for debounce < 500, but got nil")
	}

	if !strings.Contains(err.Error(), "invalid value for --debounce: 400") {
		t.Errorf("Expected invalid debounce error, got: %v", err)
	}
}

func TestWatchCommand_Success(t *testing.T) {
	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{
		"watch",
		"--path", "some/dir",
		"--task", "some-task",
		"--debounce", "500",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	/*output := buf.String()
	if len(output) == 0 {
		// Probably fine; the watcher started with no messages.
	}*/
}

func TestScriptInstallCommand_Success(t *testing.T) {
	origInstaller := scripts.LuaInstaller
	defer func() { scripts.LuaInstaller = origInstaller }()

	scripts.LuaInstaller = &MockInstaller{}

	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs(
		[]string{"script", "install", "https://example.com/test.lua"},
	)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	expectedMsg := "Script installed successfully!\n"
	if output != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, output)
	}
}

func TestScriptInstallCommand_Error(t *testing.T) {
	origInstaller := scripts.LuaInstaller
	defer func() { scripts.LuaInstaller = origInstaller }()

	scripts.LuaInstaller = &MockInstaller{
		errToReturn: fmt.Errorf("mock install error"),
	}

	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs(
		[]string{"script", "install", "https://example.com/broken.lua"},
	)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(
		output,
		"Error installing script: mock install error",
	) {
		t.Errorf(
			"Expected error message about 'mock install error', got: %s",
			output,
		)
	}
}

func TestScriptInstallCommand_NonLua(t *testing.T) {
	origInstaller := scripts.LuaInstaller
	defer func() { scripts.LuaInstaller = origInstaller }()

	scripts.LuaInstaller = &MockInstaller{}

	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs(
		[]string{"script", "install", "https://example.com/invalid.txt"},
	)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(
		output,
		"Error installing script: refusing to install non-.lua file",
	) {
		t.Errorf("Expected a refusal to install .txt message, got: %s", output)
	}
}

func TestWatchCommand_DebounceBoundary(t *testing.T) {
	tm := core.NewTaskManager()
	rootCmd := Init(tm)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{
		"watch",
		"--path", "some/dir",
		"--task", "some-task",
		"--debounce", "500",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error at the boundary of 500ms, got: %v", err)
	}
}

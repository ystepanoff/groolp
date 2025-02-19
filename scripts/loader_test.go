package scripts

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ystepanoff/groolp/core"
)

func getTask(tm *core.TaskManager, name string) *core.Task {
	for _, task := range tm.ListTasks() {
		if task.Name == name {
			return task
		}
	}
	return nil
}

func TestLoadScripts_Success(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	luaScript := `register_task("test-task", "A test Lua task", function()
  -- do something trivial
  print("Hello from test-task!")
end)
`
	scriptPath := filepath.Join(tmpDir, "test.lua")
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaScript), 0644))

	tm := core.NewTaskManager()

	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err, "LoadScripts should succeed for a valid script")

	require.Len(t, scriptEngines, 1, "expected one scriptEngine to be created")

	engine := scriptEngines[0]
	require.NotNil(t, engine.L, "the lua.LState should be initialized")
	require.Len(
		t,
		engine.tasks,
		1,
		"expected exactly one task in this scriptEngine",
	)
	require.Equal(t, "test-task", engine.tasks[0].Name)

	task := getTask(tm, "test-task")
	require.NotNil(
		t,
		task,
		"the TaskManager should have a task named 'test-task'",
	)
	require.Equal(t, "A test Lua task", task.Description)
}

func TestLoadScripts_MultipleLuaFiles(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	script1 := `register_task("task1", "First Script Task", function() end)`
	script2 := `register_task("task2", "Second Script Task", function() end)`

	require.NoError(
		t,
		os.WriteFile(filepath.Join(tmpDir, "one.lua"), []byte(script1), 0644),
	)
	require.NoError(
		t,
		os.WriteFile(filepath.Join(tmpDir, "two.lua"), []byte(script2), 0644),
	)

	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)

	require.Len(
		t,
		scriptEngines,
		2,
		"two separate scriptEngines should be created",
	)

	require.Len(t, scriptEngines[0].tasks, 1)
	require.Len(t, scriptEngines[1].tasks, 1)

	require.NotNil(
		t,
		getTask(tm, "task1"),
		"expected 'task1' to be registered in TaskManager",
	)
	require.NotNil(
		t,
		getTask(tm, "task2"),
		"expected 'task2' to be registered in TaskManager",
	)
}

func TestLoadScripts_NonExistingDir(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()
	bogusDir := filepath.Join(tmpDir, "doesnotexist")

	tm := core.NewTaskManager()
	err := LoadScripts(bogusDir, tm)
	require.Error(t, err, "expected an error if directory does not exist")

	require.Nil(
		t,
		scriptEngines,
		"scriptEngines should remain nil (no engines added)",
	)
}

func TestLoadScripts_InvalidLuaScript(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	validScript := `register_task("valid-task", "Valid script", function() end)`
	invalidScript := `register_task("invalid-task", "Invalid script", function(` // syntax error

	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "valid.lua"),
			[]byte(validScript),
			0644,
		),
	)
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "invalid.lua"),
			[]byte(invalidScript),
			0644,
		),
	)

	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(
		t,
		err,
		"LoadScripts doesn't bubble up the error globally, logs instead",
	)

	for _, engine := range scriptEngines {
		switch engine.Name {
		case "invalid.lua":
			require.Len(
				t,
				scriptEngines[0].tasks,
				0,
				"no tasks from the invalid script",
			)
		case "valid.lua":
			require.Len(
				t,
				scriptEngines[1].tasks,
				1,
				"only one task from the valid script",
			)
		}
	}

	require.Nil(
		t,
		getTask(tm, "invalid-task"),
		"the invalid script's task should not be registered",
	)
	require.NotNil(
		t,
		getTask(tm, "valid-task"),
		"the valid script's task should be registered",
	)
}

func TestLoadScripts_SkipNonLuaFiles(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	luaScript := `register_task("test-task", "Description", function() end)`
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "script.lua"),
			[]byte(luaScript),
			0644,
		),
	)

	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "not-lua.txt"),
			[]byte("txt data"),
			0644,
		),
	)

	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))

	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)

	require.Len(t, scriptEngines, 1)

	require.NotNil(t, getTask(tm, "test-task"))
	require.Nil(t, getTask(tm, "subdir"))
	require.Nil(t, getTask(tm, "not-lua"))
}

func TestLoadScripts_TaskInvocation(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	luaScript := `
register_task("invoke-task", "Invoke test", function()
  -- Here we do something trivial; no error
  local x = 1 + 1
  print("x is ", x)
end)
`
	scriptPath := filepath.Join(tmpDir, "invoke.lua")
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaScript), 0644))

	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)

	task := getTask(tm, "invoke-task")
	require.NotNil(t, task, "the task should be registered")

	require.NoError(
		t,
		task.Action(),
		"invoking the Lua function should not error",
	)
}

func TestLoadScripts_DisabledLuaFunctions(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	disabledFuncScript := `
register_task("disabled-func-task", "Should fail", function()
  dofile("/Users/estepanov/some_other.lua") -- dofile is disabled
end)
`
	scriptFile := filepath.Join(tmpDir, "disabled.lua")
	require.NoError(
		t,
		os.WriteFile(scriptFile, []byte(disabledFuncScript), 0644),
	)

	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(
		t,
		err,
		"LoadScripts logs the error but does not bubble it up",
	)

	err = tm.Run("disabled-func-task")
	require.Error(t, err, "Didn't raise Lua runtime error")
	require.Contains(t, err.Error(), "lua runtime error")
}

func TestLoadScripts_EngineState(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	luaScript := `
register_task("test-task", "Engine state check", function()
  -- trivial
end)
`
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "script.lua"),
			[]byte(luaScript),
			0644,
		),
	)

	tm := core.NewTaskManager()
	require.NoError(t, LoadScripts(tmpDir, tm))
	require.Len(t, scriptEngines, 1, "one engine expected")

	engine := scriptEngines[0]
	require.NotNil(t, engine.L, "lua.LState should be initialized")

	require.NotPanics(t, func() {
		_ = engine.L.DoString(`local t = 1 + 2`)
	}, "expected the Lua state to still be open and usable")
}

func TestLoadScripts_MultipleTasksInOneScript(t *testing.T) {
	scriptEngines = nil

	tmpDir := t.TempDir()

	luaScript := `
register_task("task-one", "First", function() end)
register_task("task-two", "Second", function() end)
`
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "multi.lua"),
			[]byte(luaScript),
			0644,
		),
	)

	tm := core.NewTaskManager()
	require.NoError(t, LoadScripts(tmpDir, tm))

	require.Len(
		t,
		scriptEngines,
		1,
		"only one scriptEngine for a single .lua file",
	)
	require.Len(
		t,
		scriptEngines[0].tasks,
		2,
		"the single engine should have two tasks",
	)

	taskOne := getTask(tm, "task-one")
	taskTwo := getTask(tm, "task-two")
	require.NotNil(t, taskOne, "task-one should be in TaskManager")
	require.NotNil(t, taskTwo, "task-two should be in TaskManager")
	require.Equal(t, "First", taskOne.Description)
	require.Equal(t, "Second", taskTwo.Description)
}

func TestLoadScripts_TasksWithDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	scriptA := `
register_task(
  "clean",
  "Clean up build directory",
  function()
    print("Cleaning...")
  end
)

register_task(
  "build",
  "Build the project",
  function()
    print("Building after cleaning...")
  end,
  { "clean" }  -- depends on "clean"
)

register_task(
  "deploy",
  "Deploy the project",
  function()
    print("Deploying after build...")
  end,
  { "build" }  -- depends on "build"
)
`
	scriptAPath := filepath.Join(tmpDir, "scriptA.lua")
	err := os.WriteFile(scriptAPath, []byte(scriptA), 0644)
	require.NoError(t, err)

	tm := core.NewTaskManager()

	err = LoadScripts(
		tmpDir,
		tm,
	)
	require.NoError(t, err)

	cleanTask := getTask(tm, "clean")
	require.Empty(t, cleanTask.Dependencies, "clean has no deps")

	buildTask := getTask(tm, "build")
	require.Equal(
		t,
		[]string{"clean"},
		buildTask.Dependencies,
		"build depends on clean",
	)

	deployTask := getTask(tm, "deploy")
	require.Equal(
		t,
		[]string{"build"},
		deployTask.Dependencies,
		"deploy depends on build",
	)
}

func TestLoadScript_NoDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "nodeps.lua")
	luaContent := `
		register_task(
			"task_no_deps",
			"Task with no dependencies",
			function()
				print("Running task with no dependencies")
			end
		)
	`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))

	tm := core.NewTaskManager()

	err := loadScript(scriptPath, "nodeps", tm)
	require.NoError(t, err)

	task := getTask(tm, "task_no_deps")
	require.NotNil(t, task, "Expected 'task_no_deps' to be registered")

	require.Empty(
		t,
		task.Dependencies,
		"Expected no dependencies for 'task_no_deps'",
	)
}

func TestLoadScript_RunCommand(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "echo.lua")
	luaContent := `
register_task(
	"echo-task",
	"Task that runs an echo command",
	function()
		local code, err = run_command("echo hello")
		if err then
			error("run_command error: " .. err)
		else
			print("Echo command returned", code)
		end
	end
)
`
	err := os.WriteFile(scriptPath, []byte(luaContent), 0644)
	require.NoError(t, err)
	tm := core.NewTaskManager()
	err = loadScript(scriptPath, "echo", tm)
	require.NoError(t, err, "expected script to load without error")

	task := getTask(tm, "echo-task")
	require.NotNil(t, task, "expected 'echo-task' to be registered")

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = task.Action()
	require.NoError(t, err, "expected task to run without error")

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	require.Contains(t, output, "hello", "expected output to contain 'hello'")
	require.Contains(
		t,
		output,
		"Echo command returned",
		"expected output to contain the echo command message",
	)
}

func TestLoadScript_DataBridging(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_test.lua")
	luaContent := `
		-- This script sets a value in the global data store.
		set_data("myKey", "myValue")
	`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))

	tm := core.NewTaskManager()

	GlobalDataStore = NewDataStore(tmpDir)

	err := loadScript(scriptPath, "data_test", tm)
	require.NoError(t, err)

	val, ok := GlobalDataStore.GetData("myKey")
	require.True(t, ok, "Expected key 'myKey' to be set")
	require.Equal(t, "myValue", val)
}

package scripts

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
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
  print("Hello from test-task!")
end)`
	scriptPath := filepath.Join(tmpDir, "test.lua")
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaScript), 0644))
	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)
	require.Len(t, scriptEngines, 1)
	engine := scriptEngines[0]
	require.NotNil(t, engine.L)
	require.Len(t, engine.tasks, 1)
	require.Equal(t, "test-task", engine.tasks[0].Name)
	task := getTask(tm, "test-task")
	require.NotNil(t, task)
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
	require.Len(t, scriptEngines, 2)
	require.Len(t, scriptEngines[0].tasks, 1)
	require.Len(t, scriptEngines[1].tasks, 1)
	require.NotNil(t, getTask(tm, "task1"))
	require.NotNil(t, getTask(tm, "task2"))
}

func TestLoadScripts_NonExistingDir(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	bogusDir := filepath.Join(tmpDir, "doesnotexist")
	tm := core.NewTaskManager()
	err := LoadScripts(bogusDir, tm)
	require.Error(t, err)
	require.Nil(t, scriptEngines)
}

func TestLoadScripts_InvalidLuaScript(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	validScript := `register_task("valid-task", "Valid script", function() end)`
	invalidScript := `register_task("invalid-task", "Invalid script", function(`
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
	require.NoError(t, err)
	for _, engine := range scriptEngines {
		switch engine.Name {
		case "invalid.lua":
			require.Len(t, engine.tasks, 0)
		case "valid.lua":
			require.Len(t, engine.tasks, 1)
		}
	}
	require.Nil(t, getTask(tm, "invalid-task"))
	require.NotNil(t, getTask(tm, "valid-task"))
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
	require.NotNil(t, task)
	require.NoError(t, task.Action())
}

func TestLoadScripts_DisabledLuaFunctions(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	disabledFuncScript := `
register_task("disabled-func-task", "Should fail", function()
  dofile("/Users/estepanov/some_other.lua")
end)
`
	scriptFile := filepath.Join(tmpDir, "disabled.lua")
	require.NoError(
		t,
		os.WriteFile(scriptFile, []byte(disabledFuncScript), 0644),
	)
	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)
	err = tm.Run("disabled-func-task")
	require.Error(t, err)
	require.Contains(t, err.Error(), "lua runtime error")
}

func TestLoadScripts_EngineState(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	luaScript := `register_task("test-task", "Engine state check", function() end)`
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
	require.Len(t, scriptEngines, 1)
	engine := scriptEngines[0]
	require.NotNil(t, engine.L)
	require.NotPanics(t, func() {
		_ = engine.L.DoString(`local t = 1 + 2`)
	})
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
	require.Len(t, scriptEngines, 1)
	require.Len(t, scriptEngines[0].tasks, 2)
	taskOne := getTask(tm, "task-one")
	taskTwo := getTask(tm, "task-two")
	require.NotNil(t, taskOne)
	require.NotNil(t, taskTwo)
	require.Equal(t, "First", taskOne.Description)
	require.Equal(t, "Second", taskTwo.Description)
}

func TestLoadScripts_TasksWithDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	scriptA := `
register_task("clean","Clean up build directory",function() print("Cleaning...") end)
register_task("build","Build the project",function() print("Building...") end, {"clean"})
register_task("deploy","Deploy the project",function() print("Deploying...") end, {"build"})
`
	scriptAPath := filepath.Join(tmpDir, "scriptA.lua")
	err := os.WriteFile(scriptAPath, []byte(scriptA), 0644)
	require.NoError(t, err)
	tm := core.NewTaskManager()
	err = LoadScripts(tmpDir, tm)
	require.NoError(t, err)
	cleanTask := getTask(tm, "clean")
	buildTask := getTask(tm, "build")
	deployTask := getTask(tm, "deploy")
	require.NotNil(t, cleanTask)
	require.NotNil(t, buildTask)
	require.NotNil(t, deployTask)
	require.Empty(t, cleanTask.Dependencies)
	require.Equal(t, []string{"clean"}, buildTask.Dependencies)
	require.Equal(t, []string{"build"}, deployTask.Dependencies)
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
	require.NotNil(t, task)
	require.Empty(t, task.Dependencies)
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
		end
		print("Echo command returned", code)
	end
)
`
	err := os.WriteFile(scriptPath, []byte(luaContent), 0644)
	require.NoError(t, err)
	tm := core.NewTaskManager()
	err = loadScript(scriptPath, "echo", tm)
	require.NoError(t, err)
	task := getTask(tm, "echo-task")
	require.NotNil(t, task)
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	err = task.Action()
	require.NoError(t, err)
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()
	require.Contains(t, output, "hello")
	require.Contains(t, output, "Echo command returned")
}

func TestLoadScript_DataBridging(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_test.lua")
	luaContent := `set_data("myKey", "myValue")`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	err = loadScript(scriptPath, "data_test", tm)
	require.NoError(t, err)
	val, ok := GlobalDataStore.GetData("myKey")
	require.True(t, ok)
	require.Equal(t, "myValue", val)
	ds.Close()
}

func TestLoadScript_DataBridgingNumber(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_number.lua")
	luaContent := `set_data("numKey", 123)`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	err = loadScript(scriptPath, "data_number", tm)
	require.NoError(t, err)
	val, ok := GlobalDataStore.GetData("numKey")
	require.True(t, ok)
	require.Equal(t, 123.0, val)
	ds.Close()
}

func TestLoadScript_DataBridgingBool(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_bool.lua")
	luaContent := `set_data("boolKey", true)`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	err = loadScript(scriptPath, "data_bool", tm)
	require.NoError(t, err)
	val, ok := GlobalDataStore.GetData("boolKey")
	require.True(t, ok)
	require.Equal(t, true, val)
	ds.Close()
}

func TestLoadScript_GetDataNonExistentKey(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_none.lua")
	luaContent := `
register_task("checkKey", "Check missing key", function()
  local v = get_data("missingKey")
  if v ~= nil then
    error("expected nil for missingKey")
  end
end)
`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	err = loadScript(scriptPath, "data_none", tm)
	require.NoError(t, err)
	task := getTask(tm, "checkKey")
	require.NotNil(t, task)
	err = task.Action()
	require.NoError(t, err)
	ds.Close()
}

func TestLoadScript_RunCommandInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "invalid_cmd.lua")
	luaContent := `
register_task("invalid-cmd-task", "Runs invalid cmd", function()
	local code, err = run_command("nonexistent_command_123")
	if err == nil then
		error("expected an error for invalid command")
	end
	print("Command code:", code, "Command err:", err)
end)
`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	err := loadScript(scriptPath, "invalid_cmd", tm)
	require.NoError(t, err)
	task := getTask(tm, "invalid-cmd-task")
	require.NotNil(t, task)
	err = task.Action()
	require.NoError(t, err)
}

func TestLoadScript_ConcurrentDataAccess(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPathA := filepath.Join(tmpDir, "scriptA.lua")
	scriptPathB := filepath.Join(tmpDir, "scriptB.lua")
	luaA := `set_data("shared", "A")`
	luaB := `set_data("shared", "B")`
	require.NoError(t, os.WriteFile(scriptPathA, []byte(luaA), 0644))
	require.NoError(t, os.WriteFile(scriptPathB, []byte(luaB), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = loadScript(scriptPathA, "scriptA", tm)
	}()
	go func() {
		defer wg.Done()
		_ = loadScript(scriptPathB, "scriptB", tm)
	}()
	wg.Wait()
	val, ok := GlobalDataStore.GetData("shared")
	require.True(t, ok)
	require.Contains(t, []string{"A", "B"}, val)
	ds.Close()
}

func TestLoadScript_CustomLuaAction(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	luaContent := `
register_task("custom-lua-action", "test", function()
	local str = "hello"
	print(str)
end)
`
	scriptPath := filepath.Join(tmpDir, "custom.lua")
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)
	task := getTask(tm, "custom-lua-action")
	require.NotNil(t, task)
	require.NoError(t, task.Action())
}

func TestLoadScript_SandboxCheck(t *testing.T) {
	scriptEngines = nil
	tmpDir := t.TempDir()
	luaContent := `
register_task("sandbox-task", "test", function()
	if collectgarbage then
		error("collectgarbage should be disabled")
	end
	if loadfile then
		error("loadfile should be disabled")
	end
	print("Sandbox looks good")
end)
`
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tmpDir, "sandbox.lua"),
			[]byte(luaContent),
			0644,
		),
	)
	tm := core.NewTaskManager()
	err := LoadScripts(tmpDir, tm)
	require.NoError(t, err)
	task := getTask(tm, "sandbox-task")
	require.NotNil(t, task)
	err = task.Action()
	require.NoError(t, err)
}

func TestLoadScript_RepeatedDataSet(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "repeat_data.lua")
	luaContent := `
set_data("repeatKey", "first")
set_data("repeatKey", "second")
`
	require.NoError(t, os.WriteFile(scriptPath, []byte(luaContent), 0644))
	tm := core.NewTaskManager()
	ds, err := NewDataStore(tmpDir)
	require.NoError(t, err)
	GlobalDataStore = ds
	err = loadScript(scriptPath, "repeat_data", tm)
	require.NoError(t, err)
	val, ok := GlobalDataStore.GetData("repeatKey")
	require.True(t, ok)
	require.Equal(t, "second", val)
	ds.Close()
}

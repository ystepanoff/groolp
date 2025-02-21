package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ystepanoff/groolp/core"
	lua "github.com/yuin/gopher-lua"
)

var GlobalDataStore *DataStore

// LoadScripts() loads all *.lua scripts from scriptsDir in a sandboxed
// Lua enviroment and registers tasks with the TaskManager.
func LoadScripts(scriptsDir string, tm *core.TaskManager) error {
	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		return fmt.Errorf(
			"failed to read scripts directory %s: %w",
			scriptsDir,
			err,
		)
	}

	for _, fi := range files {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".lua") {
			continue
		}
		scriptPath := filepath.Join(scriptsDir, fi.Name())
		if err := loadScript(scriptPath, fi.Name(), tm); err != nil {
			fmt.Printf("Error loading script %s: %v\n", scriptPath, err)
		}
	}

	return nil
}

func loadScript(scriptPath, scriptName string, tm *core.TaskManager) error {
	L := lua.NewState()

	// Provide only a minimal set of safe libraries
	sandboxLuaState(L)
	engine := NewScriptEngine(scriptName)

	// Provide a function so user scripts can register tasks
	L.SetGlobal("register_task", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		desc := L.CheckString(2)
		fn := L.CheckFunction(3)

		var deps []string
		if L.GetTop() >= 4 {
			tbl := L.CheckTable(4)
			tbl.ForEach(func(key, value lua.LValue) {
				if key.Type() == lua.LTNumber && value.Type() == lua.LTString {
					deps = append(deps, value.String())
				}
			})
		}

		task := &core.Task{
			Name:         name,
			Description:  desc,
			Dependencies: deps,
			Action: func() error {
				L.Push(fn)
				if err := L.PCall(0, 0, nil); err != nil {
					return fmt.Errorf("lua runtime error: %v", err)
				}
				return nil
			},
		}

		if err := tm.Register(task); err != nil {
			L.Push(lua.LString(err.Error()))
			L.Error(lua.LString(err.Error()), 1)
			return 0
		}
		engine.tasks = append(engine.tasks, task)

		return 0
	}))

	if err := L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("lua script error in %s: %w", scriptPath, err)
	}

	fmt.Printf("Loaded script: %s\n", scriptPath)
	return nil
}

type luaLibrary struct {
	Name string
	Func lua.LGFunction
}

func sandboxLuaState(L *lua.LState) {
	safeLibs := []luaLibrary{
		{"_G", lua.OpenBase},
		{"table", lua.OpenTable},
		{"string", lua.OpenString},
		{"math", lua.OpenMath},
	}

	for _, safeLib := range safeLibs {
		L.Push(L.NewFunction(safeLib.Func))
		L.Push(lua.LString(safeLib.Name))
		L.Call(1, 0)
	}

	disabledFunctions := []string{
		"dofile",
		"loadfile",
		"load",
		"require",
		"collectgarbage",
	}

	for _, foo := range disabledFunctions {
		L.SetGlobal(foo, lua.LNil)
	}

	L.SetGlobal("run_command", L.NewFunction(func(L *lua.LState) int {
		cmdString := L.CheckString(1)

		code, err := runCommand(cmdString)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNumber(code))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("set_data", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val := L.CheckAny(2)
		switch v := val.(type) {
		case lua.LString:
			GlobalDataStore.SetData(key, string(v))
		case lua.LNumber:
			GlobalDataStore.SetData(key, float64(v))
		case lua.LBool:
			GlobalDataStore.SetData(key, bool(v))
		default:
			GlobalDataStore.SetData(key, v.String())
		}
		return 0
	}))

	L.SetGlobal("get_data", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val, ok := GlobalDataStore.GetData(key)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		switch v := val.(type) {
		case string:
			L.Push(lua.LString(v))
		case float64:
			L.Push(lua.LNumber(v))
		case bool:
			if v {
				L.Push(lua.LTrue)
			} else {
				L.Push(lua.LFalse)
			}
		default:
			L.Push(lua.LString(fmt.Sprintf("%v", v)))
		}
		return 1
	}))
}

func runCommand(cmdString string) (int, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/c", cmdString)
	} else {
		cmd = exec.Command("sh", "-c", cmdString)
	}
	output, err := cmd.CombinedOutput()
	os.Stdout.Write(output)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if runtime.GOOS == "windows" {
				if strings.Contains(
					string(output),
					"is not recognized as an internal or external command",
				) {
					return exitCode, fmt.Errorf(
						"command not found: %s",
						string(output),
					)
				}
			} else {
				if exitCode == 127 {
					return exitCode, fmt.Errorf("command not found: %s", string(output))
				}
			}
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

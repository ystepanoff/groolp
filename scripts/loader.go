package scripts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ystepanoff/groolp/core"
	lua "github.com/yuin/gopher-lua"
)

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

		task := &core.Task{
			Name:        name,
			Description: desc,
			// Wrap the Lua function as a Go closure
			Action: func() error {
				// Attempt to call the Lua function
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
}

package scripts

import (
	"fmt"

	"github.com/ystepanoff/groolp/core"
	lua "github.com/yuin/gopher-lua"
)

// ScriptEngine keeps info about each script’s Lua state and tasks
type scriptEngine struct {
	L     *lua.LState
	tasks []*core.Task
}

var scriptEngines []*scriptEngine

func NewScriptEngine() *scriptEngine {
	engine := &scriptEngine{
		L:     lua.NewState(),
		tasks: make([]*core.Task, 0),
	}
	scriptEngines = append(scriptEngines, engine)
	return engine
}

// CloseAllStates() closes all Lua states (at program end)
func CloseAllStates() {
	for _, eng := range scriptEngines {
		if eng.L != nil {
			fmt.Println("CLOSED ", eng)
			eng.L.Close()
			eng.L = nil
		}
	}
	scriptEngines = nil
}

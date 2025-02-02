package scripts

import (
	"github.com/ystepanoff/groolp/core"
	lua "github.com/yuin/gopher-lua"
)

// ScriptEngine keeps info about each script’s Lua state and tasks
type scriptEngine struct {
	Name  string
	L     *lua.LState
	tasks []*core.Task
}

var scriptEngines []*scriptEngine

func NewScriptEngine(name string) *scriptEngine {
	engine := &scriptEngine{
		Name:  name,
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
			eng.L.Close()
			eng.L = nil
		}
	}
	scriptEngines = nil
}

package scripts

import (
	"testing"

	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestCloseAllStates(t *testing.T) {
	scriptEngines = nil

	e1 := NewScriptEngine()
	e2 := NewScriptEngine()

	CloseAllStates()

	require.Nil(t, scriptEngines, "expected scriptEngines to be reset to nil")

	require.Panics(t, func() {
		e1.L.Push(lua.LNumber(42))
	}, "pushing to a closed LState should panic")

	require.Panics(t, func() {
		e2.L.Push(lua.LString("test"))
	}, "pushing to a closed LState should panic")
}

func TestCloseAllStates_EmptyList(t *testing.T) {
	scriptEngines = nil
	require.NotPanics(t, func() {
		CloseAllStates()
	}, "calling CloseAllStates() on an empty list should not panic")
	require.Nil(t, scriptEngines, "scriptEngines should remain nil")
}

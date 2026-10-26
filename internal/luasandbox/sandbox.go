package luasandbox

import (
	"context"
	"time"

	"github.com/yuin/gopher-lua"
)

const DefaultExecutionTimeout = time.Second

// NewState creates a Lua state with only the libraries required by randgen
// grammar files. Host filesystem, process, environment, package loading,
// debugging, channels, and coroutine APIs are not exposed.
func NewState() *lua.LState {
	l := lua.NewState(lua.Options{SkipOpenLibs: true})
	openLibrary(l, lua.BaseLibName, lua.OpenBase)
	openLibrary(l, lua.TabLibName, lua.OpenTable)
	openLibrary(l, lua.StringLibName, lua.OpenString)
	openLibrary(l, lua.MathLibName, lua.OpenMath)

	// OpenBase includes file-backed/module loaders and an internal register
	// printer, so remove them explicitly.
	for _, name := range []string{
		"collectgarbage", "dofile", "getfenv", "load", "loadfile", "loadstring",
		"module", "newproxy", "require", "setfenv", "_printregs",
	} {
		l.SetGlobal(name, lua.LNil)
	}
	return l
}

func openLibrary(l *lua.LState, name string, function lua.LGFunction) {
	l.Push(l.NewFunction(function))
	l.Push(lua.LString(name))
	l.Call(1, 0)
}

// RunString executes one grammar code block with a wall-clock deadline.
func RunString(l *lua.LState, script string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	l.SetContext(ctx)
	err := l.DoString(script)
	l.RemoveContext()
	cancel()
	return err
}

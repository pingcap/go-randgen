package luasandbox

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/gopher-lua"
)

func TestNewStateExposesOnlySafeLibraries(t *testing.T) {
	l := NewState()
	defer l.Close()

	for _, name := range []string{
		"os", "io", "debug", "package", "channel", "coroutine",
		"collectgarbage", "dofile", "getfenv", "load", "loadfile", "loadstring",
		"module", "newproxy", "require", "setfenv", "_printregs",
	} {
		assert.Equal(t, lua.LTNil, l.GetGlobal(name).Type(), name)
	}

	err := RunString(l, `result = string.format("%s-%d", string.upper("safe"), math.floor(2.9))`, time.Second)
	require.NoError(t, err)
	assert.Equal(t, "SAFE-2", l.GetGlobal("result").String())
}

func TestRunStringStopsInfiniteLoop(t *testing.T) {
	l := NewState()
	defer l.Close()

	started := time.Now()
	err := RunString(l, "while true do end", 20*time.Millisecond)

	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "deadline")
	assert.True(t, time.Since(started) < time.Second)

	require.NoError(t, RunString(l, "result = 1", time.Second))
	assert.Equal(t, lua.LNumber(1), l.GetGlobal("result"))
}

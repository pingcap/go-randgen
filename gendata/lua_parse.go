package gendata

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

func runLua(script string) (*lua.LState, error) {
	l := lua.NewState()
	defer l.Close()
	err := l.DoString(script)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// extract vals from two layer table by key1 and key2
// if key1 not exist, will have a error
// if key2 not exist, will return defaul
func extractSlice(l *lua.LState, key1 string, key2 string, defaul []string) ([]string, error) {
	key1Val := l.Env.RawGetString(key1)

	key1Table, ok := key1Val.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%s must be a lua Table", key1)
	}

	key2Val := key1Table.RawGetString(key2)
	if key2Val == lua.LNil {
		return defaul, nil
	}
	key2Table, ok := key2Val.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%s.%s must be a lua Table", key1, key2)
	}

	rows := make([]string, 0, key2Table.Len())

	key2Table.ForEach(func(key lua.LValue, value lua.LValue) {
		rows = append(rows, value.String())
	})

	return rows, nil
}


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

func extractSliceTable(tab *lua.LTable) []string {
	content := make([]string, 0, tab.Len())

	tab.ForEach(func(key lua.LValue, value lua.LValue) {
		content = append(content, value.String())
	})

	return content
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

	return extractSliceTable(key2Table), nil
}

func extractAllSlice(l *lua.LState, key string) (map[string][]string, error)  {
	val := l.Env.RawGetString(key)
	valTable, ok := val.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%s must be a lua Table", key)
	}

	res := make(map[string][]string)
	var err error
	valTable.ForEach(func(key2 lua.LValue, value lua.LValue) {
		table, ok := value.(*lua.LTable)
		if !ok {
			err = fmt.Errorf("%s.%s must be a lua Table", key, key2.String())
			return
		}

		res[key2.String()] = extractSliceTable(table)
	})

	return res, err
}


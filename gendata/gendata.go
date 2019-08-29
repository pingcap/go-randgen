package gendata

import (
	"github.com/yuin/gopher-lua"
)


type ZzConfig struct {
	Tables *Tables
	Fields *Fields
	Data   *Data
}

func ByZz(zz string) ([]string, error) {

	l ,err := runLua(zz)
	if err != nil {
		return nil, err
	}

	tables, err := newTables(l)
	if err != nil {
		return nil, err
	}

	fields, err := newFields(l)
	if err != nil {
		return nil, err
	}

	data, err := extractData(l)
	if err != nil {
		return nil, err
	}

	return  ByConfig(&ZzConfig{
		tables, fields, data,
	})
}

func ByConfig(config *ZzConfig) ([]string, error)  {
	return nil, nil
}

func extractData(l *lua.LState) (*Data, error) {
	numbers, err := extractSlice(l, "data", "numbers", nil)
	if err != nil {
		return nil, err
	}

	strings, err := extractSlice(l, "data", "strings", nil)
	if err != nil {
		return nil, err
	}

	o, _ := newOptions(nil, nil,"data", nil)
	o.addField("numbers", numbers)
	o.addField("strings", strings)

	return &Data{o}, nil
}


package gendata

import (
	"bytes"
	"fmt"
	"github.com/yuin/gopher-lua"
	"strconv"
	"strings"
)


type ZzConfig struct {
	Tables *Tables
	Fields *Fields
	Data   *Data
}

func newZzConfig(l *lua.LState) (*ZzConfig, error) {
	tables, err := newTables(l)
	if err != nil {
		return nil, err
	}

	fields, err := newFields(l)
	if err != nil {
		return nil, err
	}

	data, err := newData(l)
	if err != nil {
		return nil, err
	}

	return &ZzConfig{Tables:tables, Fields:fields, Data:data}, nil
}

func (z *ZzConfig) genDdls() ([]*tableStmt, []*fieldExec, error) {
	tableStmts, err := z.Tables.gen()
	if err != nil {
		return nil, nil, err
	}

	fieldStmts, fieldExecs, err := z.Fields.gen()
	if err != nil {
		return nil, nil, err
	}

	for _, tableStmt := range tableStmts {
		tableStmt.wrapInTable(fieldStmts)
	}

	return tableStmts, fieldExecs, nil
}

func ByZz(zz string) ([]string, error) {
	l ,err := runLua(zz)
	if err != nil {
		return nil, err
	}

	config, err := newZzConfig(l)
	if err != nil {
		return nil, err
	}

	return  ByConfig(config)
}

func ByConfig(config *ZzConfig) ([]string, error)  {
	tableStmts, fieldExecs, err := config.genDdls()
	if err != nil {
		return nil, err
	}

	recordGor := config.Data.getRecordGen(fieldExecs)
	row := make([]string, len(fieldExecs))

	sqls := make([]string, 0, len(tableStmts))
	for _, tableStmt := range tableStmts {
		sqls = append(sqls, tableStmt.ddl)
		valuesStmt := make([]string, 0, tableStmt.rowNum)
		for i := 0; i < tableStmt.rowNum; i++ {
			recordGor.oneRow(row)
			valuesStmt = append(valuesStmt, wrapInDml(strconv.Itoa(i), row))
		}
		sqls = append(sqls, wrapInInsert(tableStmt.name, valuesStmt))
	}

	return sqls, nil
}

const insertTemp = "insert into %s values %s"

func wrapInInsert(tableName string, valuesStmt []string) string {
	return fmt.Sprintf(insertTemp, tableName, strings.Join(valuesStmt, ","))
}

func wrapInDml(pk string, data []string) string {
	buf := &bytes.Buffer{}
	buf.WriteString("(" + pk)

	for _, d := range data {
		buf.WriteString(",")
		if d == "null" {
			buf.WriteString(d)
		} else {
			buf.WriteString("\"" + d + "\"")
		}
	}

	buf.WriteString(")")

	return buf.String()
}




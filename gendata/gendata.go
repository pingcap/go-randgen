package gendata

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pingcap/errors"
	"github.com/pingcap/go-randgen/gendata/generators"
	"github.com/pingcap/go-randgen/resource"
	"github.com/yuin/gopher-lua"
	"math/rand"
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

	return &ZzConfig{Tables: tables, Fields: fields, Data: data}, nil
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

func ByZz(zz string) ([]string, Keyfun, error) {
	// if zz is empty string, will use built-in default zz file
	if zz == "" {
		zzBs, err := resource.Asset("resource/default.zz.lua")
		if err != nil {
			return nil, nil, errors.Wrap(err, "default resource load fail")
		}
		zz = string(zzBs)
	}

	l, err := runLua(zz)
	if err != nil {
		return nil, nil, err
	}

	config, err := newZzConfig(l)
	if err != nil {
		return nil, nil, err
	}

	return ByConfig(config)
}

func ByConfig(config *ZzConfig) ([]string, Keyfun, error) {
	tableStmts, fieldExecs, err := config.genDdls()
	if err != nil {
		return nil, nil, err
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

	return sqls, NewKeyfun(tableStmts, fieldExecs), nil
}

type dbDriverError struct {
	driver string
	msg    string
}

func (e *dbDriverError) Error() string {
	return fmt.Sprintf("%s - %s", e.msg, e.driver)
}

// generate keyfun by db, it assumes all tables have the same fields
func ByDb(db *sql.DB, dbms string) (Keyfun, error) {
	// support different databases
	var rows *sql.Rows
	var err error

	if dbms == "sqlite3" {
		rows, err = db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	} else if dbms == "mysql" {
		rows, err = db.Query("show tables")
	} else {
		err = &dbDriverError{dbms, "Cannot retrieve the tables."}
	}

	if err != nil {
		return nil, err
	}

	tableStmts := make([]*tableStmt, 0)

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tableStmts = append(tableStmts, &tableStmt{name: tableName})
	}
	rows.Close()

	fieldExecs := make([]*fieldExec, 0)

	if len(tableStmts) > 0 {
		if dbms == "sqlite3" {
			rows, err = db.Query(fmt.Sprintf("PRAGMA table_info('%s');", tableStmts[0].name))
		} else if dbms == "mysql" {
			rows, err = db.Query(fmt.Sprintf("desc %s", tableStmts[0].name))
		} else {
			err = &dbDriverError{dbms, "Cannot retrieve the fields"}
		}

		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var fieldName, fieldType string
			if dbms == "sqlite3" {
				err = rows.Scan(&sql.RawBytes{}, &fieldName,
					&fieldType, &sql.RawBytes{},
					&sql.RawBytes{}, &sql.RawBytes{})
			} else if dbms == "mysql" {
				err = rows.Scan(&fieldName, &fieldType,
					&sql.RawBytes{}, &sql.RawBytes{},
					&sql.RawBytes{}, &sql.RawBytes{})
			}

			if err != nil {
				return nil, err
			}

			fieldExecs = append(fieldExecs, &fieldExec{name: fieldName, tp: fieldType})
		}

	}

	return NewKeyfun(tableStmts, fieldExecs), nil
}

const insertTemp = "insert into %s values %s"

func wrapInInsert(tableName string, valuesStmt []string) string {
	return fmt.Sprintf(insertTemp, tableName, strings.Join(valuesStmt, ","))
}

func wrapInDml(pk string, data []string) string {
	buf := &bytes.Buffer{}
	buf.WriteString("(" + pk)

	for _, d := range data {
		buf.WriteString("," + d)
	}

	buf.WriteString(")")

	return buf.String()
}

const (
	fInt = iota
	fChar
)

var fClass = map[string]int{
	"char":      fChar,
	"varchar":   fChar,
	"binary":    fChar,
	"varbinary": fChar,
	"integer":   fInt,
	"int":       fInt,
	"smallint":  fInt,
	"tinyint":   fInt,
	"mediumint": fInt,
	"bigint":    fInt,
}

type Keyfun map[string]func() (string, error)

func joinFields(fields []*fieldExec) string {
	strBuf := bytes.Buffer{}

	for i, f := range fields {
		if i == 0 {
			strBuf.WriteRune('`')
		} else {
			strBuf.WriteString("`,`")
		}

		strBuf.WriteString(f.name)

		if i == len(fields)-1 {
			strBuf.WriteRune('`')
		}
	}

	return strBuf.String()
}

var field_invariant = ""

func NewKeyfun(tables []*tableStmt, fields []*fieldExec) Keyfun {
	fieldsInt := make([]*fieldExec, 0)
	fieldsChar := make([]*fieldExec, 0)

	for _, fieldExec := range fields {
		if class, ok := fClass[fieldExec.dType()]; ok {
			switch class {
			case fInt:
				fieldsInt = append(fieldsInt, fieldExec)
			case fChar:
				fieldsChar = append(fieldsChar, fieldExec)
			}
		}
	}

	m := map[string]func() (string, error){
		"_table": func() (string, error) {
			if len(tables) == 0 {
				return "", errors.New("there is no table")
			}
			return tables[rand.Intn(len(tables))].name, nil
		},
		"_field": func() (string, error) {
			if len(fields) == 0 {
				return "", errors.New("there is no fields")
			}
			return "`" + fields[rand.Intn(len(fields))].name + "`", nil
		},

		"_field_invariant": func() (string, error) {
			if len(fields) == 0 {
				return "", errors.New("there is no fields")
			}
			// set the invariant
			if len(field_invariant) == 0 {
				field_invariant = "`" + fields[rand.Intn(len(fields))].name + "`"
			}
			// use the invariant
			return field_invariant, nil
		},

		"_field_int": func() (string, error) {
			if len(fieldsInt) == 0 {
				return "", errors.New("there is no int fields")
			}
			return "`" + fieldsInt[rand.Intn(len(fieldsInt))].name + "`", nil
		},
		"_field_int_list": func() (s string, e error) {
			if len(fieldsInt) == 0 {
				return "", errors.New("there is no int fields")
			}
			return joinFields(fieldsInt), nil
		},
		"_field_char": func() (string, error) {
			if len(fieldsChar) == 0 {
				return "", errors.New("there is no char fields")
			}
			return "`" + fieldsChar[rand.Intn(len(fieldsChar))].name + "`", nil
		},
		"_field_char_list": func() (s string, e error) {
			if len(fieldsChar) == 0 {
				return "", errors.New("there is no char fields")
			}
			return joinFields(fieldsChar), nil
		},
		"_field_list": func() (s string, e error) {
			if len(fields) == 0 {
				return "", errors.New("there is no char fields")
			}
			return joinFields(fields), nil
		},
	}

	// port from generators
	// digit -> _digit
	generators.Traverse(func(name string, generator generators.Generator) {
		m["_"+name] = func() (string, error) {
			return generator.Gen(), nil
		}
	})

	return Keyfun(m)
}

func (k Keyfun) Gen(key string) (string, bool, error) {
	if kf, ok := k[key]; ok {
		if res, err := kf(); err != nil {
			return res, true, err
		} else {
			return res, true, nil
		}
	}
	return "", false, nil
}

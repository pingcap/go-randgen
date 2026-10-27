package gendata

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pingcap/errors"
	"github.com/pingcap/go-randgen/gendata/generators"
	"github.com/pingcap/go-randgen/resource"
	"github.com/yuin/gopher-lua"
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
	defer l.Close()

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
	dbms = strings.ToLower(dbms)

	if dbms == "postgres" {
		rows, err = db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public';")
	} else if dbms == "sqlite3" {
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
		if dbms == "postgres" {
			rows, err = db.Query("SELECT table_name, column_name, data_type "+
				"FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1;",
				tableStmts[0].name)
		} else if dbms == "sqlite3" {
			rows, err = db.Query("SELECT name, type FROM pragma_table_info(?)", tableStmts[0].name)
		} else if dbms == "mysql" {
			rows, err = db.Query("desc " + quoteIdentifier(tableStmts[0].name, '`'))
		} else {
			err = &dbDriverError{dbms, "Cannot retrieve the fields"}
		}

		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var fieldName, fieldType string
			if dbms == "postgres" {
				err = rows.Scan(&sql.RawBytes{}, &fieldName, &fieldType)
			} else if dbms == "sqlite3" {
				err = rows.Scan(&fieldName, &fieldType)
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

	quote := func(name string) string {
		if dbms == "mysql" {
			return quoteIdentifier(name, '`')
		}
		return quoteIdentifier(name, '"')
	}
	return newKeyfun(tableStmts, fieldExecs, quote, quote), nil
}

func quoteIdentifier(name string, quote rune) string {
	q := string(quote)
	return q + strings.ReplaceAll(name, q, q+q) + q
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

func joinFields(fields []*fieldExec, quote func(string) string) string {
	strBuf := bytes.Buffer{}

	for i, f := range fields {
		if i > 0 {
			strBuf.WriteRune(',')
		}
		strBuf.WriteString(quote(f.name))
	}

	return strBuf.String()
}

func NewKeyfun(tables []*tableStmt, fields []*fieldExec) Keyfun {
	return newKeyfun(tables, fields, func(name string) string { return name },
		func(name string) string { return quoteIdentifier(name, '`') })
}

func newKeyfun(tables []*tableStmt, fields []*fieldExec,
	quoteTable, quoteField func(string) string) Keyfun {
	fieldsInt := make([]*fieldExec, 0)
	fieldsChar := make([]*fieldExec, 0)
	fieldInvariant := ""

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
			return quoteTable(tables[rand.Intn(len(tables))].name), nil
		},
		"_field": func() (string, error) {
			if len(fields) == 0 {
				return "", errors.New("there is no fields")
			}
			return quoteField(fields[rand.Intn(len(fields))].name), nil
		},

		"_field_invariant": func() (string, error) {
			if len(fields) == 0 {
				return "", errors.New("there is no fields")
			}
			// set the invariant
			if len(fieldInvariant) == 0 {
				fieldInvariant = quoteField(fields[rand.Intn(len(fields))].name)
			}
			// use the invariant
			return fieldInvariant, nil
		},

		"_field_int": func() (string, error) {
			if len(fieldsInt) == 0 {
				return "", errors.New("there is no int fields")
			}
			return quoteField(fieldsInt[rand.Intn(len(fieldsInt))].name), nil
		},
		"_field_int_list": func() (s string, e error) {
			if len(fieldsInt) == 0 {
				return "", errors.New("there is no int fields")
			}
			return joinFields(fieldsInt, quoteField), nil
		},
		"_field_char": func() (string, error) {
			if len(fieldsChar) == 0 {
				return "", errors.New("there is no char fields")
			}
			return quoteField(fieldsChar[rand.Intn(len(fieldsChar))].name), nil
		},
		"_field_char_list": func() (s string, e error) {
			if len(fieldsChar) == 0 {
				return "", errors.New("there is no char fields")
			}
			return joinFields(fieldsChar, quoteField), nil
		},
		"_field_list": func() (s string, e error) {
			if len(fields) == 0 {
				return "", errors.New("there is no char fields")
			}
			return joinFields(fields, quoteField), nil
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

package gendata

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"strings"
)

var fieldsTmpl = mustParse("fields", "`{{.fname}}` {{.types}} {{.sign}} {{.keys}}")

var fieldVars = []*varWithDefault{
	{
		"types",
		[]string{"int", "varchar", "date", "time", "datetime"},
	},
	{
		"keys",
		[]string{"undef", "key"},
	},
	// to ensure ignore efficient, sign should always be the last
	{
		"sign",
		[]string{"signed"},
	},
}

// https://dev.mysql.com/doc/refman/8.0/en/numeric-type-overview.html
var canUnSign = map[string]bool{
	"tinyint":   true,
	"smallint":  true,
	"mediumint": true,
	"int":       true,
	"integer":   true,
	"bigint":    true,
	"float":     true,
	"double":    true,
	"decimal":   true,
}

const enumVals = "('a','b','c','d','e','f','g','h','i','j','k','l'," +
	"'m','n','o','p','q','r','s','t','u','v','w','x','y','z')"

var fieldFuncs = map[string]func(text string, fname string, ctx *fieldExec) (target string,
	ignore bool, extraStmt *string, err error){
	"types": func(text string, fname string, ctx *fieldExec) (string, bool, *string, error) {
		index := strings.Index(text, "(")
		var tp string
		if index != -1 {
			tp = text[:index]
		} else {
			tp = text
		}
		tp = strings.ToLower(tp)
		if canUnSign[tp] {
			ctx.canUnSign = true
		}
		if tp == "set" || tp == "enum" {
			text = tp + enumVals
		}
		return text, false, nil, nil
	},
	"keys": func(text string, fname string, ctx *fieldExec) (string, bool, *string, error) {
		if text == "undef" {
			return "", false, nil, nil
		}
		extraStmt := fmt.Sprintf("key (`%s`)", fname)
		return "", false, &extraStmt, nil
	},
	// "signed" is sign, other is "unsigned"
	"sign": func(text string, fname string, ctx *fieldExec) (string, bool, *string, error) {
		if ctx.canUnSign {
			if text == "signed" {
				return "", false, nil, nil
			}
			ctx.unsign = true
			return "unsigned", false, nil, nil
		} else if text != "signed" {
			return "", true, nil, nil
		}

		return "", false, nil, nil
	},
}

type Fields struct {
	*options
}

func newFields(l *lua.LState) (*Fields, error) {
	o, err := newOptions(fieldsTmpl, l, "fields", fieldVars)
	if err != nil {
		return nil, err
	}

	return &Fields{o}, nil
}

func (f *Fields) gen() ([]string, []*fieldExec, error) {
	fnamePrefix := "col"

	m := make(map[string]string)
	stmts := make([]string, 0, f.numbers)
	extraStmts := make([]string, 0)
	fieldExecs := make([]*fieldExec, 0, f.numbers)

	err := f.traverse(func(cur []string) error {
		fExec := &fieldExec{}

		fname := fnamePrefix + "_" + strings.Join(cur, "_")
		extraNum := 0

		for i := range cur {
			field := f.fields[i]
			if field == "types" {
				fExec.tp = strings.ToLower(cur[i])
			}
			target, ignore, extraStmt, err := fieldFuncs[field](cur[i], fname, fExec)
			if err != nil {
				return err
			}
			// may be inefficient, prune tree may be better
			if ignore {
				// delete related extraNum
				extraStmts = extraStmts[0 : len(extraStmts)-extraNum]
				return nil
			}

			if extraStmt != nil {
				extraNum++
				extraStmts = append(extraStmts, *extraStmt)
			}
			m[field] = target
		}

		m["fname"] = fname
		fExec.name = fname

		fieldExecs = append(fieldExecs, fExec)
		stmts = append(stmts, f.format(m))
		return nil
	})

	if err != nil {
		return nil, fieldExecs, err
	}

	stmts = append(stmts, extraStmts...)

	return stmts, fieldExecs, nil
}

type fieldExec struct {
	canUnSign bool
	unsign    bool
	name      string
	// tp writen by user zz file
	tp string
}

func (f *fieldExec) dType() string {
	index := strings.Index(f.tp, "(")
	if index == -1 {
		return f.tp
	}
	return f.tp[:index]
}

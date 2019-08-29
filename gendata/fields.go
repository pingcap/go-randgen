package gendata

import (
	"bytes"
	"github.com/yuin/gopher-lua"
)

type fieldsCtx struct {
	canUnSign bool
}

var fieldsTmpl = mustParse("fields", `{{.fname}} {{.types}} {{.sign}} {{.keys}}`)

var fieldVars = []*varWithDefault{
	{
		"types",
		[]string{"int", "varchar", "date", "time", "datetime"},
	},
	{
		"sign",
		[]string{"undef"},
	},
	{
		"keys",
		[]string{"undef", "key"},
	},
}

// https://dev.mysql.com/doc/refman/8.0/en/numeric-type-overview.html
var canUnSign = map[string]bool{
	"tinyint": true,
	"smallint": true,
	"mediumint": true,
	"int": true,
	"integer": true,
	"bigint": true,
	"float": true,
	"double": true,
	"decimal": true,
}


var fieldFuncs = map[string]func(string, *fieldsCtx) (target string, ignore bool, err error){
	"types": func(text string, ctx *fieldsCtx) (string, bool, error) {
		if canUnSign[text] {
			ctx.canUnSign = true
		}
		return text, false, nil
	},
	// "signed" is sign, other is "unsigned"
	"sign": func(text string, ctx *fieldsCtx) (string, bool, error) {
		if ctx.canUnSign {
			if text == "signed" {
				return "", false, nil
			}
			return "unsigned", false, nil
		} else if text != "signed" {
			return "", true, nil
		}

		return "", false, nil
	},
	"keys": func(text string, ctx *fieldsCtx) (string, bool, error) {
		if text == "undef" {
			return "", false, nil
		}
		return "unique key", false, nil
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

func (f *Fields) gen() ([]string, error) {
	fnamePrefix := "col"

	fnameBuf := &bytes.Buffer{}
	m := make(map[string]string)
	stmts := make([]string, 0, f.numbers)

	err := f.traverse(func(cur []string) error {
		fnameBuf.Reset()
		fnameBuf.WriteString(fnamePrefix)
		fCtx := &fieldsCtx{}

		for i := range cur {
			field := f.fields[i]
			fnameBuf.WriteString("_" + cur[i])
			target, ignore, err := fieldFuncs[field](cur[i], fCtx)
			if err != nil {
				return err
			}
			// may be inefficient, prune tree may be better
			if ignore {
				return nil
			}

			m[field] = target
		}

		m["fname"] = fnameBuf.String()

		stmts = append(stmts, f.format(m))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return stmts, nil
}

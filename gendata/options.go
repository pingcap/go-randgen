package gendata

import (
	"bytes"
	"log"
	"text/template"

	"github.com/pingcap/errors"
	"github.com/yuin/gopher-lua"
)

type varWithDefault struct {
	name   string
	defaul []string
}

type options struct {
	// numbers of all option combination
	numbers int
	// fields sequence
	fields []string
	pos    []int
	datas  map[string][]string
	tmpl   *template.Template
}

func newOptions(tmpl *template.Template, l *lua.LState, option string,
	vars []*varWithDefault) (*options, error) {
	o := &options{
		datas: make(map[string][]string),
		tmpl:  tmpl,
	}

	for _, ovar := range vars {
		vals, err := extractSlice(l, option, ovar.name, ovar.defaul)
		if err != nil {
			return nil, err
		}
		o.addField(ovar.name, vals)
	}

	return o, nil
}

func (o *options) addField(fieldName string, optionData []string) {
	// can only add once
	if _, ok := o.datas[fieldName]; ok {
		return
	}
	o.datas[fieldName] = optionData
	o.fields = append(o.fields, fieldName)
	o.pos = append(o.pos, 0)
	if o.numbers != 0 {
		o.numbers *= len(optionData)
	} else {
		o.numbers = len(optionData)
	}
}

func (o *options) format(vals map[string]string) string {
	buf := &bytes.Buffer{}
	err := o.tmpl.Execute(buf, vals)
	if err != nil {
		log.Fatalf("template execute error ,%+v\n", errors.Trace(err))
	}
	return buf.String()
}

// traverse all conbination of options
func (o *options) traverse(handler func(cur []string) error) error {
	container := make([]string, len(o.fields))
	return traverse(o, container, 0, handler)
}

func traverse(o *options, container []string, index int, handler func([]string) error) error {
	if index == len(o.fields) {
		return handler(container)
	}
	data := o.datas[o.fields[index]]
	for _, d := range data {
		container[index] = d
		err := traverse(o, container, index+1, handler)
		if err != nil {
			return err
		}
	}

	return nil
}

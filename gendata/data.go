package gendata

import (
	"github.com/yuin/gopher-lua"
	"github.com/dqyuan/go-randgen/gendata/generators"
	"log"
	"math/rand"
	"runtime/debug"
	"strings"
)

const numberType = "numbers"
const blobsType = "blobs"
const temporalType = "temporals"
const enumType = "enum"

// default summary type
const stringsType = "strings"

// https://github.com/DQinYuan/randgenx/blob/master/lib/GenTest/App/Gendata.pm#L218
// https://github.com/DQinYuan/randgenx/blob/master/lib/GenTest/App/Gendata.pm#L505
var summaryType = map[string]string{
	"int":     numberType,
	"bigint":  numberType,
	"float":   numberType,
	"double":  numberType,
	"decimal": numberType,
	"numeric": numberType,
	"fixed":   numberType,
	"bool":    numberType,
	"bit":     numberType,
	"blob":    blobsType,
	"text":    blobsType,
	"binary":  blobsType,
	"date":    temporalType,
	"time":    temporalType,
	"year":    temporalType,
	"enum":    enumType,
	"set":     enumType,
}

var defaultData = []*varWithDefault{
	{
		name:numberType,
		defaul:[]string{"digit", "digit", "digit", "digit", "null"},
	},
	{
		name:stringsType,
		defaul:[]string{"letter", "letter", "letter", "letter", "null"},
	},
	{
		name:blobsType,
		defaul:[]string{"letter", "letter", "letter", "letter", "null"},
	},
	{
		name:temporalType,
		defaul:[]string{"date", "time", "datetime", "year", "timestamp", "null"},
	},
	{
		name:enumType,
		defaul:[]string{"letter", "letter", "letter", "letter", "null"},
	},
}

type Data struct {
	gens map[string]generators.Generator
}

func newData(l *lua.LState) (*Data, error) {
	datas, err := extractAllSlice(l, "data")
	if err != nil {
		return nil, err
	}

	gens := make(map[string]generators.Generator)
	for name, genNames := range datas {
		gens[name] = composeFromGenName(genNames)
	}

	for _, dval := range defaultData {
		_, ok := gens[dval.name]
		if !ok {
			gens[dval.name] = composeFromGenName(dval.defaul)
		}
	}

	return &Data{gens}, nil
}

func composeFromGenName(genNames []string ) generators.Generator {
	gs := make([]generators.Generator, 0)
	for _, gName := range genNames {
		gor := generators.Get(gName)
		if gor != nil {
			gs = append(gs, gor)
		} else { // constant
			gs = append(gs, &constGen{gName})
		}
	}

	return &composeGen{gs}
}

func (d *Data) getRecordGen(fields []*fieldExec) recordGen {
	gens := make([]generators.Generator, 0)
	for _, f := range fields {
		// full type name
		name := f.tp
		generator, ok := d.gens[name]
		if ok {
			gens = append(gens, generator)
			continue
		}

		// simple type name
		index := strings.Index(name, "(")
		if index != -1 {
			name = name[:index]
			generator, ok := d.gens[name]
			if ok {
				gens = append(gens, generator)
				continue
			}
		}

		// finally summary name
		summaryName, ok := summaryType[name]
		if !ok {
			summaryName = stringsType
		}
		generator, ok = d.gens[summaryName]
		if !ok {
			log.Fatalf("shouldn't run here, summary name %s \n %s", summaryName, debug.Stack())
		}

		if f.unsign {
			gens = append(gens, &unsignGen{generator, 10, "1"})
		} else {
			gens = append(gens, generator)
		}
	}

	return recordGen(gens)
}

type recordGen []generators.Generator

func (r recordGen) oneRow(row []string)  {
	if len(r) != len(row) {
		log.Fatalf("record gen illegal, expect len: %d, real row container len %d\n",
			len(r), len(row))
	}

	for i := range r {
		row[i] = r[i].Gen()
	}
}

type composeGen struct {
	gs []generators.Generator
}

func (g *composeGen) Gen() string {
	return g.gs[rand.Intn(len(g.gs))].Gen()
}

type constGen struct {
	constant string
}

func (c *constGen) Gen() string {
	return c.constant
}

type unsignGen struct {
	gen generators.Generator
	retryNum int
	defaul string
}

func (u *unsignGen) Gen() string {
	for i := 0; i < u.retryNum; i++ {
		cur := u.gen.Gen()
		if !strings.HasPrefix(cur, "-") {
			return cur
		}
	}

	return u.defaul
}
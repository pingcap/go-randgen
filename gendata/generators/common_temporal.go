package generators

import "bytes"

/*
0    1  2  3  4  5  6
yyyy-MM-dd HH:mm:ss.SSS
*/
const (
	yyyy = iota
	MM
	dd
	HH
	mm
	ss
	SSS
)

type Temporal struct {
	from int
	to   int
}

type genAndPrefix struct {
	gen    Generator
	prefix string
}

var tplComponents = []genAndPrefix{
	{
		newInt(2000, 2019, "%.4d"), // year
		"",
	},
	{
		newInt(1, 12, "%.2d"), // month
		"-",
	},
	{
		newInt(1, 28, "%.2d"), // day
		"-",
	},
	{
		newInt(0, 23, "%.2d"), // hour
		" ",
	},
	{
		newInt(0, 59, "%.2d"), //minute
		":",
	},
	{
		newInt(0, 59, "%.2d"), //second
		":",
	},
	{
		newInt(0, 999999, ""), //microsecond
		".",
	},
}

func newTemporal(from int, to int) *Temporal {
	return &Temporal{from, to}
}

func (t *Temporal) Gen() string {
	buf := &bytes.Buffer{}
	buf.WriteString(tplComponents[t.from].gen.Gen())

	for i := t.from + 1; i <= t.to; i++ {
		genpre := tplComponents[i]
		buf.WriteString(genpre.prefix + genpre.gen.Gen())
	}

	return `'` + buf.String() + `'`
}

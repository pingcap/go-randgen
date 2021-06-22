package yacc_parser

import (
	"fmt"
	_ "net/http/pprof"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	originAndExpecs := []struct {
		origin string
		expect []string
	}{
		{
			`column_attribute_list: column_attribute_list column_attribute | column_attribute`,
			[]string{"column_attribute_list", ":", "column_attribute_list", "column_attribute",
				"|", "column_attribute"},
		},
		{
			`this: is a test with 'colon appears inside a string :)'`,
			[]string{"this", ":", "is", "a", "test", "with", "'colon appears inside a string :)'"},
		},
		{
			`a: 'b' c`,
			[]string{"a", ":", "'b'", "c"},
		},
		{
			`a: '"b' "'c"`,
			[]string{"a", ":", `'"b'`, `"'c"`},
		},
		{
			`a: 'b"' "c'"`,
			[]string{"a", ":", `'b"'`, `"c'"`},
		},
		{
			`a: # this is a comment # o
'b"' "c'"`,
			[]string{"a", ":", "# this is a comment # o\n", `'b"'`, `"c'"`},
		},
		{
			`a: b,d m; count(cc)`,
			[]string{"a", ":", "b", ",", "d", "m", ";", "count", "(", "cc", ")"},
		},
		{
			`a: /* this is
a muti line comment
*/
'b"' /*sss*/ "c'"`,
			[]string{"a", ":", "/* this is\na muti line comment\n*/", `'b"'`, "/*sss*/", `"c'"`},
		},
		{
			`a: m dd {a = 1;b="aaa"; print(m)} ddd | haha {a = 2 * a; print(a)} nana`,
			[]string{"a", ":", "m", "dd", `{a = 1;b="aaa"; print(m)}`, "ddd", "|",
				"haha", "{a = 2 * a; print(a)}", "nana"},
		},
		{
			`query: select @A := 'sdwe'`,
			[]string{"query", ":", "select", "@A", ":=", "'sdwe'"},
		},
		{
			`t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' t3 'g'

t2: 'd'
    | t3 'e'

t3: 'f'
    | 'g' 'h'
	| 'i'`,
			[]string{`t1`, `:`, `'a'`, `'b'`, `t2`, `|`,
				`'c'`, `t3`, `|`, `t2`, `'f'`, `t3`, `'g'`, `t2`,
				`:`, `'d'`, `|`, `t3`, `'e'`, `t3`, `:`, `'f'`,
				`|`, `'g'`, `'h'`, `|`, `'i'`},
		},
		{
			`
{i=1
f1={a = 1, b = 2}
f2={a = 2, b = 3}
arr={4, 6, 'undef'}
}

t1: c | o | p
`,
			[]string{"{i=1\nf1={a = 1, b = 2}\nf2={a = 2, b = 3}\narr={4, 6, 'undef'}\n}",
				"t1", ":", "c", "|", "o", "|", "p"},
		},
		{
			`aasf " oo lp '`,
			[]string{"aasf", "\"", "oo", "lp", "'"},
		},
		{
			`
quote:
   "

test:
   quote '' ' quote "
`,
			[]string{"quote", ":", "\"", "test", ":", "quote", "''", "'", "quote", "\""},
		},
		{
			`{print("{")} m {print("}")}`,
			[]string{`{print("{")}`, `m`, `{print("}")}`},
		},
		{
			`{print("\"{");print('\'}')}llll`,
			[]string{`{print("\"{");print('\'}')}`, `llll`},
		},
		{
			`{print('"')}select{print('"')}`,
			[]string{`{print('"')}`, `select`, `{print('"')}`},
		},
		{
			`dsd /* dd ttee `,
			[]string{"dsd", "/*", "dd", "ttee"},
		},
		{
			`dsd { { dd ttee `,
			[]string{"dsd", "{", "{", "dd", "ttee"},
		},
		{
			`
{
-- {
--[==[
}
]==]
asd
}
`,
			[]string{"{\n-- {\n--[==[\n}\n]==]\nasd\n}"},
		},
	}

	for _, originAndExpec := range originAndExpecs {
		assertExpectedTokenResult(t, originAndExpec.origin, originAndExpec.expect)
	}

}

func assertExpectedTokenResult(t *testing.T, origin string, expected []string) {
	withTokenizeResult(t, origin, func(idx int, s string) {
		assert.Equal(t, expected[idx], s, origin)
	})
}

func withTokenizeResult(t *testing.T, origin string, visitor func(index int, tkn string)) {
	next := Tokenize(&RuneSeq{Runes: []rune(origin)})
	for i := 0; ; i++ {
		tkn, err := next()
		assert.Equal(t, nil, err)
		if isEOF(tkn) {
			break
		}
		visitor(i, tkn.OriginString())
	}
}

func TestRuneSeq(t *testing.T) {
	testStr := "t名称est哈哈"
	seq := &RuneSeq{Runes: []rune(testStr)}
	r, _ := seq.ReadRune()
	assert.Equal(t, 't', r)
	r, _ = seq.ReadRune()
	assert.Equal(t, '名', r)
	r, _ = seq.ReadRune()
	assert.Equal(t, '称', r)

	seq.UnreadRune()
	seq.UnreadRune()
	r, _ = seq.ReadRune()
	assert.Equal(t, '名', r)
}

func TestSimpleTokenPrint(t *testing.T) {
	t.SkipNow()
	origin := `
{
-- {
--[==[
}
]==]
asd
}
`

	next := Tokenize(&RuneSeq{Runes: []rune(origin)})
	for {
		tkn, err := next()
		assert.Equal(t, nil, err)

		fmt.Println("=========")
		fmt.Printf("%T\n", tkn)
		fmt.Println(tkn.OriginString())

		if isEOF(tkn) {
			break
		}
	}
}

package yacc_parser

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	_ "net/http/pprof"
	"testing"
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
			[]string{"this", ":", "is", "a", "test", "with", "colon appears inside a string :)"},
		},
		{
			`a: 'b' c`,
			[]string{"a", ":", "b", "c"},
		},
		{
			`a: '"b' "'c"`,
			[]string{"a", ":", `"b`, `'c`},
		},
		{
			`a: 'b"' "c'"`,
			[]string{"a", ":", `b"`, `c'`},
		},
		{
			`a: # this is a comment # o
'b"' "c'"`,
			[]string{"a", ":", "# this is a comment # o\n", `b"`, `c'`},
		},
		{
			`a: /* this is
a muti line comment
*/
'b"' /*sss*/ "c'"`,
			[]string{"a", ":", "/* this is\na muti line comment\n*/", `b"`, "/*sss*/", `c'`},
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
			[]string{`t1`, `:`, `a`, `b`, `t2`, `|`,
				`c`, `t3`, `|`, `t2`, `f`, `t3`, `g`, `t2`,
				`:`, `d`, `|`, `t3`, `e`, `t3`, `:`, `f`,
				`|`, `g`, `h`, `|`, `i`},
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
	next := Tokenize(bytes.NewBufferString(origin))
	for i := 0; ; i++ {
		tkn, err := next()
		assert.Equal(t, nil, err)
		if isEOF(tkn) {
			break
		}
		visitor(i, tkn.ToString())
	}
}

func TestSimpleTokenPrint(t *testing.T) {
	origin := `a: /* hahah
faad
*/

'"b' "'c"`

	next := Tokenize(bytes.NewBufferString(origin))
	for {
		tkn, err := next()
		assert.Equal(t, nil, err)

		fmt.Println("=========")
		fmt.Printf("%T\n", tkn)
		fmt.Println(tkn.ToString())

		if isEOF(tkn) {
			break
		}
	}
}
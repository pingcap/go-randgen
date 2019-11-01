package grammar

import (
	"fmt"
	"github.com/pingcap/go-randgen/gendata"
	"github.com/pingcap/go-randgen/grammar/sql_generator"
	"github.com/stretchr/testify/assert"
	"testing"
)

type yyTestCase struct {
	name      string
	yy        string
	num       int
	keyFun    gendata.Keyfun
	simpleExp string
	expected  func(string) bool
	expSeq    []string
}

func TestIter(t *testing.T) {
	cases := []*yyTestCase{
		{
			name: "test embeded lua",
			yy: `
{
f={a=1, b=3}
arr={0,2,3,4}
}

query:
  {print(arr[f.a])} | {print(arr[f.b])}
`,
			num: 10,
			expected: func(sql string) bool {
				if sql == "0" || sql == "3" {
					return true
				}
				return false
			},
		},
		{
			name: "test first write",
			yy: `
query:
	{n=1} select

select:
    sub_select | haha_select

sub_select:
    {n = 2} SELECT

haha_select:
    {m = 4} SELECT
`,
			num:       50,
			simpleExp: "SELECT",
		},
		{
			name: "test semi colon",
			yy: `
query:
	select ; create

select:
    SET @stmt = "FFF";
	PREPARE stmt FROM @stmt_create ; EXECUTE stmt ;
	EXECUTE stmt;

create:
	CREATE OOO; CCC
`,
			num: 6,
			expSeq: []string{
				`SET @stmt = "FFF"`,
				`PREPARE stmt FROM @stmt_create`,
				`EXECUTE stmt`,
				`EXECUTE stmt`,
				`CREATE OOO`,
				`CCC`,
			},
		},
		{
			name: "test pre space",
			yy: `
query: frame_clause

frame_units:
    RANGE

frame_between:
	BETWEEN

frame_clause:
	frame_units frame_between
`,
			num:       6,
			simpleExp: "RANGE BETWEEN",
		},
		{
			name: "test key word",
			yy: `
query:
    A _table B _field
`,
			keyFun: map[string]func() (string, error){
				"_table": func() (string, error) {
					return "aaa_tabl", nil
				},
				"_field": func() (string, error) {
					return "ffff", nil
				},
			},
			num:       10,
			simpleExp: "A aaa_tabl B ffff",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			iterator, err := NewIter(c.yy, "query", 5,
				c.keyFun, false)
			assert.Equal(t, nil, err)

			iterator.Visit(sql_generator.FixedTimesVisitor(func(i int, sql string) {
				if c.expected != nil {
					assert.Condition(t, func() (success bool) {
						return c.expected(sql)
					})
				} else if c.expSeq != nil {
					assert.Equal(t, c.expSeq[i], sql)
				} else {
					assert.Equal(t, c.simpleExp, sql)
				}
			}, c.num))
		})
	}

}

func TestMaxRetry(t *testing.T) {
	recurYy := `
query:
   select

select:
   SELECT select
`
	iterator, err := NewIter(recurYy, "query", 5,
		nil,  false)
	assert.Equal(t, nil, err)

	err = iterator.Visit(func(sql string) bool {
		return false
	})

	assert.Equal(t,
		"recursive num exceed max loop back 5\n [query select select select select select]",
		err.Error())
}

var halfRecur = `
query:
   select_char

select_char:
   haha haha select_char
   | SELECT * FROM mmm
`

func TestRecur(t *testing.T) {
	iter, err := NewIter(halfRecur, "query", 1, nil, false)
	assert.Equal(t, nil, err)

	counter := 0
	err = iter.Visit(sql_generator.FixedTimesVisitor(func(i int, sql string) {
		assert.Equal(t, counter, i)
		counter++

		assert.Equal(t, "SELECT * FROM mmm", sql)
	}, 100))

	assert.Equal(t, nil, err)
}

const yy = `
query:
    {if(a==nil) then a = 1 end} select

select:
    SELECT
           fieldA,
           fieldB,
           {print(string.format("field%d", a)); a = a + 1}
    FROM (
	SELECT _field AS fieldA, _field AS fieldB
	FROM _table
	ORDER BY LOWER(fieldA), LOWER(fieldB)
    ) as t
    WINDOW w AS (ORDER BY fieldA);
`

func TestByYySimplePrint(t *testing.T) {
	t.SkipNow()

	iter, err := NewIter(yy, "query", 5, map[string]func() (string, error){
		"_table": func() (string, error) {
			return "aaa_tabl", nil
		},
		"_field": func() (string, error) {
			return "ffff", nil
		},
	}, false)
	assert.Equal(t, nil, err)

	iter.Visit(sql_generator.FixedTimesVisitor(func(_ int, sql string) {
		fmt.Println(sql)
		fmt.Println("===============")
	}, 10))
}

func TestPathInfo(t *testing.T) {
	t.SkipNow()
	code := `
query:
    select

select:
    SELECT func(value)

func:
/*    | UPPER
    | ASCII
    | CONCAT */
    | QUOTE

value:
    1 | 2 | 256 | 23445 | NULL
`
	iter ,err := NewIter(code, "query", 5, nil, false)
	assert.Equal(t, nil, err)

	err = iter.Visit(sql_generator.FixedTimesVisitor(func(i int, sql string) {
		fmt.Println(sql)
		for _, production := range iter.PathInfo().ProductionSet.Productions {
			fmt.Println(production.Head.OriginString())
		}
		for _, seq := range iter.PathInfo().SeqSet.Seqs {
			fmt.Println(seq.String())
		}
		fmt.Println("============")
	}, 5))
	assert.Equal(t, nil, err)
}



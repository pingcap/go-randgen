package grammar

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const yyWithOutKeyword = `
query:
    select

select:
    SELECT
        fields
    FROM (
        select
    ) as t
    WHERE t.val > 10
    | SELECT *
	FROM table1
	ORDER BY LOWER(fieldA), LOWER(fieldB)

fields:
    fieldA
    | fieldB
    | fieldA, fieldB
`

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

func TestByYyWithoutKeyword(t *testing.T) {
	t.SkipNow()
	num := 10
	sqls, err := ByYy(yyWithOutKeyword, num, "query", nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, num, len(sqls))

	for _, sql := range sqls {
		fmt.Println(sql)
	}
}


func TestByYy(t *testing.T) {
	t.SkipNow()
	sqls, err := ByYy(yy, 10, "query", map[string]func() string{
		"_table": func() string {
			return "aaa_tabl"
		},
		"_field": func() string {
			return "ffff"
		},
	})
	assert.Equal(t, nil, err)

	for _, sql := range sqls {
		fmt.Println(sql)
	}
}



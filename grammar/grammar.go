package grammar

import (
	"bytes"
	"go-randgen/grammar/sql_generator"
	"go-randgen/grammar/yacc_parser"
)

func ByYy(yy string, num int, root string) ([]string, error) {
	reader := bytes.NewBufferString(yy)
	productions := yacc_parser.Parse(yacc_parser.Tokenize(reader))

	sqlIter := sql_generator.GenerateSQLRandomly(productions, root)

	sqls := make([]string, 0, num)
	for i := 0; i < num; i++ {
		sqls = append(sqls, sqlIter.Next())
	}

	return sqls, nil
}

package grammar

import (
	"fmt"
	"go-randgen/gendata"
	"go-randgen/grammar/sql_generator"
	"go-randgen/grammar/yacc_parser"
	"log"
)

const maxRetry = 10

func ByYy(yy string, num int, root string, maxRecursive int,
	keyFunc gendata.Keyfun, debug bool) ([]string, error) {
	reader := &yacc_parser.RuneSeq{Runes:[]rune(yy), Pos:0}
	codeblocks, productions, err := yacc_parser.Parse(yacc_parser.Tokenize(reader))
	if err != nil {
		return nil, err
	}

	sqlIter, err := sql_generator.GenerateSQLRandomly(codeblocks,
		productions, keyFunc, root, maxRecursive, debug)
	if err != nil {
		return nil, err
	}

	sqls := make([]string, 0, num)
	counter := 0
	for i := 0; i < num; {
		sql, err := sqlIter.Next()
		if err != nil{
			counter++
			if counter > maxRetry {
				return nil, fmt.Errorf("next retry num exceed %d, %v", maxRetry, err)
			}
			continue
		}

		if debug {
			log.Println(sql)
		}

		sqls = append(sqls, sql)
		i++
		counter = 0
	}

	return sqls, nil
}

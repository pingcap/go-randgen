package grammar

import (
	"github.com/dqinyuan/go-randgen/gendata"
	"github.com/dqinyuan/go-randgen/grammar/sql_generator"
	"github.com/dqinyuan/go-randgen/grammar/yacc_parser"
)

// get Iterator by yy,
// note that this iterator is not thread safe
func NewIter(yy string, root string, maxRecursive int,
	keyFunc gendata.Keyfun, debug bool) (sql_generator.SQLIterator, error) {
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

	return sqlIter, nil
}

func ByYy(yy string, num int, root string, maxRecursive int,
	keyFunc gendata.Keyfun, debug bool) ([]string, error) {

	sqlIter, err := NewIter(yy, root, maxRecursive, keyFunc, debug)
	if err != nil {
		return nil, err
	}

	sqls := make([]string, 0, num)
	for i := 0; i < num; i++ {
		sql, err := sqlIter.NextWithRetry()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

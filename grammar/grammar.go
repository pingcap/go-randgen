package grammar

import (
	"github.com/pingcap/go-randgen/gendata"
	"github.com/pingcap/go-randgen/grammar/sql_generator"
	"github.com/pingcap/go-randgen/grammar/yacc_parser"
)

// get Iterator by yy,
// note that this iterator is not thread safe
func NewIter(yy string, root string, maxRecursive int,
	keyFunc gendata.Keyfun, analyze bool, debug bool) (sql_generator.SQLIterator, error) {
	reader := &yacc_parser.RuneSeq{Runes:[]rune(yy), Pos:0}
	codeblocks, productions, err := yacc_parser.Parse(yacc_parser.Tokenize(reader))
	if err != nil {
		return nil, err
	}

	sqlIter, err := sql_generator.GenerateSQLRandomly(codeblocks,
		productions, keyFunc, root, maxRecursive, analyze, debug)
	if err != nil {
		return nil, err
	}

	return sqlIter, nil
}

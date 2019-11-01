package grammar

import (
	"github.com/pingcap/go-randgen/gendata"
	"github.com/pingcap/go-randgen/grammar/sql_generator"
	"github.com/pingcap/go-randgen/grammar/yacc_parser"
)

// get Iterator by yy,
// note that this iterator is not thread safe
func NewIter(yy string, root string, maxRecursive int,
	keyFunc gendata.Keyfun, debug bool) (sql_generator.SQLIterator, error) {

	codeblocks, _, productionMap, err := Parse(yy)
	if err != nil {
		return nil, err
	}

	sqlIter, err := sql_generator.GenerateSQLRandomly(codeblocks,
		productionMap, keyFunc, root, maxRecursive, debug)
	if err != nil {
		return nil, err
	}

	return sqlIter, nil
}

func initProductionMap(productions []*yacc_parser.Production) map[string]*yacc_parser.Production {
	// Head string -> production
	productionMap := make(map[string]*yacc_parser.Production)
	for _, production := range productions {
		if pm, exist := productionMap[production.Head.OriginString()]; exist {
			pm.Alter = append(pm.Alter, production.Alter...)
			productionMap[production.Head.OriginString()] = pm
			continue
		}
		productionMap[production.Head.OriginString()] = production
	}

	return productionMap
}

func Parse(yy string) ([]*yacc_parser.CodeBlock, []*yacc_parser.Production,
	map[string]*yacc_parser.Production, error) {
	reader := &yacc_parser.RuneSeq{Runes:[]rune(yy), Pos:0}
	codeblocks, productions, err := yacc_parser.Parse(yacc_parser.Tokenize(reader))
	if err != nil {
		return nil, nil, nil, err
	}

	return codeblocks, productions, initProductionMap(productions), nil
}
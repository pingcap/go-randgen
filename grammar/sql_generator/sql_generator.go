package sql_generator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yuin/gopher-lua"
	"github.com/dqinyuan/go-randgen/gendata"
	"github.com/dqinyuan/go-randgen/grammar/yacc_parser"
	"io"
	"log"
	"math/rand"
)

const maxBuildTreeTime = 10 * 1000

type node interface {
	walk() bool
	materialize(writer io.StringWriter) error
	loopbackDetection(productionName string, sameParent uint) bool
}

type literalNode struct {
	value string
}

func (ln *literalNode) walk() bool {
	return true
}

func (ln *literalNode) materialize(writer io.StringWriter) error {
	if len(ln.value) != 0 {
		_, err := writer.WriteString(ln.value)
		if err != nil {
			return err
		}
		_, err = writer.WriteString(" ")
		if err != nil {
			return err
		}
	}
	return nil
}

func (ln *literalNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	panic("unreachable")
}

// SQLIterator is a iterator interface of sql generator
type SQLIterator interface {
	// HasNext returns whether the iterator exists next sql case
	HasNext() bool

	// Next returns next sql case in iterator
	Next() (string, error)
}

func initProductionMap(productions []yacc_parser.Production) map[string]yacc_parser.Production {
	// Head string -> production
	productionMap := make(map[string]yacc_parser.Production)
	for _, production := range productions {
		if pm, exist := productionMap[production.Head.ToString()]; exist {
			pm.Alter = append(pm.Alter, production.Alter...)
			productionMap[production.Head.ToString()] = pm
			continue
		}
		productionMap[production.Head.ToString()] = production
	}

	return productionMap
}

// SQLRandomlyIterator is a iterator of sql generator
type SQLRandomlyIterator struct {
	productionName string
	productionMap  map[string]yacc_parser.Production
	keyFunc        gendata.Keyfun
	luaVM          *lua.LState
	printBuf       *bytes.Buffer
	maxRecursive   int
	cachedStmts    *queue
	debug          bool
}

// HasNext returns whether the iterator exists next sql case
func (i *SQLRandomlyIterator) HasNext() bool {
	return true
}

// Next returns next sql case in iterator
// it will panic when the iterator doesn't exist next sql case
func (i *SQLRandomlyIterator) Next() (string, error) {
	// from cache first
	if !i.cachedStmts.isEmpty() {
		return i.cachedStmts.dequeue(), nil
	}

	stringBuffer := bytes.NewBuffer([]byte{})
	_, err := i.generateSQLRandomly(i.productionName, nil, stringBuffer, false)
	if err != nil {
		return "", err
	}
	if stringBuffer.Len() > 0 {
		i.cachedStmts.enqueue(stringBuffer.String())
	}

	if !i.cachedStmts.isEmpty() {
		return i.cachedStmts.dequeue(), nil
	} else {
		// generate empty string
		return "", nil
	}
}

func getLuaPrintFun(buf *bytes.Buffer) func(*lua.LState) int {
	return func(state *lua.LState) int {
		buf.WriteString(state.ToString(1))
		return 0
	}
}

// GenerateSQLSequentially returns a `SQLSequentialIterator` which can generate sql case by case randomly
// productions is a `Production` array created by `yacc_parser.Parse`
// productionName assigns a production name as the root node.
func GenerateSQLRandomly(headCodeBlocks []*yacc_parser.CodeBlock,
	productions []yacc_parser.Production,
	keyFunc gendata.Keyfun, productionName string, maxRecursive int, debug bool) (SQLIterator, error) {
	pMap := initProductionMap(productions)
	l := lua.NewState()
	// run head code blocks
	for _, codeblock := range headCodeBlocks {
		if err := l.DoString(codeblock.ToString()); err != nil {
			return nil, err
		}
	}

	pBuf := &bytes.Buffer{}
	// cover the origin lua print function
	l.SetGlobal("print", l.NewFunction(getLuaPrintFun(pBuf)))

	return &SQLRandomlyIterator{
		productionName: productionName,
		productionMap:  pMap,
		keyFunc:        keyFunc,
		luaVM:          l,
		printBuf:       pBuf,
		maxRecursive:   maxRecursive,
		cachedStmts:    newQueue(),
		debug:          debug,
	}, nil
}

func (i *SQLRandomlyIterator) printDebugInfo(word string, path []string)  {
	if i.debug {
		log.Printf("word `%s` path: %v\n", word, path)
	}
}

func (i *SQLRandomlyIterator) generateSQLRandomly(productionName string,
	parents []string, writer *bytes.Buffer, parentPreSpace bool) (hasWrite bool, err error) {
	// get root production
	production, exist := i.productionMap[productionName]
	if !exist {
		return false, fmt.Errorf("Production '%s' not found", productionName)
	}
	sameParentNum := 0
	for _, parent := range parents {
		if parent == productionName {
			sameParentNum++
		}
	}
	if sameParentNum >= i.maxRecursive {
		return false, fmt.Errorf("`%s` expression recursive num exceed max loop back %d\n %v",
			productionName, i.maxRecursive, parents)
	}
	parents = append(parents, productionName)
	// random an alter
	selectIndex := rand.Intn(len(production.Alter))
	seqs := production.Alter[selectIndex]
	firstWrite := true
	for index, item := range seqs.Items {
		if yacc_parser.IsTerminal(item) || yacc_parser.NonTerminalNotInMap(i.productionMap, item) {
			// terminal
			i.printDebugInfo(item.ToString(), parents)

			// semicolon
			if item.ToString() == ";" {
				// not last rune in bnf expression
				if selectIndex != len(production.Alter) - 1 || index != len(seqs.Items) - 1 {
					i.cachedStmts.enqueue(writer.String())
					writer.Reset()
					firstWrite = true
					continue
				} else {
					// it is last rune -> just skip
					continue
				}
			}

			if err = handlePreSpace(firstWrite, parentPreSpace, item, writer); err != nil {
				return !firstWrite, err
			}

			if _, err := writer.WriteString(item.ToString()); err != nil {
				return !firstWrite, err
			}

			firstWrite = false

		} else if yacc_parser.IsKeyword(item) {
			if err = handlePreSpace(firstWrite, parentPreSpace, item, writer); err != nil {
				return !firstWrite, err
			}

			// key word parse
			if res, ok := i.keyFunc.Gen(item.ToString()); ok {
				i.printDebugInfo(res, parents)
				_, err := writer.WriteString(res)
				if err != nil {
					return !firstWrite, errors.New("fail to write `io.StringWriter`")
				}

				firstWrite = false
			} else {
				return !firstWrite, fmt.Errorf("'%s' key word not support", item.ToString())
			}
		} else if yacc_parser.IsCodeBlock(item) {
			if err = handlePreSpace(firstWrite, parentPreSpace, item, writer); err != nil {
				return !firstWrite, err
			}

			// lua code block
			if err := i.luaVM.DoString(item.ToString()); err != nil {
				log.Printf("lua code `%s`, run fail\n %v\n",
					item.ToString(), err)
				return !firstWrite, err
			}
			if i.printBuf.Len() > 0 {
				i.printDebugInfo(i.printBuf.String(), parents)
				writer.WriteString(i.printBuf.String())
				i.printBuf.Reset()
				firstWrite = false
			}
		} else {
			// nonTerminal recursive
			var hasSubWrite bool
			if firstWrite {
				hasSubWrite, err = i.generateSQLRandomly(item.ToString(), parents, writer, parentPreSpace)
			} else {
				hasSubWrite, err = i.generateSQLRandomly(item.ToString(), parents, writer, item.HasPreSpace())
			}

			if !firstWrite && hasSubWrite {
				firstWrite = false
			}

			if err != nil {
				return !firstWrite, err
			}
		}
	}

	return !firstWrite, nil
}

func handlePreSpace(firstWrite bool, parentSpace bool, tkn yacc_parser.Token, writer io.StringWriter) error {
	if firstWrite {
		if parentSpace {
			if err := writePreSpace(tkn, writer); err != nil {
				return errors.New("fail to write `io.StringWriter`")
			}
		}
		return nil
	}

	if tkn.HasPreSpace() {
		if err := writePreSpace(tkn, writer); err != nil {
			return errors.New("fail to write `io.StringWriter`")
		}
	}

	return nil
}

func writePreSpace(tkn yacc_parser.Token, writer io.StringWriter) error {
	if _, err := writer.WriteString(" "); err != nil {
		return err
	}

	return nil
}

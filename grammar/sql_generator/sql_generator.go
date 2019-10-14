package sql_generator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pingcap/go-randgen/gendata"
	"github.com/pingcap/go-randgen/grammar/yacc_parser"
	"github.com/yuin/gopher-lua"
	"io"
	"log"
	"math/rand"
)

type BranchAnalyze struct {
	NonTerminal string
	// serial number of this branch
	Branch int
	// confilct number of this branch
	Conflicts int
	// the content expanded by this branch
	Content string
	// one of confict sqls
	ExampleSql string
}

// return false means normal stop visit
type SqlVisitor func(sql string) bool

// visit fixed times to get `times` sqls
func FixedTimesVisitor(f func(i int, sql string), times int) SqlVisitor {
	i := 0
	return func(sql string) bool {
		f(i, sql)
		i++
		if i == times {
			return false
		}
		return true
	}
}

// SQLIterator is a iterator interface of sql generator
// SQLIterator is not thread safe
type SQLIterator interface {
	// Next returns next sql case in iterator
	Visit(visitor SqlVisitor) error

	// push the lastest generated sql into the analyze heap, you should call it in Visit callback
	PushInPathHeap(sql string)

	// return the top n branch in the analyzed sql
	Analyze(n int) ([]*BranchAnalyze, error)
}

// SQLRandomlyIterator is a iterator of sql generator
// note that it is not thread safe
type SQLRandomlyIterator struct {
	productionName string
	productionMap  map[string]*yacc_parser.Production
	keyFunc        gendata.Keyfun
	luaVM          *lua.LState
	printBuf       *bytes.Buffer
	pendingPaths   []*yacc_parser.PendingPath
	maxRecursive   int
	analyze        bool
	debug          bool
}

// if want to push path analyze in the path heap, you should it in Visit callback
func (i *SQLRandomlyIterator) PushInPathHeap(sql string) {
	for _, pending := range i.pendingPaths {
		pending.TheSeq.MaxHeap.Push(pending.Content, sql)
	}
}

func (i *SQLRandomlyIterator) Analyze(n int) ([]*BranchAnalyze, error) {
	if !i.analyze {
		return nil, errors.New("this iterator not support analyze")
	}

	return nil, nil
}

func (i *SQLRandomlyIterator) clearPendingAnalyzes() {
	i.pendingPaths = nil
}

// visitor sqls generted by the iterator
func (i *SQLRandomlyIterator) Visit(visitor SqlVisitor) error {

	wrapper := func(sql string) bool {
		res := visitor(sql)
		i.clearPendingAnalyzes()
		return res
	}

	sqlBuffer := &bytes.Buffer{}

	for {
		_, _, err := i.generateSQLRandomly(i.productionName, newLinkedMap(), sqlBuffer,
			false, wrapper)
		if err != nil && err != normalStop {
			return err
		}

		if err == normalStop || !wrapper(sqlBuffer.String()) {
			return nil
		}

		sqlBuffer.Reset()
	}

	return nil
}

func getLuaPrintFun(buf *bytes.Buffer) func(*lua.LState) int {
	return func(state *lua.LState) int {
		buf.WriteString(state.ToString(1))
		return 0
	}
}

// GenerateSQLSequentially returns a `SQLSequentialIterator` which can generate sql case by case randomly
// productions is a `Production` array created by `yacc_parser.Parse`
// productionName assigns a production name as the root node
// maxRecursive is max bnf extend recursive number in sql generation
// analyze flag is to open root cause analyze feature
// if debug is true, the iterator will print all paths during generation
func GenerateSQLRandomly(headCodeBlocks []*yacc_parser.CodeBlock,
	productionMap map[string]*yacc_parser.Production,
	keyFunc gendata.Keyfun, productionName string, maxRecursive int, analyze bool,
	debug bool) (SQLIterator, error) {
	l := lua.NewState()
	// run head code blocks
	for _, codeblock := range headCodeBlocks {
		if err := l.DoString(codeblock.OriginString()[1:len(codeblock.OriginString())-1]); err != nil {
			return nil, err
		}
	}

	pBuf := &bytes.Buffer{}
	// cover the origin lua print function
	l.SetGlobal("print", l.NewFunction(getLuaPrintFun(pBuf)))

	return &SQLRandomlyIterator{
		productionName: productionName,
		productionMap:  productionMap,
		keyFunc:        keyFunc,
		luaVM:          l,
		printBuf:       pBuf,
		maxRecursive:   maxRecursive,
		analyze:        analyze,
		debug:          debug,
	}, nil
}

var normalStop = errors.New("generateSQLRandomly: normal stop visit")

func (i *SQLRandomlyIterator) printDebugInfo(word string, path *linkedMap) {
	if i.debug {
		log.Printf("word `%s` path: %v\n", word, path.order)
	}
}

func willRecursive(seq *yacc_parser.Seq, set map[string]bool) bool {
	for _, item := range seq.Items {
		if yacc_parser.IsTknNonTerminal(item) && set[item.OriginString()] {
			return true
		}
	}
	return false
}

func (i *SQLRandomlyIterator) generateSQLRandomly(productionName string,
	recurCounter *linkedMap, sqlBuffer *bytes.Buffer,
	parentPreSpace bool, visitor SqlVisitor) (analyzeBuf *bytes.Buffer, hasWrite bool, err error) {
	// get root production
	production, exist := i.productionMap[productionName]
	if !exist {
		return nil, false, fmt.Errorf("Production '%s' not found", productionName)
	}

	// check max recursive count
	recurCounter.enter(productionName)
	defer func() {
		recurCounter.leave(productionName)
	}()
	if recurCounter.m[productionName] > i.maxRecursive {
		return nil, false, fmt.Errorf("`%s` expression recursive num exceed max loop back %d\n %v",
			productionName, i.maxRecursive, recurCounter.order)
	}
	nearMaxRecur := make(map[string]bool)
	for name, count := range recurCounter.m {
		if count == i.maxRecursive {
			nearMaxRecur[name] = true
		}
	}
	selectableSeqs := make([]*yacc_parser.Seq, 0)
	for _, seq := range production.Alter {
		if !willRecursive(seq, nearMaxRecur) {
			selectableSeqs = append(selectableSeqs, seq)
		}
	}
	if len(selectableSeqs) == 0 {
		return nil, false, fmt.Errorf("recursive num exceed max loop back %d\n %v",
			i.maxRecursive, recurCounter.order)
	}

	// random an alter
	selectIndex := rand.Intn(len(selectableSeqs))
	seqs := selectableSeqs[selectIndex]
	firstWrite := true

	// it will record origin producted sql without interpretation
	analyzeBuf = &bytes.Buffer{}

	for index, item := range seqs.Items {
		var subAnalyBuf *bytes.Buffer
		if yacc_parser.IsTerminal(item) || yacc_parser.NonTerminalNotInMap(i.productionMap, item) {
			// terminal
			i.printDebugInfo(item.OriginString(), recurCounter)

			// semicolon
			if item.OriginString() == ";" {
				// not last rune in bnf expression
				if selectIndex != len(production.Alter)-1 || index != len(seqs.Items)-1 {
					if !visitor(sqlBuffer.String()) {
						return nil, !firstWrite, normalStop
					}
					sqlBuffer.Reset()
					firstWrite = true
					continue
				} else {
					// it is last rune -> just skip
					continue
				}
			}

			if err = handlePreSpace(firstWrite, parentPreSpace, item, sqlBuffer); err != nil {
				return nil, !firstWrite, err
			}

			if _, err := sqlBuffer.WriteString(item.OriginString()); err != nil {
				return nil, !firstWrite, err
			}

			firstWrite = false

		} else if yacc_parser.IsKeyword(item) {
			if err = handlePreSpace(firstWrite, parentPreSpace, item, sqlBuffer); err != nil {
				return nil, !firstWrite, err
			}

			// key word parse
			if res, ok, err := i.keyFunc.Gen(item.OriginString()); err != nil {
				return nil, !firstWrite, err
			} else if ok {
				i.printDebugInfo(res, recurCounter)
				_, err := sqlBuffer.WriteString(res)
				if err != nil {
					return nil, !firstWrite, errors.New("fail to write `io.StringWriter`")
				}

				firstWrite = false
			} else {
				return nil, !firstWrite, fmt.Errorf("'%s' key word not support", item.OriginString())
			}
		} else if yacc_parser.IsCodeBlock(item) {
			if err = handlePreSpace(firstWrite, parentPreSpace, item, sqlBuffer); err != nil {
				return nil, !firstWrite, err
			}

			// lua code block
			if err := i.luaVM.DoString(item.OriginString()[1:len(item.OriginString())-1]); err != nil {
				log.Printf("lua code `%s`, run fail\n %v\n",
					item.OriginString(), err)
				return nil, !firstWrite, err
			}
			if i.printBuf.Len() > 0 {
				i.printDebugInfo(i.printBuf.String(), recurCounter)
				sqlBuffer.WriteString(i.printBuf.String())
				i.printBuf.Reset()
				firstWrite = false
			}
		} else {
			// nonTerminal recursive
			var hasSubWrite bool
			if firstWrite {
				subAnalyBuf, hasSubWrite, err = i.generateSQLRandomly(item.OriginString(), recurCounter,
					sqlBuffer, parentPreSpace, visitor)
			} else {
				subAnalyBuf, hasSubWrite, err = i.generateSQLRandomly(item.OriginString(), recurCounter,
					sqlBuffer, item.HasPreSpace(), visitor)
			}

			if firstWrite && hasSubWrite {
				firstWrite = false
			}

			if err != nil {
				return nil, !firstWrite, err
			}
		}

		if subAnalyBuf == nil {
			if item.HasPreSpace() {
				analyzeBuf.WriteRune(' ')
			}
			analyzeBuf.WriteString(item.OriginString())
		} else {
			analyzeBuf.Write(subAnalyBuf.Bytes())
		}
	}

	i.pendingPaths = append(i.pendingPaths, &yacc_parser.PendingPath{
		Content: analyzeBuf.String(),
		TheSeq:  seqs,
	})

	return analyzeBuf, !firstWrite, nil
}

func handlePreSpace(firstWrite bool, parentSpace bool, tkn yacc_parser.Token, writer io.StringWriter) error {
	if firstWrite {
		if parentSpace {
			if err := writePreSpace(writer); err != nil {
				return errors.New("fail to write `io.StringWriter`")
			}
		}
		return nil
	}

	if tkn.HasPreSpace() {
		if err := writePreSpace(writer); err != nil {
			return errors.New("fail to write `io.StringWriter`")
		}
	}

	return nil
}

func writePreSpace(writer io.StringWriter) error {
	if _, err := writer.WriteString(" "); err != nil {
		return err
	}

	return nil
}

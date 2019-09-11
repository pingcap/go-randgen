package sql_generator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yuin/gopher-lua"
	"go-randgen/gendata"
	"go-randgen/grammar/yacc_parser"
	"io"
	"log"
	"math/rand"
)

const maxLoopback = 5
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

type terminator struct {
}

func (t *terminator) walk() bool {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

func (t *terminator) materialize(writer io.StringWriter) error {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

func (t *terminator) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

type expressionNode struct {
	items  []node
	parent *productionNode
}

func (en *expressionNode) materialize(writer io.StringWriter) error {
	for _, item := range en.items {
		err := item.materialize(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *expressionNode) walk() (carry bool) {
	previousCarry := true
	for i := len(en.items) - 1; i >= 0 && previousCarry; i-- {
		previousCarry = en.items[i].walk()
	}
	return previousCarry
}

func (en *expressionNode) existTerminator() bool {
	for _, item := range en.items {
		if _, ok := item.(*terminator); ok {
			return true
		}
	}
	for _, item := range en.items {
		if pNode, ok := item.(*productionNode); ok {
			pNode.pruneTerminator()
		}
	}
	return false
}

func (en *expressionNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	if sameParent >= maxLoopback {
		return true
	}
	return en.parent != nil && en.parent.loopbackDetection(productionName, sameParent)
}

type productionNode struct {
	name      string
	exprs     []*expressionNode
	parent    *expressionNode
	walkIndex int
	pruned    bool
}

func (pn *productionNode) walk() (carry bool) {
	if pn.exprs[pn.walkIndex].walk() {
		pn.walkIndex += 1
		if pn.walkIndex >= len(pn.exprs) {
			pn.walkIndex = 0
			carry = true
		}
	}
	return
}

func (pn *productionNode) materialize(writer io.StringWriter) error {
	return pn.exprs[pn.walkIndex].materialize(writer)
}

func (pn *productionNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	if pn.name == productionName {
		sameParent++
	}
	if sameParent >= maxLoopback {
		return true
	}
	return pn.parent != nil && pn.parent.loopbackDetection(productionName, sameParent)
}

// pruneTerminator remove the branch whose include terminator node.
func (pn *productionNode) pruneTerminator() {
	if pn.pruned {
		return
	}
	var newExprs []*expressionNode
	for _, expr := range pn.exprs {
		if !expr.existTerminator() {
			newExprs = append(newExprs, expr)
		}
	}
	pn.exprs = newExprs
	pn.pruned = true
}

// SQLIterator is a iterator interface of sql generator
type SQLIterator interface {
	// HasNext returns whether the iterator exists next sql case
	HasNext() bool

	// Next returns next sql case in iterator
	Next() (string, error)
}

// SQLSequentialIterator is a iterator of sql generator
type SQLSequentialIterator struct {
	root             *productionNode
	alreadyPointNext bool
	noNext           bool
}

// HasNext returns whether the iterator exists next sql case
func (i *SQLSequentialIterator) HasNext() bool {
	if !i.alreadyPointNext {
		i.noNext = i.root.walk()
		i.alreadyPointNext = true
	}
	return !i.noNext
}

// Next returns next sql case in iterator
// it will panic when the iterator doesn't exist next sql case
func (i *SQLSequentialIterator) Next() string {
	if !i.HasNext() {
		panic("there isn't next item in this sql iterator")
	}
	i.alreadyPointNext = false
	stringBuffer := bytes.NewBuffer([]byte{})
	err := i.root.materialize(stringBuffer)
	if err != nil {
		panic("buffer write failure" + err.Error())
	}
	return stringBuffer.String()
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
}

// HasNext returns whether the iterator exists next sql case
func (i *SQLRandomlyIterator) HasNext() bool {
	return true
}

// Next returns next sql case in iterator
// it will panic when the iterator doesn't exist next sql case
func (i *SQLRandomlyIterator) Next() (string, error) {
	stringBuffer := bytes.NewBuffer([]byte{})
	err := i.generateSQLRandomly(i.productionName, nil, stringBuffer, false)
	if err != nil {
		return "", err
	}
	return stringBuffer.String(), nil
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
func GenerateSQLRandomly(headCodeBlocks []*yacc_parser.CodeBlock, productions []yacc_parser.Production, keyFunc gendata.Keyfun, productionName string) (SQLIterator, error) {
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
	}, nil
}

func (i *SQLRandomlyIterator) generateSQLRandomly(productionName string,
	parents []string, writer io.StringWriter, parentPreSpace bool) (err error) {
	// get root production
	production, exist := i.productionMap[productionName]
	if !exist {
		return fmt.Errorf("Production '%s' not found", productionName)
	}
	sameParentNum := 0
	for _, parent := range parents {
		if parent == productionName {
			sameParentNum++
		}
	}
	if sameParentNum >= maxLoopback {
		return fmt.Errorf("%s recursive num exceed max loop back %d", productionName, maxLoopback)
	}
	parents = append(parents, productionName)
	// random an alter
	selectIndex := rand.Intn(len(production.Alter))
	seqs := production.Alter[selectIndex]
	for index, item := range seqs.Items {
		if yacc_parser.IsTerminal(item) || yacc_parser.NonTerminalNotInMap(i.productionMap, item) {
			// terminal
			if err = handlePreSpace(index, parentPreSpace, item, writer); err != nil {
				return err
			}

			if _, err := writer.WriteString(item.ToString()); err != nil {
				return err
			}
		} else if yacc_parser.IsKeyword(item) {
			if err = handlePreSpace(index, parentPreSpace, item, writer); err != nil {
				return err
			}

			// key word parse
			if res, ok := i.keyFunc.Gen(item.ToString()); ok {
				_, err := writer.WriteString(res)
				if err != nil {
					return errors.New("fail to write `io.StringWriter`")
				}
			} else {
				return fmt.Errorf("'%s' key word not support", item.ToString())
			}
		} else if yacc_parser.IsCodeBlock(item) {
			if err = handlePreSpace(index, parentPreSpace, item, writer); err != nil {
				return err
			}

			// lua code block
			if err := i.luaVM.DoString(item.ToString()); err != nil {
				log.Printf("lua code `%s`, run fail\n %v\n",
					item.ToString(), err)
				return err
			}
			if i.printBuf.Len() > 0 {
				writer.WriteString(i.printBuf.String())
				i.printBuf.Reset()
			}
		} else {
			// nonTerminal recursive
			if index == 0 {
				err = i.generateSQLRandomly(item.ToString(), parents, writer, parentPreSpace)
			} else {
				err = i.generateSQLRandomly(item.ToString(), parents, writer, item.HasPreSpace())
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handlePreSpace(index int, parentSpace bool, tkn yacc_parser.Token, writer io.StringWriter) error {
	if index == 0 {
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

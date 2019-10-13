package yacc_parser

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"
)

// token sequence of one branch
type Seq struct {
	Items []Token
	MaxHeap  Heap
}

func NewSeq(items []Token) *Seq {
	return &Seq{Items:items}
}

func (s Seq) String() string {
	buf := &bytes.Buffer{}
	for i, tkn := range s.Items {
		if i == 0 {
			buf.WriteString(tkn.ToString())
			continue
		}

		if tkn.HasPreSpace() {
			buf.WriteRune(' ')
		}
		if IsCodeBlock(tkn) {
			buf.WriteString("{" + tkn.ToString() + "}")
		} else {
			buf.WriteString(tkn.ToString())
		}
	}

	return buf.String()
}

// one bnf expression
type Production struct {
	// serial Number of this production
	Number int
	// left value of bnf expression
	Head Token
	// right expression of bnf expression,
	// every Seq represents a branch of this expression
	Alter []*Seq
}

const (
	initState            = 0
	delimFetchedState    = 1
	termFetchedState     = 2
	prepareNextProdState = 3
	endState             = 4
)

func skipComment(nextToken func() (Token, error)) (t Token, err error)  {
	for {
		t, err = nextToken()
		if err != nil {
			return nil, err
		}

		if !isComment(t) {
			return t, nil
		}
	}
}

func collectHeadCodeBlocks(nextToken func() (Token, error)) (t Token, cbs []*CodeBlock, err error)  {
	cbs = make([]*CodeBlock, 0)
	for {
		t, err = skipComment(nextToken)
		if err != nil {
			return nil, nil, err
		}

		if cb, ok := t.(*CodeBlock); ok {
			cbs = append(cbs, cb)
		} else {
			break
		}
	}

	return t, cbs, nil
}

func Parse(nextToken func() (Token, error)) ([]*CodeBlock, []*Production, error) {
	var tkn Token
	var prods []*Production
	p := &Production{}
	// production serial Number
	pNumber := 0
	s := NewSeq(nil)
	var lastTerm Token

	state := initState
	t, codeblocks, err := collectHeadCodeBlocks(nextToken)
	if err != nil {
		return nil, nil, err
	}
	if !IsTknNonTerminal(t) {
		return nil, nil, fmt.Errorf("%s is not nonterminal", t.ToString())
	}

	p.Head = t
	p.Number = pNumber
	pNumber++

	//
	// initState -> delimFetchedState -> termFetchedState ->...
	//
	for state != endState {
		tkn, err = skipComment(nextToken)
		if err != nil {
			return nil, nil, err
		}
		switch state {
		case initState:
			if tkn.ToString() != ":" {
				return nil, nil, errors.New("expect ':'")
			}
			state = delimFetchedState
		case delimFetchedState:
			if isEOF(tkn) {
				s.Items = append(s.Items, &terminal{val:""})
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
				continue
			}
			if tkn.ToString() == "|" || isEOF(tkn) {
				// multi delimiter will have empty alter
				s.Items = append(s.Items, &terminal{val:""})
				p.Alter = append(p.Alter, s)
				s = NewSeq(nil)
			} else if tkn.ToString() == ":" {
				continue
			} else {
				state = termFetchedState
				s.Items = append(s.Items, tkn)
			}
			// state after first term fetched
		case termFetchedState:
			switch v := tkn.(type) {
			case *eof:
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
			case *operator:
				if v.ToString() == "|" {
					p.Alter = append(p.Alter, s)
					s = NewSeq(nil)
				}
				if v.ToString() == ":" {
					p.Alter = append(p.Alter, NewSeq([]Token{&terminal{val:""}}))
					prods = append(prods, p)
					p = &Production{Head:s.Items[0], Number:pNumber}
					pNumber++
					s = NewSeq(nil)
				}
				state = delimFetchedState
			case *nonTerminal, *keyword, *terminal, *CodeBlock:
				// record last term
				lastTerm = v
				state = prepareNextProdState
			}
			// if one branch has many Token, it will always in this state
		case prepareNextProdState:
			switch v := tkn.(type) {
			case *eof:
				s.Items = append(s.Items, lastTerm)
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
			case *operator:
				if v.val == "|" {
					s.Items = append(s.Items, lastTerm)
					p.Alter = append(p.Alter, s)
					s = NewSeq(nil)
				} else if v.val == ":" {
					// enter next bnf expression
					p.Alter = append(p.Alter, s)
					s = NewSeq(nil)
					prods = append(prods, p)
					if !IsTknNonTerminal(lastTerm) {
						return nil, nil, fmt.Errorf("%s is not nonterminal \n %s",
							lastTerm.ToString(), debug.Stack())
					}
					p = &Production{Head: lastTerm, Number:pNumber}
					pNumber++
				}
				state = delimFetchedState
			case *nonTerminal, *keyword, *terminal, *CodeBlock:
				// push last tern in Seq
				s.Items = append(s.Items, lastTerm)
				lastTerm = v
			}
		}
	}
	return codeblocks, prods, nil
}

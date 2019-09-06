package yacc_parser

import (
	"errors"
	"fmt"
)

// 代表一个分支中的token序列
type Seq struct {
	Items []token
}

// 代表一个bnf表达式
type Production struct {
	// bnf表达式的左边
	Head  token
	// bnf表达式的右边 每个Seq代表一个可行的分支
	Alter []Seq
}

type stateType int

const (
	initState            = 0
	delimFetchedState    = 1
	termFetchedState     = 2
	prepareNextProdState = 3
	endState             = 4
)

func skipComment(nextToken func() (token, error)) (t token, err error)  {
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

func Parse(nextToken func() (token, error)) ([]Production, error) {
	var tkn token
	var prods []Production
	var p Production
	var s Seq
	var lastTerm token

	state := initState
	t, err := skipComment(nextToken)
	if err != nil {
		return nil, err
	}
	if isTknNonTerminal(t) {
		return nil, fmt.Errorf("%s is not nonterminal", t.toString())
	}

	p.Head = t

	//
	// initState -> delimFetchedState -> termFetchedState ->...
	//
	for state != endState {
		tkn, err = skipComment(nextToken)
		if err != nil {
			return nil, err
		}
		switch state {
		case initState:
			if tkn.toString() != ":" {
				return nil, errors.New("expect ':'")
			}
			state = delimFetchedState
		case delimFetchedState:
			_, isNt := tkn.(*nonTerminal)
			_, isT := tkn.(*terminal)
			_, isKw := tkn.(*keyword)
			if !isNt && !isT && !isKw {
				return nil, fmt.Errorf("%s is not nonterminal, terminal or keyword")
			}
			state = termFetchedState
			// token of one branch
			s.Items = append(s.Items, tkn)
		case termFetchedState:
			switch v := tkn.(type) {
			case *eof:
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
			case *operator:
				p.Alter = append(p.Alter, s)
				s = Seq{}
				state = termFetchedState
			case *nonTerminal, *keyword, *terminal:
				// record last term
				lastTerm = v
				state = prepareNextProdState
			}
			// if one branch has many token, it will always in this state
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
					s = Seq{}
				} else if v.val == ":" {
					// enter next bnf expression
					p.Alter = append(p.Alter, s)
					s = Seq{}
					prods = append(prods, p)
					p = Production{Head: lastTerm}
				}
				state = delimFetchedState
			case *nonTerminal, *keyword, *terminal:
				// push last tern in Seq
				s.Items = append(s.Items, lastTerm)
				lastTerm = v
			}
		}
	}
	return prods, nil
}

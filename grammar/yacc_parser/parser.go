package yacc_parser

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// 代表一个分支中的token序列
type Seq struct {
	Items []Token
}

// one bnf expression
type Production struct {
	// left value of bnf expression
	Head Token
	// right expression of bnf expression,
	// every Seq represents a branch of this expression
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

func Parse(nextToken func() (Token, error)) ([]Production, error) {
	var tkn Token
	var prods []Production
	var p Production
	var s Seq
	var lastTerm Token

	state := initState
	t, err := skipComment(nextToken)
	if err != nil {
		return nil, err
	}
	if !IsTknNonTerminal(t) {
		return nil, fmt.Errorf("%s is not nonterminal", t.ToString())
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
			if tkn.ToString() != ":" {
				return nil, errors.New("expect ':'")
			}
			state = delimFetchedState
		case delimFetchedState:
			if isOperator(tkn) {
				continue
			}
			state = termFetchedState
			// Token of one branch
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
					s = Seq{}
				} else if v.val == ":" {
					// enter next bnf expression
					p.Alter = append(p.Alter, s)
					s = Seq{}
					prods = append(prods, p)
					if !IsTknNonTerminal(lastTerm) {
						return nil, fmt.Errorf("%s is not nonterminal \n %s",
							lastTerm.ToString(), debug.Stack())
					}
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

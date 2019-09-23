package yacc_parser

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"github.com/emirpasic/gods/stacks/arraystack"
)

type Token interface {
	ToString() string
	HasPreSpace() bool
}

type commonAttr struct {
	hasPreSpace bool
}

type eof struct{}

func (*eof) HasPreSpace() bool {
	return false
}

func (*eof) ToString() string {
	return "EOF"
}

// ':' or '|'
type operator struct {
	val string
}

func (op *operator) HasPreSpace() bool {
	return false
}

func (op *operator) ToString() string {
	return op.val
}

type keyword struct {
	commonAttr
	val string
}

func (kw *keyword) HasPreSpace() bool {
	return kw.commonAttr.hasPreSpace
}

func (kw *keyword) ToString() string {
	return kw.val
}

type nonTerminal struct {
	commonAttr
	val string
}

func (nt *nonTerminal) HasPreSpace() bool {
	return nt.commonAttr.hasPreSpace
}

func (nt *nonTerminal) ToString() string {
	return nt.val
}

type terminal struct {
	commonAttr
	val string
}

func (t *terminal) HasPreSpace() bool {
	return t.commonAttr.hasPreSpace
}

func (t *terminal) ToString() string {
	return t.val
}

type comment struct {
	val string
}

func (c *comment) HasPreSpace() bool {
	return false
}

func (c *comment) ToString() string {
	return c.val
}

type CodeBlock struct {
	commonAttr
	val string
}

func (c *CodeBlock) HasPreSpace() bool {
	return c.commonAttr.hasPreSpace
}

func (c *CodeBlock) ToString() string {
	return c.val
}

const (
	inSingQuoteStr = iota + 1
	inDoubleQuoteStr
	inOneLineComment
	inComment
)

func getByQuote(r rune) int {
	if r == '"' {
		return inDoubleQuoteStr
	}
	return inSingQuoteStr
}

type quote struct {
	c int
}

func (q *quote) isInsideStr() bool {
	return q.c == inSingQuoteStr || q.c == inDoubleQuoteStr
}

func (q *quote) isInComment() bool {
	return q.c == inOneLineComment || q.c == inComment
}

func (q *quote) isInOneLineComment() bool {
	return q.c == inOneLineComment
}

func (q *quote) isInSome() bool {
	return q.c != 0
}

func (q *quote) tryToggle(other int) bool {
	if q.c == 0 {
		q.c = other
		return true
	} else if q.c == other {
		q.c = 0
		return true
	}
	return false
}

func skipSpace(reader *RuneSeq) (hasSpace bool, r rune, err error) {
	for {
		r, err = reader.ReadRune()
		if err != nil {
			return false, 0, err
		}

		if !unicode.IsSpace(r) {
			return hasSpace, r, nil
		} else {
			hasSpace = true
		}
	}
}

func unreadTwice(reader *RuneSeq) error {
	reader.UnreadRune()
	reader.UnreadRune()
	return nil
}

type RuneSeq struct {
	Runes []rune
	Pos int
}

func (r *RuneSeq) ReadRune() (rune, error) {
	if r.Pos >= len(r.Runes) {
		return 0, io.EOF
	}

	cur := r.Runes[r.Pos]
	r.Pos++
	return cur, nil
}

func (r *RuneSeq) UnreadRune() {
	r.Pos--
}

// see if next rune equals expect  without read
func (r *RuneSeq) PeekEqual(expect rune) bool {
	if r.Pos >= len(r.Runes) {
		return false
	}

	return r.Runes[r.Pos] == expect
}

// Tokenize is used to wrap a reader into a Token producer.
// simple lexer not look back, have some problem when quote not pair
// runeScanner must support unread twice
func Tokenize(reader *RuneSeq) func() (Token, error) {
	q := quote{0}
	pStack := arraystack.New()
	return func() (Token, error) {
		var r rune
		var err error
		// Skip spaces.
		hasSpace, r, err := skipSpace(reader)
		if err == io.EOF {
			return &eof{}, nil
		} else if err != nil {
			return nil, err
		}

		common := commonAttr{hasPreSpace:hasSpace}

		// Handle delimiter.
		if (r == ':' && !reader.PeekEqual('=')) || r == '|' {
			return &operator{string(r)}, nil
		}

		// Toggle isInsideStr.
		if r == '\'' || r == '"' {
			q.tryToggle(getByQuote(r))
		}

		// handle one line comment
		if r == '#' {
			q.tryToggle(inOneLineComment)
		}

		// handle code block
		if r == '{' {
			pStack.Push(r)
		}

		// handle special rune
		if isSpecialRune(r) {
			return &terminal{common,string(r)}, nil
		}

		// Handle stringLiteral or identifier
		var last rune
		var stringBuf string
		stringBuf = string(r)

		for {
			last = r
			r, err = reader.ReadRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			// in code block
			if !pStack.Empty() {
				stringBuf += string(r)
				if r == '{' {
					pStack.Push(r)
				} else if r == '}' {
					pStack.Pop()
					if pStack.Empty() {
						break
					}
				}
				continue
			}

			// enter comment
			if !q.isInComment() {
				if last == '/' && r == '*' {
					q.tryToggle(inComment)
				}
			}

			if (unicode.IsSpace(r) || r == '|' || isSpecialRune(r) ||
				(r == ':' && !reader.PeekEqual('='))) && !q.isInSome() {
				reader.UnreadRune()
				break
			}

			stringBuf += string(r)
			if !q.isInComment() {
				// Handle end str.
				if r == '\'' || r == '"' {
					// identifier can not have ' or "
					if !q.isInsideStr() {
						return nil, fmt.Errorf("unexpected character: `%s` in `%s`", string(r), stringBuf)
					}
					if q.tryToggle(getByQuote(r)) {
						break
					}
				}
			} else {
				// in comment
				if r == '\n' && q.isInOneLineComment() {
					q.tryToggle(inOneLineComment)
					return &comment{stringBuf}, nil
				}
				if last == '*' && r == '/' && q.isInComment() {
					q.tryToggle(inComment)
					return &comment{stringBuf}, nil
				}
			}
		}

		// CodeBlock
		if strings.HasPrefix(stringBuf, "{") {
			return &CodeBlock{common, stringBuf[1: len(stringBuf) - 1]}, nil
		}

		// stringLiteral
		if strings.HasPrefix(stringBuf, "'") || strings.HasPrefix(stringBuf, "\"") {
			return &terminal{common, stringBuf}, nil
		}

		// keyword
		if strings.HasPrefix(stringBuf, "_") {
			return &keyword{common, stringBuf}, nil
		}

		// nonTerminal
		if isNonTerminal(stringBuf) {
			return &nonTerminal{common, stringBuf}, nil
		}

		// terminal
		return &terminal{common, stringBuf}, nil
	}
}

var specialRune = map[rune]bool{
	',': true,
	';': true,
	'(': true,
	')': true,
}

func isSpecialRune(r rune) bool {
	_, ok := specialRune[r]
	return ok
}

func isNonTerminal(token string) bool {
	allDigit := true
	for _, c := range token {
		if !unicode.IsLower(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
		if !unicode.IsDigit(c) {
			allDigit = false
		}
	}
	return !allDigit
}

func isEOF(tkn Token) bool {
	_, ok := tkn.(*eof)
	return ok
}

func isComment(tkn Token) bool {
	_, ok := tkn.(*comment)
	return ok
}

func IsTknNonTerminal(tkn Token) bool {
	_, ok := tkn.(*nonTerminal)
	return ok
}

func IsTerminal(tkn Token) bool {
	_, ok := tkn.(*terminal)
	return ok
}

func IsKeyword(tkn Token) bool {
	_, ok := tkn.(*keyword)
	return ok
}

func IsCodeBlock(tkn Token) bool {
	_, ok := tkn.(*CodeBlock)
	return ok
}

func NonTerminalNotInMap(pmap map[string]Production, tkn Token) bool  {
	non, ok := tkn.(*nonTerminal)
	if !ok {
		return false
	}

	_, ok = pmap[non.ToString()]
	return !ok
}

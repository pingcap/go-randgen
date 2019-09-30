package yacc_parser

import (
	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/pkg/errors"
	"io"
	"unicode"
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
	inSingQuoteStr = iota
	inDoubleQuoteStr
	inOneLineComment
	inComment
	inCodeBlock
	inCodeBlockStr
	inKeyWord
	inNonTerminal
	inTerminal
)

var stateMap = map[rune]int{
	'\'': inSingQuoteStr,
	'"': inDoubleQuoteStr,
	'#': inOneLineComment,
	'{': inCodeBlock,
	'_': inKeyWord,
}

func runeInitState(r rune) int {
	if s, ok := stateMap[r]; ok {
		return s
	}

	if unicode.IsLower(r) {
		return inNonTerminal
	}

	return inTerminal
}


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

type RuneSeq struct {
	Runes []rune
	// next position to read
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

func (r *RuneSeq) SetPos(newPos int)  {
	r.Pos = newPos
}

// see if next rune equals expect  without read
func (r *RuneSeq) PeekEqual(expect rune) bool {
	if r.Pos >= len(r.Runes) {
		return false
	}

	return r.Runes[r.Pos] == expect
}

// see if last rune equals expect
func (r *RuneSeq) LastEqual(exepect rune) bool {
	if r.Pos <= 1 {
		return false
	}

	return r.Runes[r.Pos-2] == exepect
}

func (r *RuneSeq) Slice(from int) string {
	return string(r.Runes[from:r.Pos])
}

func tknEnd(reader *RuneSeq, r rune) bool {
	return unicode.IsSpace(r) || r == '|' || isSpecialRune(r) || r == '#' || r == '{' ||
		(r == ':' && !reader.PeekEqual('=')) || (r == '/' && reader.PeekEqual('*'))
}

// Tokenize is used to wrap a reader into a Token producer.
// simple lexer not look back, have some problem when quote not pair
// runeScanner must support unread twice
func Tokenize(reader *RuneSeq) func() (Token, error) {
	stack := arraystack.New()
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

		// handle special rune
		if isSpecialRune(r) {
			return &terminal{common,string(r)}, nil
		}

		state := runeInitState(r)
		if state == inCodeBlock {
			stack.Push('{')
		}

		initPos := reader.Pos - 1

		for {
			r, err = reader.ReadRune()
			if err != nil && err != io.EOF {
				return nil, err
			}

			// all state must handle io.EOF first
			switch state {
			case inNonTerminal:
				if err == io.EOF {
					return &nonTerminal{common, reader.Slice(initPos)}, nil
				}
				if tknEnd(reader, r) {
					reader.UnreadRune()
					return &nonTerminal{common, reader.Slice(initPos)}, nil
				}
				// nonTerminal can only be composed of lower word, digit or '_'
				if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
					state = inTerminal
				}

			case inTerminal:
				if err == io.EOF {
					return &terminal{common, reader.Slice(initPos)}, nil
				}
				if tknEnd(reader, r) {
					reader.UnreadRune()
					return &terminal{common, reader.Slice(initPos)}, nil
				}

				if reader.LastEqual('/') && r == '*' {
					state = inComment
				}
			case inKeyWord:
				if err == io.EOF || tknEnd(reader, r) {
					if err != io.EOF {
						reader.UnreadRune()
					}
					keywordLiteral := reader.Slice(initPos)
					if keywordLiteral == "_" {
						return &terminal{common, keywordLiteral}, nil
					}
					return &keyword{common, keywordLiteral}, nil
				}
			case inOneLineComment:
				if err == io.EOF || r == '\n' {
					return &comment{reader.Slice(initPos)}, nil
				}
			case inComment:
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(initPos)
					continue
				}
				if reader.LastEqual('*') && r == '/' {
					return &comment{reader.Slice(initPos)}, nil
				}
			case inSingQuoteStr:
				// look back
				if err == io.EOF || r == '\n' {
					state = inTerminal
					reader.SetPos(initPos)
					continue
				}
				if r == '\'' {
					return &terminal{common, reader.Slice(initPos)}, nil
				}

			case inDoubleQuoteStr:
				// look back
				if err == io.EOF || r == '\n' {
					state = inTerminal
					reader.SetPos(initPos)
					continue
				}
				if r == '"' {
					return &terminal{common, reader.Slice(initPos)}, nil
				}
			case inCodeBlock:
				// look back
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(initPos)
					stack.Clear()
					continue
				}

				if r == '{' {
					stack.Push(r)
				} else if r == '}' {
					stack.Pop()
					if stack.Empty() {
						return &CodeBlock{common,
							string(reader.Runes[initPos+1:reader.Pos-1])}, nil
					}
				} else if r == '\'' || r== '"' {
					stack.Push(r)
					state = inCodeBlockStr
			    }
			case inCodeBlockStr:
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(initPos)
					stack.Clear()
					continue
				}

				if p, ok := stack.Peek(); ok {
					if r == p.(rune) && !reader.LastEqual('\\') {
						stack.Pop()
						state = inCodeBlock
					}
				} else {
					return nil, errors.New("impossible code path")
				}
			}
		}
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

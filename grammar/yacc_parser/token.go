package yacc_parser

import (
	"io"
	"unicode"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/pkg/errors"
)

type Token interface {
	OriginString() string
	HasPreSpace() bool
}

type commonAttr struct {
	hasPreSpace bool
}

type eof struct{}

func (*eof) HasPreSpace() bool {
	return false
}

func (*eof) OriginString() string {
	return "EOF"
}

// ':' or '|'
type operator struct {
	val string
}

func (op *operator) HasPreSpace() bool {
	return false
}

func (op *operator) OriginString() string {
	return op.val
}

type keyword struct {
	commonAttr
	val string
}

func (kw *keyword) HasPreSpace() bool {
	return kw.commonAttr.hasPreSpace
}

func (kw *keyword) OriginString() string {
	return kw.val
}

type nonTerminal struct {
	commonAttr
	val string
}

func (nt *nonTerminal) HasPreSpace() bool {
	return nt.commonAttr.hasPreSpace
}

func (nt *nonTerminal) OriginString() string {
	return nt.val
}

type terminal struct {
	commonAttr
	val string
}

func (t *terminal) HasPreSpace() bool {
	return t.commonAttr.hasPreSpace
}

func (t *terminal) OriginString() string {
	return t.val
}

type comment struct {
	val string
}

func (c *comment) HasPreSpace() bool {
	return false
}

func (c *comment) OriginString() string {
	return c.val
}

type CodeBlock struct {
	commonAttr
	val string
}

func (c *CodeBlock) HasPreSpace() bool {
	return c.commonAttr.hasPreSpace
}

func (c *CodeBlock) OriginString() string {
	return c.val
}

const (
	tknInit = iota
	inSingQuoteStr
	inDoubleQuoteStr
	inOneLineComment
	inComment
	inCodeBlock
	inCodeBlockStr               // in lua string
	inCodeBlockSingleLineComment // in lua single line comment
	prepareCodeBlockMultiLineComment
	inCodeBlockMultiLineComment // in lua multiline comment
	endCodeBlockMultiLineComment
	inKeyWord
	inNonTerminal
	inTerminal
)

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

func (r *RuneSeq) SetPos(newPos int) {
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

var stateMap = map[rune]int{
	'\'': inSingQuoteStr,
	'"':  inDoubleQuoteStr,
	'#':  inOneLineComment,
	'{':  inCodeBlock,
	'_':  inKeyWord,
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

// Tokenize is used to wrap a reader into a Token producer.
func Tokenize(reader *RuneSeq) func() (Token, error) {
	stack := arraystack.New()
	return func() (Token, error) {

		common := commonAttr{hasPreSpace: false}

		state := tknInit

		var lookBackPos int
		// variable to analyze lua multiline comments
		var luaCommentDepth int
		var endLuaCommentDepCounter int

		lastState := state

		// state machine
		for {
			r, err := reader.ReadRune()
			if err != nil && err != io.EOF {
				return nil, err
			}

			stateCopy := state

			// all state must handle io.EOF first
			switch state {
			case tknInit:
				if err == io.EOF {
					return &eof{}, nil
				}

				// skip space
				if unicode.IsSpace(r) {
					common.hasPreSpace = true
					continue
				}

				// Handle delimiter.
				// prevent `:` operator to be conflict with sql assign oprator `:=`
				if (r == ':' && !reader.PeekEqual('=')) || r == '|' {
					return &operator{string(r)}, nil
				}

				// handle special rune
				if isSpecialRune(r) {
					return &terminal{common, string(r)}, nil
				}

				// transfer state
				lookBackPos = reader.Pos - 1
				switch r {
				case '\'':
					state = inSingQuoteStr
				case '"':
					state = inDoubleQuoteStr
				case '#':
					state = inOneLineComment
				case '_':
					state = inKeyWord
				case '{':
					state = inCodeBlock
					stack.Push('{')
				default:
					if unicode.IsLower(r) {
						state = inNonTerminal
					} else {
						state = inTerminal
					}
				}

			case inNonTerminal:
				if err == io.EOF {
					return &nonTerminal{common, reader.Slice(lookBackPos)}, nil
				}
				if tknEnd(reader, r) {
					reader.UnreadRune()
					return &nonTerminal{common, reader.Slice(lookBackPos)}, nil
				}
				// nonTerminal can only be composed of lower word, digit or '_'
				if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
					state = inTerminal
				}

			case inTerminal:
				if err == io.EOF {
					return &terminal{common, reader.Slice(lookBackPos)}, nil
				}
				if tknEnd(reader, r) {
					reader.UnreadRune()
					return &terminal{common, reader.Slice(lookBackPos)}, nil
				}

				if lastState == tknInit && reader.LastEqual('/') && r == '*' {
					state = inComment
				}
			case inKeyWord:
				if err == io.EOF || tknEnd(reader, r) {
					if err != io.EOF {
						reader.UnreadRune()
					}
					keywordLiteral := reader.Slice(lookBackPos)
					if keywordLiteral == "_" {
						return &terminal{common, keywordLiteral}, nil
					}
					return &keyword{common, keywordLiteral}, nil
				}
			case inOneLineComment:
				if err == io.EOF || r == '\n' {
					return &comment{reader.Slice(lookBackPos)}, nil
				}
			case inComment:
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(lookBackPos + 1)
					continue
				}
				if reader.LastEqual('*') && r == '/' {
					return &comment{reader.Slice(lookBackPos)}, nil
				}
			case inSingQuoteStr:
				// look back
				if err == io.EOF || r == '\n' {
					state = inTerminal
					reader.SetPos(lookBackPos + 1)
					continue
				}
				if r == '\'' {
					return &terminal{common, reader.Slice(lookBackPos)}, nil
				}

			case inDoubleQuoteStr:
				// look back
				if err == io.EOF || r == '\n' {
					state = inTerminal
					reader.SetPos(lookBackPos + 1)
					continue
				}
				if r == '"' {
					return &terminal{common, reader.Slice(lookBackPos)}, nil
				}

			//code block related states
			case inCodeBlock:
				// look back
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(lookBackPos + 1)
					stack.Clear()
					continue
				}

				if r == '{' {
					stack.Push(r)
				} else if r == '}' {
					stack.Pop()
					if stack.Empty() {
						return &CodeBlock{common,
							string(reader.Slice(lookBackPos))}, nil
					}
				} else if r == '\'' || r == '"' {
					stack.Push(r)
					state = inCodeBlockStr
				} else if r == '-' && reader.LastEqual('-') {
					state = inCodeBlockSingleLineComment
				}

			case inCodeBlockStr:
				if err == io.EOF {
					state = inTerminal
					reader.SetPos(lookBackPos + 1)
					stack.Clear()
					continue
				}

				if p, ok := stack.Peek(); ok {
					// in consider of escaped string
					if r == p.(rune) && !reader.LastEqual('\\') {
						stack.Pop()
						state = inCodeBlock
					}
				} else {
					return nil, errors.New("impossible code path")
				}

			case inCodeBlockSingleLineComment:
				if err == io.EOF || r == '\n' {
					state = inCodeBlock
					continue
				}

				if lastState == inCodeBlock && r == '[' {
					luaCommentDepth = 0
					state = prepareCodeBlockMultiLineComment
				}

			case prepareCodeBlockMultiLineComment:
				if err == io.EOF {
					state = inCodeBlockSingleLineComment
					continue
				}

				if r == '[' {
					state = inCodeBlockMultiLineComment
				} else if r == '=' {
					luaCommentDepth++
				} else {
					state = inCodeBlockSingleLineComment
				}
			case inCodeBlockMultiLineComment:
				if err == io.EOF {
					return nil, errors.New("error at EOF, invalid multiline comment")
				}

				if r == ']' {
					endLuaCommentDepCounter = 0
					state = endCodeBlockMultiLineComment
				}
			case endCodeBlockMultiLineComment:
				if err == io.EOF {
					return nil, errors.New("error at EOF, invalid multiline comment")
				}

				if r == '=' {
					endLuaCommentDepCounter++
				} else if r == ']' && endLuaCommentDepCounter == luaCommentDepth {
					state = inCodeBlock
				} else {
					state = inCodeBlockMultiLineComment
				}
			}
			lastState = stateCopy
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

func NonTerminalNotInMap(pmap map[string]*Production, tkn Token) bool {
	non, ok := tkn.(*nonTerminal)
	if !ok {
		return false
	}

	_, ok = pmap[non.OriginString()]
	return !ok
}

func NonTerminalInMap(pmap map[string]*Production, tkn Token) bool {
	non, ok := tkn.(*nonTerminal)
	if !ok {
		return false
	}

	_, ok = pmap[non.OriginString()]
	return ok
}

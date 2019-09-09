package yacc_parser

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Token interface {
	ToString() string
}
type eof struct{}

func (*eof) ToString() string {
	return "EOF"
}

// ':' or '|'
type operator struct {
	val string
}

func (op *operator) ToString() string {
	return op.val
}

type keyword struct {
	val string
}

func (kw *keyword) ToString() string {
	return kw.val
}

type nonTerminal struct {
	val string
}

func (nt *nonTerminal) ToString() string {
	return nt.val
}

type terminal struct {
	val string
}

func (t *terminal) ToString() string {
	return t.val
}

type comment struct {
	val string
}

func (c *comment) ToString() string {
	return c.val
}

type codeBlock struct {
	val string
}

func (c *codeBlock) ToString() string {
	return c.val
}

const (
	inSingQuoteStr = iota + 1
	inDoubleQuoteStr
	inOneLineComment
	inComment
	inCodeBlock
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

func (q *quote) isInCodeBlock() bool {
	return q.c == inCodeBlock
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

func skipSpace(reader io.RuneScanner) (r rune, err error) {
	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			return 0, err
		}

		if !unicode.IsSpace(r) {
			return r, nil
		}
	}
}

// Tokenize is used to wrap a reader into a Token producer.
// simple lexer not look back, have some problem when quote not pair
func Tokenize(reader io.RuneScanner) func() (Token, error) {
	q := quote{0}
	return func() (Token, error) {
		var r rune
		var err error
		// Skip spaces.
		r, err = skipSpace(reader)
		if err == io.EOF {
			return &eof{}, nil
		} else if err != nil {
			return nil, err
		}

		// Handle delimiter.
		if r == ':' || r == '|' {
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
			q.tryToggle(inCodeBlock)
		}

		// handle special rune
		if isSpecialRune(r) {
			return &terminal{string(r)}, nil
		}

		// Handle stringLiteral or identifier
		var last rune
		var stringBuf string
		stringBuf = string(r)

		for {
			last = r
			r, _, err = reader.ReadRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			if q.isInCodeBlock(){
				stringBuf += string(r)
				if r == '}' {
					q.tryToggle(inCodeBlock)
					break
				}
				continue
			}

			// enter comment
			if !q.isInComment() {
				if last == '/' && r == '*' {
					q.tryToggle(inComment)
				}
			}

			if (unicode.IsSpace(r) || isDelimiter(r) || isSpecialRune(r)) && !q.isInSome() {
				if err := reader.UnreadRune(); err != nil {
					panic(fmt.Sprintf("Unable to unread rune: %s.", string(r)))
				}
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

		// codeBlock
		if strings.HasPrefix(stringBuf, "{") {
			return &codeBlock{stringBuf[1: len(stringBuf) - 1]}, nil
		}

		// stringLiteral
		if strings.HasPrefix(stringBuf, "'") || strings.HasPrefix(stringBuf, "\"") {
			return &terminal{stringBuf}, nil
		}

		// keyword
		if strings.HasPrefix(stringBuf, "_") {
			return &keyword{stringBuf}, nil
		}

		// nonTerminal
		if isNonTerminal(stringBuf) {
			return &nonTerminal{stringBuf}, nil
		}

		// terminal
		return &terminal{stringBuf}, nil
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

func isDelimiter(r rune) bool {
	return r == '|' || r == ':'
}

func isNonTerminal(token string) bool {
	for _, c := range token {
		if !unicode.IsLower(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}
	return true
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
	_, ok := tkn.(*codeBlock)
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

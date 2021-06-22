package yacc_parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	next := Tokenize(&RuneSeq{Runes: []rune(`sql_statement: simple_statement_or_begin EMPTY ';' 
          opt_end_of_input
		|simple_statement_or_begin END_OF_INPUT;
        |

        ;

		opt_end_of_input:
                | {a = 1} empty {print(a+1);arr={1,2,3,4}}
                | END_OF_INPUT _table

        tail_empty:
           SELECT @A := 1
           |
`)})
	_, productions, err := Parse(next)
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, len(productions))

	assertProduct(t, [][]string{
		{"*yacc_parser.nonTerminal"}, // Head
		{"*yacc_parser.nonTerminal", "*yacc_parser.terminal",
			"*yacc_parser.terminal", "*yacc_parser.nonTerminal"}, //Seq 0
		{"*yacc_parser.nonTerminal", "*yacc_parser.terminal", "*yacc_parser.terminal"}, //Seq1
		{"*yacc_parser.terminal"}, // Seq2
	}, productions[0], 0)

	assertProduct(t, [][]string{
		{"*yacc_parser.nonTerminal"},
		{"*yacc_parser.terminal"},
		{"*yacc_parser.CodeBlock", "*yacc_parser.nonTerminal", "*yacc_parser.CodeBlock"},
		{"*yacc_parser.terminal", "*yacc_parser.keyword"},
	}, productions[1], 1)

	assertProduct(t, [][]string{
		{"*yacc_parser.nonTerminal"},
		{"*yacc_parser.terminal", "*yacc_parser.terminal",
			"*yacc_parser.terminal", "*yacc_parser.terminal"},
		{"*yacc_parser.terminal"},
	}, productions[2], 2)
}

func tokenType(tkn Token) string {
	return fmt.Sprintf("%T", tkn)
}

func assertProduct(t *testing.T, expect [][]string, real *Production, expectNum int) {
	assert.Equal(t, expect[0][0], tokenType(real.Head))
	assert.Equal(t, len(expect)-1, len(real.Alter))
	assert.Equal(t, expectNum, real.Number)

	for i := 1; i < len(expect); i++ {
		s := real.Alter[i-1]
		assert.Equal(t, len(expect[i]), len(s.Items))
		assert.Equal(t, i-1, s.SNumber)
		assert.Equal(t, expectNum, s.PNumber)

		for i, seqType := range expect[i] {
			assert.Equal(t, seqType, tokenType(s.Items[i]))
		}
	}
}

func TestPaserPrint(t *testing.T) {
	t.SkipNow()
	next := Tokenize(&RuneSeq{Runes: []rune(`sql_statement: simple_statement_or_begin EMPTY ';' 
          opt_end_of_input
		|simple_statement_or_begin END_OF_INPUT;
        |

        ;

		opt_end_of_input:
                | {a = 1} empty {print(a+1);arr={1,2,3,4}}
                | END_OF_INPUT _table

        tail_empty:
           SELECT @A := 1
           |
`)})
	_, productions, err := Parse(next)
	assert.Equal(t, nil, err)

	for _, p := range productions {
		fmt.Println("==========")
		fmt.Printf("%T\n", p.Head)
		fmt.Println(p.Head.OriginString())
		fmt.Println(p.Number)
		fmt.Printf("Alter len: %d\n", len(p.Alter))
		for _, s := range p.Alter {
			fmt.Println("---------")
			for _, t := range s.Items {
				fmt.Printf("%T\n", t)
				fmt.Println(t.OriginString())
			}
		}
	}
}

package yacc_parser

type Heap interface {
	// conctent is the extended branch
	// sql is the related sql
	Push(content string, sql *string)
}

package yacc_parser

type PendingPath struct {
	// product result in this branch
	Content string
	TheSeq  *Seq
}

type Heap interface {
	Push(content string, sql string)
}

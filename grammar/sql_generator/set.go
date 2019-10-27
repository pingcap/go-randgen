package sql_generator

import "github.com/pingcap/go-randgen/grammar/yacc_parser"

type ProductionSet struct {
	Productions []*yacc_parser.Production
	set         map[int]bool
}

func newProductionSet() *ProductionSet {
	return &ProductionSet{set:make(map[int]bool)}
}

func (p *ProductionSet) add(producton *yacc_parser.Production)  {
	if !p.set[producton.Number] {
		p.set[producton.Number] = true
		p.Productions = append(p.Productions, producton)
	}
}

func (p *ProductionSet) clear()  {
	p.Productions = p.Productions[:0]
	for pnumber := range p.set {
		delete(p.set, pnumber)
	}
}


type SeqSet struct {
	Seqs []*yacc_parser.Seq
	set  map[[2]int]bool
}

func newSeqSet() *SeqSet {
	return &SeqSet{set:make(map[[2]int]bool)}
}

func (s *SeqSet) add(seq *yacc_parser.Seq)  {
	if !s.set[[2]int{seq.PNumber, seq.SNumber}] {
		s.set[[2]int{seq.PNumber, seq.SNumber}] = true
		s.Seqs = append(s.Seqs, seq)
	}
}

func (s *SeqSet) clear() {
	s.Seqs = s.Seqs[:0]
	for ps := range s.set {
		delete(s.set, ps)
	}
}

package yacc_parser

type stack struct {
	data []rune
}

func (s *stack) push(r rune)  {
	s.data = append(s.data, r)
}

func (s *stack) pop() rune {
	last := s.data[len(s.data) - 1]
	s.data = s.data[:len(s.data)-1]
	return last
}

func (s *stack) peek() rune {
	return s.data[len(s.data) - 1]
}

func (s *stack) empty() bool {
	return len(s.data) == 0
}



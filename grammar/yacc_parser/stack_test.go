package yacc_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack(t *testing.T)  {
	s := stack{}
	s.push('{')
	assert.Equal(t, false, s.empty())

	s.push('}')
	assert.Equal(t, '}', s.peek())

	s.push('p')
	assert.Equal(t, 'p', s.peek())

	p := s.pop()
	assert.Equal(t, 'p', p)
	assert.Equal(t, '}', s.peek())

	s.pop()
	s.pop()
	assert.Equal(t, true, s.empty())
}

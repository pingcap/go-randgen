package generators

import "math/rand"

var randChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Char struct {
	length int
}

func NewChar(length int) *Char {
	return &Char{length:length}
}

func (c *Char) Gen() string {
	b := make([]rune, c.length)
	for i := range b {
		b[i] = randChars[rand.Intn(len(randChars))]
	}
	return `"` + string(b) + `"`
}


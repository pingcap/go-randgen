package generators

import (
	"math/rand"
	"strconv"
)

type Digit struct {
}

func (d *Digit) Gen() string {
	return strconv.Itoa(rand.Intn(10))
}

package generators

import (
	"math/rand"
	"strconv"
)

type Uint struct {
}

func (*Uint) Gen() string {
	return strconv.FormatInt(int64(rand.Uint32()), 10)
}

package generators

import (
	"math/rand"
	"strconv"
)

type BigInt struct {
	unsigned bool
	flags    []int64
}

func newBigInt(unsigned bool) *BigInt {
	return &BigInt{
		unsigned:unsigned,
		flags:[]int64{-1, 1},
	}
}

func (b *BigInt) Gen() string {
	if b.unsigned {
		return strconv.FormatUint(rand.Uint64(), 10)
	} else {
		flag := b.flags[rand.Intn(len(b.flags))]
		return strconv.FormatInt(flag * rand.Int63(), 10)
	}
}


package generators

import (
	"math/rand"
	"strconv"
)

// min and max are inclusive
// when max < min, it will generate random int
type Int struct {
	min int
	max int
}

func newInt(min int, max int) *Int {
	return &Int{min, max}
}

var flag = []int{-1, 1}

func (i *Int) Gen() string {
	if i.max < i.min {
		return strconv.Itoa(flag[rand.Intn(len(flag))] * rand.Int())
	}
	return strconv.Itoa(randInRange(i.min, i.max))
}
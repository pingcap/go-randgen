package generators

import (
	"fmt"
	"math/rand"
	"strconv"
)

// min and max are inclusive
// when max < min, it will generate random int
type Int struct {
	min int
	max int
	// string template
	tmpl string
}

func newInt(min int, max int, tmpl string) *Int {
	return &Int{min, max, tmpl}
}

var flag = []int{-1, 1}

func (i *Int) Gen() string {
	var intRes int
	if i.max < i.min {
		intRes = flag[rand.Intn(len(flag))] * rand.Int()
	} else {
		intRes = randInRange(i.min, i.max)
	}

	if i.tmpl == "" {
		return strconv.Itoa(intRes)
	}

	return fmt.Sprintf(i.tmpl, intRes)
}

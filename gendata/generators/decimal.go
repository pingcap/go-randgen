package generators

import (
	"fmt"
	"strconv"
)

type Decimal struct {
}

func (*Decimal) Gen() string {
	return strconv.Itoa(randInRange(0, 100)) +
		"." + fmt.Sprintf("%04d", randInRange(0, 1999))
}

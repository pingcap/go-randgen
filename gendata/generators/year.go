package generators

import "strconv"

type Year struct {
}

func (y *Year) Gen() string {
	return strconv.Itoa(randInRange(2000, 2019))
}


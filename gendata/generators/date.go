package generators

import "fmt"

type Date struct {
}

func (d *Date) Gen() string {
	return fmt.Sprintf("%04d-%02d-%02d",
		randInRange(2000, 2019),
		randInRange(1, 12),
		randInRange(1, 28))
}


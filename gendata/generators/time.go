package generators

import "fmt"

type Time struct {
}

func (time *Time) Gen() string {
	return fmt.Sprintf("%02d:%02d:%02d",
		randInRange(0 ,23),
		randInRange(0, 59),
		randInRange(0, 59))
}


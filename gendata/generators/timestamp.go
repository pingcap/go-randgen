package generators

import "fmt"

// 'timestamp' like mysql randgen, points to time format 'yyyymmddhhmmss'
type Timestamp struct {
}

func (ts *Timestamp) Gen() string {
	return fmt.Sprintf("%04d%02d%02d%02d%02d%02d",
		randInRange(2000, 2019),
		randInRange(1, 12),
		randInRange(1, 28),
		randInRange(0, 23),
		randInRange(0, 59),
		randInRange(0, 59))
}

package generators

import "math/rand"

// include start and end
func randInRange(start int, end int) int {
	return start + rand.Intn(end-start+1)
}

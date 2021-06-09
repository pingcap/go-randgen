package generators

import "fmt"

type Letter struct {
}

func (l *Letter) Gen() string {
	return fmt.Sprintf("'%c'", randInRange('a', 'z'))
}

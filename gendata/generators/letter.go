package generators

type Letter struct {
}

func (l *Letter) Gen() string {
	return string(randInRange('a', 'z'))
}


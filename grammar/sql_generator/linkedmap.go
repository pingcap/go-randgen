package sql_generator

type linkedMap struct {
	order []string
	m     map[string]int
}

func newLinkedMap() *linkedMap {
	return &linkedMap{m: make(map[string]int)}
}

func (l *linkedMap) enter(key string) {
	l.order = append(l.order, key)
	l.m[key]++
}

func (l *linkedMap) leave(key string)  {
	l.m[key]--
	l.order = l.order[:len(l.order) - 1]
}

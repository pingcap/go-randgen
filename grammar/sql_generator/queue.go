package sql_generator

import "container/list"

type queue struct {
	datas *list.List
}

func newQueue() *queue {
	return &queue{datas:list.New()}
}

func (q *queue) enqueue(item string) {
	q.datas.PushBack(item)
}

func (q *queue) dequeue() string {
	return q.datas.Remove(q.datas.Front()).(string)
}

func (q *queue) isEmpty() bool {
	return q.datas.Len() == 0
}

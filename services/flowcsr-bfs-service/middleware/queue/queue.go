package queue

import "github.com/gin-gonic/gin"

var __q *ItemQueue

func Q() *ItemQueue {
	return __q
}

func InitQ() {
	__q = &ItemQueue{items: []Item{}}
}

type Item struct {
	C *gin.Context
}

// Item the type of the queue
type ItemQueue struct {
	items []Item
}

type ItemQueuer interface {
	Enqueue(t Item)
	Dequeue() *Item
	IsEmpty() bool
	Size() int
}

// Enqueue adds an Item to the end of the queue
func (s *ItemQueue) Enqueue(t Item) {
	s.items = append(s.items, t)
}

// dequeue
func (s *ItemQueue) Dequeue() Item {
	item := s.items[0] // 先进先出
	if s.Size() == 1 {
		s.items = []Item{}
	} else {
		s.items = s.items[1:len(s.items)]
	}
	return item
}

func (s *ItemQueue) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of Items in the queue
func (s *ItemQueue) Size() int {
	return len(s.items)
}

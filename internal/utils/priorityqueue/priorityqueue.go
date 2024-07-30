package priorityqueue

import (
	"container/heap"
	"time"
)

type Item struct {
	Value     int64
	Timestamp time.Time
	Priority  int
	index     int
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Empty() bool {
	return len(pq) <= 0
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	if pq[i].Priority < pq[j].Priority {
		return true
	}
	if pq[i].Priority > pq[j].Priority {
		return false
	}
	return pq[i].Timestamp.Before(pq[j].Timestamp)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Clear() {
	for !pq.Empty() {
		pq.Pop()
	}
}

func (pq *PriorityQueue) update(item *Item, value int64, priority int) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

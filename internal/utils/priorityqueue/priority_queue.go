package priorityqueue

import "github.com/quic-go/quic-go/internal/monotime"

type Item struct {
	Value     int64
	Timestamp monotime.Time
	Priority  uint32
	index     int
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Empty() bool {
	return len(pq) <= 0
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Priority > pq[j].Priority {
		return true
	}
	if pq[i].Priority < pq[j].Priority {
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

package search

import "container/heap"

type item struct {
	node     int
	priority uint64
	distance uint64
}
type minHeap []item

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].priority == h[j].priority {
		return h[i].distance < h[j].distance
	}
	return h[i].priority < h[j].priority
}
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)   { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() any     { old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x }
func push(h *minHeap, x item)   { heap.Push(h, x) }
func pop(h *minHeap) item       { return heap.Pop(h).(item) }

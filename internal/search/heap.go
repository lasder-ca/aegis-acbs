package search

type item struct {
	node     int
	priority uint64
	distance uint64
}

type minHeap []item

func (h minHeap) Len() int { return len(h) }

func lessItem(a, b item) bool {
	if a.priority == b.priority {
		return a.distance < b.distance
	}
	return a.priority < b.priority
}

// push inserts an item without interface boxing or container/heap overhead.
// The backing slice is owned by a pooled search workspace and reused between
// queries, so steady-state routing performs no priority-queue allocations.
func push(h *minHeap, x item) {
	*h = append(*h, x)
	i := len(*h) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if !lessItem((*h)[i], (*h)[parent]) {
			break
		}
		(*h)[i], (*h)[parent] = (*h)[parent], (*h)[i]
		i = parent
	}
}

func pop(h *minHeap) item {
	n := len(*h)
	root := (*h)[0]
	last := (*h)[n-1]
	(*h)[n-1] = item{}
	*h = (*h)[:n-1]
	if n == 1 {
		return root
	}
	(*h)[0] = last
	for i := 0; ; {
		left := i*2 + 1
		if left >= len(*h) {
			break
		}
		right := left + 1
		smallest := left
		if right < len(*h) && lessItem((*h)[right], (*h)[left]) {
			smallest = right
		}
		if !lessItem((*h)[smallest], (*h)[i]) {
			break
		}
		(*h)[i], (*h)[smallest] = (*h)[smallest], (*h)[i]
		i = smallest
	}
	return root
}

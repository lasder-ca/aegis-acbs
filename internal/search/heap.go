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

const heapArity = 4

// push inserts an item into a four-ary heap without interface boxing or
// container/heap overhead. A four-ary heap reduces the number of cache-missing
// levels visited by pop on the large OPEN lists common in road routing.
func push(h *minHeap, x item) {
	*h = append(*h, x)
	i := len(*h) - 1
	for i > 0 {
		parent := (i - 1) / heapArity
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
		first := i*heapArity + 1
		if first >= len(*h) {
			break
		}
		smallest := first
		limit := min(first+heapArity, len(*h))
		for child := first + 1; child < limit; child++ {
			if lessItem((*h)[child], (*h)[smallest]) {
				smallest = child
			}
		}
		if !lessItem((*h)[smallest], (*h)[i]) {
			break
		}
		(*h)[i], (*h)[smallest] = (*h)[smallest], (*h)[i]
		i = smallest
	}
	return root
}

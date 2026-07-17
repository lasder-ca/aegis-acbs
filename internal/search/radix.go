package search

import "math/bits"

// radixHeap is a monotone priority queue for uint64 keys. It is valid for
// Dijkstra-style searches whose newly inserted keys are never smaller than the
// most recently removed key. ACBS reduced edge costs are non-negative, so each
// directional frontier satisfies that requirement independently.
type radixHeap struct {
	buckets [65][]item
	last    uint64
	size    int
}

func (h *radixHeap) Len() int { return h.size }

func (h *radixHeap) reset() {
	for i := range h.buckets {
		h.buckets[i] = h.buckets[i][:0]
	}
	h.last = 0
	h.size = 0
}

func radixBucket(key, last uint64) int {
	if key == last {
		return 0
	}
	return bits.Len64(key ^ last)
}

func radixPush(h *radixHeap, x item) {
	// A lower key would violate monotonicity and invalidate the data structure.
	// Exact reduced-cost searches should never reach this branch. Keeping a
	// defensive fallback in bucket zero would silently corrupt ordering, so panic
	// in tests and fail fast in development instead.
	if x.priority < h.last {
		panic("radix heap received a non-monotone key")
	}
	b := radixBucket(x.priority, h.last)
	h.buckets[b] = append(h.buckets[b], x)
	h.size++
}

func (h *radixHeap) prepareMin() {
	if len(h.buckets[0]) > 0 || h.size == 0 {
		return
	}
	bucket := 1
	for bucket < len(h.buckets) && len(h.buckets[bucket]) == 0 {
		bucket++
	}
	items := h.buckets[bucket]
	newLast := items[0].priority
	for i := 1; i < len(items); i++ {
		if items[i].priority < newLast {
			newLast = items[i].priority
		}
	}
	h.last = newLast
	h.buckets[bucket] = h.buckets[bucket][:0]
	for _, x := range items {
		b := radixBucket(x.priority, h.last)
		h.buckets[b] = append(h.buckets[b], x)
	}
}

func radixPeek(h *radixHeap) (item, bool) {
	if h.size == 0 {
		return item{}, false
	}
	h.prepareMin()
	items := h.buckets[0]
	return items[len(items)-1], true
}

func radixPop(h *radixHeap) item {
	h.prepareMin()
	items := h.buckets[0]
	n := len(items)
	x := items[n-1]
	items[n-1] = item{}
	h.buckets[0] = items[:n-1]
	h.size--
	return x
}

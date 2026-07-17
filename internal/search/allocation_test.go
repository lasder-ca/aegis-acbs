package search

import (
	"context"
	"math/rand"
	"testing"
)

func TestMinHeapOrdering(t *testing.T) {
	rng := rand.New(rand.NewSource(5050))
	var h minHeap
	items := make([]item, 10_000)
	for i := range items {
		items[i] = item{
			node:     i,
			priority: uint64(rng.Intn(1_000)),
			distance: uint64(rng.Intn(1_000)),
		}
		push(&h, items[i])
	}
	previous := pop(&h)
	for h.Len() > 0 {
		current := pop(&h)
		if lessItem(current, previous) {
			t.Fatalf("heap order regressed: previous=%+v current=%+v", previous, current)
		}
		previous = current
	}
}

func TestACBSSteadyStateAllocationsRemainBounded(t *testing.T) {
	g := gridGraph(t, 120, 120, true)
	ctx := context.Background()
	if _, err := Run(ctx, g, 0, len(g.Nodes)-1, Aegis); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(20, func() {
		r, err := Run(ctx, g, 0, len(g.Nodes)-1, Aegis)
		if err != nil || !r.Stats.Reachable {
			t.Fatalf("ACBS failed: reachable=%v err=%v", r.Stats.Reachable, err)
		}
	})
	// One exact-sized path slice is expected. Leave a small margin for runtime
	// and race-detector instrumentation while preventing priority-queue allocation regressions.
	if allocations > 64 {
		t.Fatalf("steady-state ACBS allocations too high: %.2f", allocations)
	}
}

func BenchmarkACBSLargeGrid(b *testing.B) {
	g := gridGraph(b, 180, 180, true)
	ctx := context.Background()
	if _, err := Run(ctx, g, 0, len(g.Nodes)-1, Aegis); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := Run(ctx, g, 0, len(g.Nodes)-1, Aegis)
		if err != nil || !r.Stats.Reachable {
			b.Fatalf("ACBS failed: reachable=%v err=%v", r.Stats.Reachable, err)
		}
	}
}

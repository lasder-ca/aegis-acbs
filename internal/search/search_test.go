package search

import (
	"context"
	"math/rand"
	"testing"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
)

func TestBidirectionalAlgorithmsMatchDijkstraExhaustiveSmallDirected(t *testing.T) {
	ctx := context.Background()
	for n := 2; n <= 4; n++ {
		pairs := make([][2]int, 0, n*(n-1))
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i != j {
					pairs = append(pairs, [2]int{i, j})
				}
			}
		}
		total := 1 << len(pairs)
		for mask := 0; mask < total; mask++ {
			g := graph.New("test", "", "", graph.MetricDistance)
			g.Nodes = make([]graph.Node, n)
			g.Adj = make([][]graph.Edge, n)
			for i := range g.Nodes {
				g.Nodes[i] = graph.Node{ID: int64(i + 1), Lat: 35 + float64(i)*.001, Lon: 139 + float64(i)*.001}
			}
			for bit, p := range pairs {
				if mask&(1<<bit) != 0 {
					g.Adj[p[0]] = append(g.Adj[p[0]], graph.Edge{To: p[1], Cost: uint64(1 + (bit % 7))})
				}
			}
			if err := g.Finalize(); err != nil {
				t.Fatal(err)
			}
			for s := 0; s < n; s++ {
				for d := 0; d < n; d++ {
					a, err := Run(ctx, g, s, d, Dijkstra)
					if err != nil {
						t.Fatal(err)
					}
					for _, alg := range []Algorithm{BiDijkstra, Aegis, AegisStatic, AegisNoPrune} {
						b, err := Run(ctx, g, s, d, alg)
						if err != nil {
							t.Fatal(err)
						}
						if a.Stats.Reachable != b.Stats.Reachable || a.Stats.Distance != b.Stats.Distance || (b.Stats.Reachable && !Validate(g, s, d, b)) {
							t.Fatalf("n=%d mask=%d %d->%d alg=%s dijkstra=%+v got=%+v", n, mask, s, d, alg, a.Stats, b.Stats)
						}
					}
				}
			}
		}
	}
}

func TestAStarAndAegisMatchDijkstraRoadLike(t *testing.T) {
	rnd := rand.New(rand.NewSource(1010))
	ctx := context.Background()
	for caseNo := 0; caseNo < 100; caseNo++ {
		n := 25 + rnd.Intn(50)
		g := graph.New("road", "", "car", graph.MetricDistance)
		g.Nodes = make([]graph.Node, n)
		g.Adj = make([][]graph.Edge, n)
		for i := 0; i < n; i++ {
			g.Nodes[i] = graph.Node{ID: int64(i + 1), Lat: 35 + rnd.Float64()*.1, Lon: 139 + rnd.Float64()*.1}
		}
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if rnd.Float64() < .08 {
					m := graph.HaversineMeters(g.Nodes[i].Lat, g.Nodes[i].Lon, g.Nodes[j].Lat, g.Nodes[j].Lon)
					cost := uint64(m*1000) + uint64(rnd.Intn(5000))
					g.Adj[i] = append(g.Adj[i], graph.Edge{To: j, Cost: cost})
					g.Adj[j] = append(g.Adj[j], graph.Edge{To: i, Cost: cost})
				}
			}
		}
		if err := g.Finalize(); err != nil {
			t.Fatal(err)
		}
		for q := 0; q < 30; q++ {
			s, d := rnd.Intn(n), rnd.Intn(n)
			a, _ := Run(ctx, g, s, d, Dijkstra)
			for _, alg := range []Algorithm{AStar, BiDijkstra, Aegis, AegisStatic, AegisNoPrune, AegisRace} {
				b, err := Run(ctx, g, s, d, alg)
				if err != nil {
					t.Fatal(err)
				}
				if a.Stats.Reachable != b.Stats.Reachable || a.Stats.Distance != b.Stats.Distance {
					t.Fatalf("case=%d alg=%s %d->%d expected=%+v got=%+v", caseNo, alg, s, d, a.Stats, b.Stats)
				}
			}
		}
	}
}

func TestLegacyPortfolioUsesDijkstraForTinyGraph(t *testing.T) {
	g := gridGraph(t, 10, 10, true)
	if got := Select(g, 0, len(g.Nodes)-1); got != Dijkstra {
		t.Fatalf("expected dijkstra for tiny graph, got %s", got)
	}
	d := Explain(g, 0, len(g.Nodes)-1)
	if d.Selected != Dijkstra || d.Reason != "small_graph" {
		t.Fatalf("unexpected decision: %+v", d)
	}
}

func TestLegacyPortfolioUsesAStarForLargeRegionalRoadGraph(t *testing.T) {
	g := gridGraph(t, 72, 72, true)
	if got := Select(g, 0, len(g.Nodes)-1); got != AStar {
		t.Fatalf("expected astar for large regional road graph, got %s; decision=%+v", got, Explain(g, 0, len(g.Nodes)-1))
	}
}

func TestLegacyPortfolioUsesBiDijkstraWithoutCoordinates(t *testing.T) {
	g := gridGraph(t, 72, 72, false)
	if got := Select(g, 0, len(g.Nodes)-1); got != BiDijkstra {
		t.Fatalf("expected bidijkstra without coordinates, got %s; decision=%+v", got, Explain(g, 0, len(g.Nodes)-1))
	}
}

func gridGraph(t testing.TB, rows, cols int, coordinates bool) *graph.Graph {
	t.Helper()
	n := rows * cols
	g := graph.New("grid", "", "car", graph.MetricDistance)
	g.Nodes = make([]graph.Node, n)
	g.Adj = make([][]graph.Edge, n)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			i := r*cols + c
			lat, lon := 0.0, 0.0
			if coordinates {
				lat = 35 + float64(r)*0.001
				lon = 139 + float64(c)*0.001
			}
			g.Nodes[i] = graph.Node{ID: int64(i + 1), Lat: lat, Lon: lon}
			if c > 0 {
				g.Adj[i] = append(g.Adj[i], graph.Edge{To: i - 1, Cost: 100_000})
			}
			if c+1 < cols {
				g.Adj[i] = append(g.Adj[i], graph.Edge{To: i + 1, Cost: 100_000})
			}
			if r > 0 {
				g.Adj[i] = append(g.Adj[i], graph.Edge{To: i - cols, Cost: 100_000})
			}
			if r+1 < rows {
				g.Adj[i] = append(g.Adj[i], graph.Edge{To: i + cols, Cost: 100_000})
			}
		}
	}
	if err := g.Finalize(); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestLegacyPortfolioSplitsTimeRoutesByDistance(t *testing.T) {
	g := gridGraph(t, 72, 72, true)
	g.Metric = graph.MetricTime
	g.HeuristicStrength = 0.25

	localTarget := 8*72 + 8
	if got := Select(g, 0, localTarget); got != AStar {
		t.Fatalf("expected astar for local time route, got %s; decision=%+v", got, Explain(g, 0, localTarget))
	}
	if got := Select(g, 0, len(g.Nodes)-1); got != BiDijkstra {
		t.Fatalf("expected bidijkstra for long time route, got %s; decision=%+v", got, Explain(g, 0, len(g.Nodes)-1))
	}
	local := Explain(g, 0, localTarget)
	if local.PolicyVersion != "road-v3-time-aware" || local.AStarRatioLimit <= 0 || local.Metric != graph.MetricTime {
		t.Fatalf("unexpected time decision metadata: %+v", local)
	}
}

func TestWorkspaceReuseDoesNotLeakDistances(t *testing.T) {
	g := gridGraph(t, 20, 20, true)
	ctx := context.Background()
	pairs := [][2]int{{0, 399}, {399, 0}, {10, 11}, {50, 350}, {200, 25}}
	for round := 0; round < 50; round++ {
		for _, pair := range pairs {
			want, err := Run(ctx, g, pair[0], pair[1], Dijkstra)
			if err != nil {
				t.Fatal(err)
			}
			for _, alg := range []Algorithm{AStar, BiDijkstra, Aegis} {
				got, err := Run(ctx, g, pair[0], pair[1], alg)
				if err != nil {
					t.Fatal(err)
				}
				if got.Stats.Reachable != want.Stats.Reachable || got.Stats.Distance != want.Stats.Distance || !Validate(g, pair[0], pair[1], got) {
					t.Fatalf("round=%d pair=%v alg=%s want=%+v got=%+v", round, pair, alg, want.Stats, got.Stats)
				}
			}
		}
	}
}

func TestACBSUsesBothDirectionsOnLargeGrid(t *testing.T) {
	g := gridGraph(t, 72, 72, true)
	r, err := Run(context.Background(), g, 0, len(g.Nodes)-1, Aegis)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Stats.Reachable || !Validate(g, 0, len(g.Nodes)-1, r) {
		t.Fatalf("invalid ACBS result: %+v", r.Stats)
	}
	if r.Stats.ForwardExpanded == 0 || r.Stats.BackwardExpanded == 0 || r.Stats.Chunks < 2 {
		t.Fatalf("expected coupled bidirectional work, got %+v", r.Stats)
	}
	if r.Stats.TerminationLowerBound > r.Stats.Distance {
		t.Fatalf("lower bound exceeds solution: %+v", r.Stats)
	}
}

func TestACBSBalancedPotentialHasNonnegativeReducedEdges(t *testing.T) {
	g := gridGraph(t, 40, 40, true)
	w := acquireBiWorkspace(len(g.Nodes))
	defer releaseBiWorkspace(w)
	source, target := 0, len(g.Nodes)-1
	potential := newACBSPotential(g, source, target)
	for from, edges := range g.Adj {
		phiFrom := w.potential(g, potential, from)
		for _, e := range edges {
			phiTo := w.potential(g, potential, e.To)
			forward := int64(2*e.Cost) + phiTo - phiFrom
			backward := int64(2*e.Cost) + phiFrom - phiTo
			if forward < 0 || backward < 0 {
				t.Fatalf("negative reduced edge %d->%d: forward=%d backward=%d", from, e.To, forward, backward)
			}
		}
	}
}

func TestChordPotentialIsAdmissibleAgainstGreatCircle(t *testing.T) {
	rnd := rand.New(rand.NewSource(20260717))
	for i := 0; i < 10_000; i++ {
		a := graph.Node{Lat: -80 + rnd.Float64()*160, Lon: -180 + rnd.Float64()*360}
		b := graph.Node{Lat: -80 + rnd.Float64()*160, Lon: -180 + rnd.Float64()*360}
		ax, ay, az := graph.EarthUnit(a.Lat, a.Lon)
		bx, by, bz := graph.EarthUnit(b.Lat, b.Lon)
		chord := chordUnitMeters(ax, ay, az, bx, by, bz)
		arc := graph.HaversineMeters(a.Lat, a.Lon, b.Lat, b.Lon)
		if chord > arc+1e-6 {
			t.Fatalf("chord exceeds arc: chord=%f arc=%f", chord, arc)
		}
	}
}

func TestACBSReturnsZeroGapCertificateAndPrunes(t *testing.T) {
	g := gridGraph(t, 96, 96, true)
	r, err := Run(context.Background(), g, 0, len(g.Nodes)-1, Aegis)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Stats.Reachable || !Validate(g, 0, len(g.Nodes)-1, r) {
		t.Fatalf("invalid ACBS result: %+v", r.Stats)
	}
	if r.Stats.UpperBound != r.Stats.Distance || r.Stats.LowerBound != r.Stats.Distance || r.Stats.OptimalityGap != 0 {
		t.Fatalf("invalid optimality certificate: %+v", r.Stats)
	}
	if r.Stats.PotentialModel != acbsPotentialModel || r.Stats.SchedulerVersion != acbsSchedulerVersion {
		t.Fatalf("missing algorithm metadata: %+v", r.Stats)
	}
	if r.Stats.PotentialEvaluations == 0 || r.Stats.UpperBoundUpdates == 0 {
		t.Fatalf("expected potential and incumbent activity: %+v", r.Stats)
	}
}

func TestACBSMatchesDijkstraOnRandomDirectedTimeGraphs(t *testing.T) {
	rnd := rand.New(rand.NewSource(1202))
	ctx := context.Background()
	for caseNo := 0; caseNo < 250; caseNo++ {
		n := 8 + rnd.Intn(70)
		g := graph.New("time-road", "", "car", graph.MetricTime)
		g.Nodes = make([]graph.Node, n)
		g.Adj = make([][]graph.Edge, n)
		for i := range g.Nodes {
			g.Nodes[i] = graph.Node{ID: int64(i + 1), Lat: 35 + rnd.Float64()*.25, Lon: 139 + rnd.Float64()*.25}
		}
		// A directed backbone guarantees reachability in one direction, while
		// random reverse and shortcut edges exercise asymmetric frontiers.
		for i := 0; i+1 < n; i++ {
			m := graph.HaversineMeters(g.Nodes[i].Lat, g.Nodes[i].Lon, g.Nodes[i+1].Lat, g.Nodes[i+1].Lon)
			cost := uint64(m*(40+rnd.Float64()*140)) + 1
			g.Adj[i] = append(g.Adj[i], graph.Edge{To: i + 1, Cost: cost})
			if rnd.Float64() < .7 {
				g.Adj[i+1] = append(g.Adj[i+1], graph.Edge{To: i, Cost: cost + uint64(rnd.Intn(2000))})
			}
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i == j || rnd.Float64() >= .035 {
					continue
				}
				m := graph.HaversineMeters(g.Nodes[i].Lat, g.Nodes[i].Lon, g.Nodes[j].Lat, g.Nodes[j].Lon)
				g.Adj[i] = append(g.Adj[i], graph.Edge{To: j, Cost: uint64(m*(40+rnd.Float64()*160)) + 1})
			}
		}
		if err := g.Finalize(); err != nil {
			t.Fatal(err)
		}
		for q := 0; q < 40; q++ {
			s, d := rnd.Intn(n), rnd.Intn(n)
			want, err := Run(ctx, g, s, d, Dijkstra)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Run(ctx, g, s, d, Aegis)
			if err != nil {
				t.Fatal(err)
			}
			if got.Stats.Reachable != want.Stats.Reachable || got.Stats.Distance != want.Stats.Distance || (got.Stats.Reachable && !Validate(g, s, d, got)) {
				t.Fatalf("case=%d query=%d %d->%d want=%+v got=%+v", caseNo, q, s, d, want.Stats, got.Stats)
			}
			if got.Stats.Reachable && got.Stats.OptimalityGap != 0 {
				t.Fatalf("non-zero exactness gap: %+v", got.Stats)
			}
		}
	}
}

func TestACBSPrunesAfterEarlyIncumbent(t *testing.T) {
	g := graph.New("pruning", "", "car", graph.MetricDistance)
	g.Nodes = []graph.Node{
		{ID: 1, Lat: 35, Lon: 139},
		{ID: 2, Lat: 35, Lon: 139.001},
		{ID: 3, Lat: 35, Lon: 139.0004},
		{ID: 4, Lat: 35.01, Lon: 139.01},
		{ID: 5, Lat: 35.02, Lon: 139.02},
	}
	g.Adj = make([][]graph.Edge, len(g.Nodes))
	g.Adj[0] = []graph.Edge{
		{To: 1, Cost: 100_000}, // early feasible incumbent
		{To: 2, Cost: 20_000},  // better two-edge path
		{To: 3, Cost: 200_000}, // safely pruned after incumbent
		{To: 4, Cost: 250_000}, // safely pruned after incumbent
	}
	g.Adj[2] = []graph.Edge{{To: 1, Cost: 20_000}}
	if err := g.Finalize(); err != nil {
		t.Fatal(err)
	}
	want, _ := Run(context.Background(), g, 0, 1, Dijkstra)
	got, err := Run(context.Background(), g, 0, 1, Aegis)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stats.Distance != want.Stats.Distance || !Validate(g, 0, 1, got) {
		t.Fatalf("want=%+v got=%+v", want.Stats, got.Stats)
	}
	if got.Stats.UpperBoundUpdates < 2 || got.Stats.BoundPruned == 0 {
		t.Fatalf("expected early incumbent refinement and pruning: %+v", got.Stats)
	}
	if got.Stats.BoundPruned != got.Stats.PrunedAtPop+got.Stats.PrunedAtRelax {
		t.Fatalf("pruning counters do not add up: %+v", got.Stats)
	}
	if got.Stats.QueuePops == 0 || got.Stats.QueuePushes == 0 || got.Stats.MeetingChecks == 0 {
		t.Fatalf("missing queue or meeting counters: %+v", got.Stats)
	}
	if got.Stats.ConnectionChecks < got.Stats.FiniteMeetings || got.Stats.FiniteMeetings != got.Stats.MeetingChecks || got.Stats.FiniteMeetings < got.Stats.UpperBoundUpdates {
		t.Fatalf("invalid connection accounting: %+v", got.Stats)
	}
}

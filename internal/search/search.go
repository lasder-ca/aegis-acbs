package search

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
)

type Algorithm string

const (
	Dijkstra     Algorithm = "dijkstra"
	BiDijkstra   Algorithm = "bidijkstra"
	AStar        Algorithm = "astar"
	Aegis        Algorithm = "aegis"
	AegisStatic  Algorithm = "aegis-static"
	AegisNoPrune Algorithm = "aegis-no-prune"
	Portfolio    Algorithm = "portfolio"
	AegisRace    Algorithm = "aegis-race"
)

const policyVersion = "road-v3-time-aware"

// Decision exposes the deterministic features used by the Aegis selector.
// PredictedWork is a unitless relative estimate; it is not presented as time.
type Decision struct {
	PolicyVersion       string                `json:"policyVersion"`
	Selected            Algorithm             `json:"selected"`
	Reason              string                `json:"reason"`
	NodeCount           int                   `json:"nodeCount"`
	EdgeCount           int                   `json:"edgeCount"`
	StraightLineMeters  float64               `json:"straightLineMeters"`
	DistanceRatio       float64               `json:"distanceRatio"`
	AverageDegree       float64               `json:"averageDegree"`
	HeuristicStrength   float64               `json:"heuristicStrength"`
	Metric              graph.Metric          `json:"metric"`
	AStarRatioLimit     float64               `json:"aStarRatioLimit,omitempty"`
	SourceDegree        int                   `json:"sourceDegree"`
	TargetReverseDegree int                   `json:"targetReverseDegree"`
	PredictedWork       map[Algorithm]float64 `json:"predictedWork"`
}

type Stats struct {
	Algorithm               Algorithm `json:"algorithm"`
	Selected                Algorithm `json:"selected,omitempty"`
	DurationNS              int64     `json:"durationNs"`
	Expanded                uint64    `json:"expanded"`
	Relaxed                 uint64    `json:"relaxed"`
	QueuePushes             uint64    `json:"queuePushes"`
	Distance                uint64    `json:"distance"`
	Reachable               bool      `json:"reachable"`
	PathNodes               int       `json:"pathNodes"`
	ForwardExpanded         uint64    `json:"forwardExpanded,omitempty"`
	BackwardExpanded        uint64    `json:"backwardExpanded,omitempty"`
	DirectionSwitches       uint64    `json:"directionSwitches,omitempty"`
	Chunks                  uint64    `json:"chunks,omitempty"`
	FirstUpperBoundExpanded uint64    `json:"firstUpperBoundExpanded,omitempty"`
	TerminationLowerBound   uint64    `json:"terminationLowerBound,omitempty"`
	ForwardEfficiency       float64   `json:"forwardEfficiency,omitempty"`
	BackwardEfficiency      float64   `json:"backwardEfficiency,omitempty"`
	UpperBoundUpdates       uint64    `json:"upperBoundUpdates,omitempty"`
	BoundPruned             uint64    `json:"boundPruned,omitempty"`
	PotentialEvaluations    uint64    `json:"potentialEvaluations,omitempty"`
	UpperBound              uint64    `json:"upperBound,omitempty"`
	LowerBound              uint64    `json:"lowerBound,omitempty"`
	OptimalityGap           uint64    `json:"optimalityGap,omitempty"`
	SchedulerVersion        string    `json:"schedulerVersion,omitempty"`
	PotentialModel          string    `json:"potentialModel,omitempty"`
}

type Result struct {
	Path  []int `json:"path"`
	Stats Stats `json:"stats"`
}

func Run(ctx context.Context, g *graph.Graph, source, target int, alg Algorithm) (Result, error) {
	if source < 0 || source >= len(g.Nodes) || target < 0 || target >= len(g.Nodes) {
		return Result{}, errors.New("source or target is out of range")
	}
	started := time.Now()
	var r Result
	var err error
	switch alg {
	case Dijkstra:
		r, err = dijkstra(ctx, g, source, target, false)
	case AStar:
		if g.MinCostPerMeter <= 0 {
			return Result{}, errors.New("A* requires coordinates and an admissible cost-per-meter bound")
		}
		r, err = dijkstra(ctx, g, source, target, true)
	case BiDijkstra:
		r, err = bidirectionalDijkstra(ctx, g, source, target)
	case Aegis:
		r, err = acbs(ctx, g, source, target)
	case AegisStatic:
		r, err = acbsStatic(ctx, g, source, target)
	case AegisNoPrune:
		r, err = acbsNoPrune(ctx, g, source, target)
	case Portfolio:
		selected := Select(g, source, target)
		switch selected {
		case Dijkstra:
			r, err = dijkstra(ctx, g, source, target, false)
		case AStar:
			r, err = dijkstra(ctx, g, source, target, true)
		case BiDijkstra:
			r, err = bidirectionalDijkstra(ctx, g, source, target)
		default:
			err = fmt.Errorf("selector returned unsupported algorithm %q", selected)
		}
		r.Stats.Algorithm = Portfolio
		r.Stats.Selected = selected
	case AegisRace:
		r, err = race(ctx, g, source, target)
	default:
		return Result{}, fmt.Errorf("unknown algorithm %q", alg)
	}
	r.Stats.DurationNS = time.Since(started).Nanoseconds()
	return r, err
}

// Select returns the exact core algorithm chosen by the allocation-free road policy.
func Select(g *graph.Graph, source, target int) Algorithm {
	if len(g.Nodes) < 4096 {
		return Dijkstra
	}
	_, ratio, hasGeography := queryGeometry(g, source, target)
	if !hasGeography || g.HeuristicStrength < 0.05 {
		return BiDijkstra
	}

	// Travel-time weights have a much weaker lower bound than pure distance.
	// A* wins on local and medium routes, while two balanced frontiers are more
	// stable on long cross-region routes. The threshold scales with the measured
	// admissible-heuristic strength of the imported graph.
	if g.Metric == graph.MetricTime {
		if ratio <= timeAStarRatioLimit(g.HeuristicStrength) {
			return AStar
		}
		return BiDijkstra
	}

	// Distance graphs normally have a near-perfect geographic lower bound. Use
	// A* whenever that bound is useful; otherwise fall back to balanced frontiers.
	if g.HeuristicStrength >= 0.18 {
		return AStar
	}
	return BiDijkstra
}

// Explain returns the complete deterministic selector explanation for UI and
// diagnostics. It is kept off the timed routing hot path.
func Explain(g *graph.Graph, source, target int) Decision {
	selected := Select(g, source, target)
	return explainDecision(g, source, target, selected)
}

func queryGeometry(g *graph.Graph, source, target int) (straight, ratio float64, hasGeography bool) {
	hasGeography = g.MinCostPerMeter > 0 && g.DiameterMeters > 0
	if !hasGeography {
		return 0, 0.35, false
	}
	straight = graph.HaversineMeters(g.Nodes[source].Lat, g.Nodes[source].Lon, g.Nodes[target].Lat, g.Nodes[target].Lon)
	ratio = clamp(straight/g.DiameterMeters, 0, 1)
	return straight, ratio, true
}

func timeAStarRatioLimit(strength float64) float64 {
	// Tokyo's travel-time graph has strength ~=0.25, producing a limit ~=0.43.
	// This keeps local/random routes on A* and sends the longest regional routes
	// to bidirectional Dijkstra, reducing the heavy A* tail.
	return clamp(0.18+strength, 0.22, 0.62)
}

func explainDecision(g *graph.Graph, source, target int, selected Algorithm) Decision {
	n := len(g.Nodes)
	edges := g.EdgeCount
	avgDegree := g.AverageDegree
	if avgDegree <= 0 {
		avgDegree = float64(edges) / math.Max(1, float64(n))
	}
	straight, ratio, hasGeography := queryGeometry(g, source, target)
	strength := clamp(g.HeuristicStrength, 0, 1)
	limit := 0.0
	if g.Metric == graph.MetricTime {
		limit = timeAStarRatioLimit(strength)
	}

	searchFraction := clamp(0.008+0.82*math.Sqrt(ratio), 0.008, 0.95)
	dijkstraWork := 550 + float64(edges)*searchFraction*(1+0.018*float64(len(g.Adj[source])))
	biFraction := clamp(0.018+0.42*math.Sqrt(searchFraction), 0.018, 0.78)
	biWork := 1250 + float64(edges)*biFraction*(1.05+0.025*avgDegree)
	work := map[Algorithm]float64{Dijkstra: dijkstraWork, BiDijkstra: biWork}
	if hasGeography && strength >= 0.05 {
		heuristicGain := clamp(strength*(0.32+0.68*math.Sqrt(ratio)), 0, 0.9)
		aFraction := clamp(searchFraction*(1-0.82*heuristicGain), 0.006, 0.95)
		work[AStar] = 1450 + float64(edges)*aFraction*(1.20+0.035*avgDegree)
	}

	reason := "balanced_frontiers"
	switch {
	case selected == Dijkstra:
		reason = "small_graph"
	case selected == AStar && g.Metric == graph.MetricTime:
		reason = "time_local_or_medium_route"
	case selected == BiDijkstra && g.Metric == graph.MetricTime:
		reason = "time_long_route_tail_control"
	case selected == AStar:
		reason = "strong_distance_heuristic"
	case !hasGeography:
		reason = "coordinates_unavailable"
	case strength < 0.05:
		reason = "weak_geographic_heuristic"
	}

	return Decision{
		PolicyVersion:       policyVersion,
		Selected:            selected,
		Reason:              reason,
		NodeCount:           n,
		EdgeCount:           edges,
		StraightLineMeters:  straight,
		DistanceRatio:       ratio,
		AverageDegree:       avgDegree,
		HeuristicStrength:   g.HeuristicStrength,
		Metric:              g.Metric,
		AStarRatioLimit:     limit,
		SourceDegree:        len(g.Adj[source]),
		TargetReverseDegree: len(g.Rev[target]),
		PredictedWork:       work,
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func reconstruct(prev []int, source, target int) []int {
	if source == target {
		return []int{source}
	}
	if target < 0 || target >= len(prev) || prev[target] < 0 {
		return nil
	}
	path := make([]int, 0, 32)
	for v := target; ; v = prev[v] {
		path = append(path, v)
		if v == source {
			break
		}
		if v < 0 || prev[v] < 0 {
			return nil
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// Validate verifies that a reported path is continuous and that its edge costs
// add up to the reported distance. It is intended for benchmark correctness checks.
func Validate(g *graph.Graph, source, target int, r Result) bool {
	if !r.Stats.Reachable {
		return len(r.Path) == 0
	}
	if len(r.Path) == 0 || r.Path[0] != source || r.Path[len(r.Path)-1] != target {
		return false
	}
	var total uint64
	for i := 0; i+1 < len(r.Path); i++ {
		from, to := r.Path[i], r.Path[i+1]
		best := inf
		for _, e := range g.Adj[from] {
			if e.To == to && e.Cost < best {
				best = e.Cost
			}
		}
		if best == inf || total > inf-best {
			return false
		}
		total += best
	}
	return total == r.Stats.Distance
}

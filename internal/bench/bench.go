package bench

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
	"github.com/lasder-ca/aegis-acbs/internal/search"
	"github.com/lasder-ca/aegis-acbs/internal/version"
)

type Config struct {
	Queries    int                `json:"queries"`
	Seed       uint64             `json:"seed"`
	Algorithms []search.Algorithm `json:"algorithms"`
	Warmup     int                `json:"warmup"`
	Repeats    int                `json:"repeats"`
	BatchSize  int                `json:"batchSize"`
	Timeout    time.Duration      `json:"-"`
	Suite      string             `json:"suite"`
	PairMode   string             `json:"pairMode"`
}

type Query struct {
	Source             int     `json:"source"`
	Target             int     `json:"target"`
	StraightLineMeters float64 `json:"straightLineMeters"`
	Class              string  `json:"class"`
}

type Sample struct {
	QueryIndex         int          `json:"queryIndex"`
	QueryClass         string       `json:"queryClass"`
	StraightLineMeters float64      `json:"straightLineMeters"`
	SourceID           int64        `json:"sourceId"`
	TargetID           int64        `json:"targetId"`
	Stats              search.Stats `json:"stats"`
	Correct            bool         `json:"correct"`
	Error              string       `json:"error,omitempty"`
}

type Summary struct {
	Algorithm           search.Algorithm `json:"algorithm"`
	Runs                int              `json:"runs"`
	Reachable           int              `json:"reachable"`
	Correct             int              `json:"correct"`
	MedianNS            int64            `json:"medianNs"`
	P95NS               int64            `json:"p95Ns"`
	MedianEdges         uint64           `json:"medianRelaxed"`
	MedianExpanded      uint64           `json:"medianExpanded"`
	SpeedupVsDijkstra   float64          `json:"speedupVsDijkstra"`
	RelaxedReductionPct float64          `json:"relaxedReductionPct"`
}

type RegretPoint struct {
	QueryIndex int              `json:"queryIndex"`
	Class      string           `json:"class"`
	Fastest    search.Algorithm `json:"fastest"`
	Regret     float64          `json:"regret"`
}

type ClassSummary struct {
	Class       string           `json:"class"`
	Algorithm   search.Algorithm `json:"algorithm"`
	Runs        int              `json:"runs"`
	MedianNS    int64            `json:"medianNs"`
	P95NS       int64            `json:"p95Ns"`
	MedianEdges uint64           `json:"medianRelaxed"`
}

type DirectionTotals struct {
	ForwardExpanded  uint64 `json:"forwardExpanded"`
	BackwardExpanded uint64 `json:"backwardExpanded"`
}

type AegisSummary struct {
	Comparisons                   int                        `json:"comparisons"`
	MedianRuntimeRegret           float64                    `json:"medianRuntimeRegret"`
	P95RuntimeRegret              float64                    `json:"p95RuntimeRegret"`
	MaxRuntimeRegret              float64                    `json:"maxRuntimeRegret"`
	MedianSpeedupVsDijkstra       float64                    `json:"medianSpeedupVsDijkstra"`
	MedianForwardShare            float64                    `json:"medianForwardShare"`
	P95ForwardShare               float64                    `json:"p95ForwardShare"`
	MedianDirectionSwitches       uint64                     `json:"medianDirectionSwitches"`
	MedianChunks                  uint64                     `json:"medianChunks"`
	MedianFirstUpperBoundFraction float64                    `json:"medianFirstUpperBoundFraction"`
	MedianTerminationTightness    float64                    `json:"medianTerminationTightness"`
	MedianBoundPruned             uint64                     `json:"medianBoundPruned"`
	MedianPotentialEvaluations    uint64                     `json:"medianPotentialEvaluations"`
	MedianUpperBoundUpdates       uint64                     `json:"medianUpperBoundUpdates"`
	MedianOptimalityGap           uint64                     `json:"medianOptimalityGap"`
	DirectionByClass              map[string]DirectionTotals `json:"directionByClass"`
	Regrets                       []RegretPoint              `json:"regrets"`
}

type Report struct {
	Version           string         `json:"version"`
	GeneratedAt       time.Time      `json:"generatedAt"`
	GoVersion         string         `json:"goVersion"`
	GOOS              string         `json:"goos"`
	GOARCH            string         `json:"goarch"`
	CPUs              int            `json:"cpus"`
	GraphName         string         `json:"graphName"`
	GraphSource       string         `json:"graphSource"`
	Nodes             int            `json:"nodes"`
	Edges             int            `json:"edges"`
	Metric            graph.Metric   `json:"metric"`
	Profile           string         `json:"profile"`
	HeuristicStrength float64        `json:"heuristicStrength"`
	Config            Config         `json:"config"`
	Queries           []Query        `json:"queryPairs"`
	Samples           []Sample       `json:"samples"`
	Summary           []Summary      `json:"summary"`
	ClassSummary      []ClassSummary `json:"classSummary"`
	Aegis             AegisSummary   `json:"aegis"`
	AllCorrect        bool           `json:"allCorrect"`
	QueryPoolSize     int            `json:"queryPoolSize"`
}

func Run(ctx context.Context, g *graph.Graph, cfg Config) (Report, error) {
	if cfg.Queries <= 0 {
		cfg.Queries = 100
	}
	if cfg.Warmup < 0 {
		cfg.Warmup = 0
	}
	if cfg.Repeats <= 0 {
		cfg.Repeats = 7
	}
	if cfg.Repeats > 101 {
		return Report{}, errors.New("repeats must be <= 101")
	}
	if cfg.Repeats%2 == 0 {
		cfg.Repeats++
	}
	if cfg.BatchSize <= 0 {
		switch {
		case len(g.Nodes) < 1_000:
			cfg.BatchSize = 64
		case len(g.Nodes) < 10_000:
			cfg.BatchSize = 16
		case len(g.Nodes) < 100_000:
			cfg.BatchSize = 4
		default:
			cfg.BatchSize = 1
		}
	}
	if cfg.BatchSize > 256 {
		return Report{}, errors.New("batch size must be <= 256")
	}
	if len(cfg.Algorithms) == 0 {
		cfg.Algorithms = []search.Algorithm{search.Dijkstra, search.BiDijkstra}
		if g.MinCostPerMeter > 0 {
			cfg.Algorithms = append(cfg.Algorithms, search.AStar)
		}
		cfg.Algorithms = append(cfg.Algorithms, search.Aegis)
	}
	if len(g.Nodes) < 2 {
		return Report{}, errors.New("graph needs at least two nodes")
	}
	if cfg.Suite == "" {
		cfg.Suite = "mixed"
	}
	if cfg.PairMode == "" {
		cfg.PairMode = "strongly-connected"
	}
	pool := makePool(g, cfg.PairMode)
	if len(pool) < 2 {
		return Report{}, errors.New("query pool needs at least two nodes")
	}
	queries := makeQueries(g, pool, cfg.Queries, cfg.Seed, cfg.Suite)
	for i := 0; i < cfg.Warmup && i < len(queries); i++ {
		for _, alg := range cfg.Algorithms {
			_, _ = search.Run(ctx, g, queries[i].Source, queries[i].Target, alg)
		}
	}

	report := Report{
		Version:           version.Version,
		GeneratedAt:       time.Now().UTC(),
		GoVersion:         runtime.Version(),
		GOOS:              runtime.GOOS,
		GOARCH:            runtime.GOARCH,
		CPUs:              runtime.NumCPU(),
		GraphName:         g.Name,
		GraphSource:       g.Source,
		Nodes:             len(g.Nodes),
		Edges:             g.EdgeCount,
		Metric:            g.Metric,
		Profile:           g.Profile,
		HeuristicStrength: g.HeuristicStrength,
		Config:            cfg,
		Queries:           queries,
		AllCorrect:        true,
		QueryPoolSize:     len(pool),
	}

	for qi, q := range queries {
		// Correctness reference is deliberately outside measured samples.
		expected, err := runOne(ctx, g, q, search.Dijkstra, cfg.Timeout)
		if err != nil {
			return report, err
		}
		expectedReachable, expectedDistance := expected.Stats.Reachable, expected.Stats.Distance

		order := rotated(cfg.Algorithms, qi)
		for _, alg := range order {
			r, runErr := runRepeated(ctx, g, q, alg, cfg.Timeout, cfg.Repeats, cfg.BatchSize)
			correct := runErr == nil && r.Stats.Reachable == expectedReachable && (!expectedReachable || r.Stats.Distance == expectedDistance) && search.Validate(g, q.Source, q.Target, r)
			if !correct {
				report.AllCorrect = false
			}
			s := Sample{
				QueryIndex:         qi,
				QueryClass:         q.Class,
				StraightLineMeters: q.StraightLineMeters,
				SourceID:           g.Nodes[q.Source].ID,
				TargetID:           g.Nodes[q.Target].ID,
				Stats:              r.Stats,
				Correct:            correct,
			}
			if runErr != nil {
				s.Error = runErr.Error()
			}
			report.Samples = append(report.Samples, s)
		}
	}

	report.Summary = summarize(report.Samples, cfg.Algorithms)
	report.ClassSummary = summarizeClasses(report.Samples, cfg.Algorithms)
	report.Aegis = summarizeAegis(report.Samples)
	return report, nil
}

func rotated(in []search.Algorithm, offset int) []search.Algorithm {
	if len(in) == 0 {
		return nil
	}
	out := make([]search.Algorithm, len(in))
	for i := range in {
		out[i] = in[(i+offset)%len(in)]
	}
	return out
}

func runRepeated(parent context.Context, g *graph.Graph, q Query, alg search.Algorithm, timeout time.Duration, repeats, batchSize int) (search.Result, error) {
	results := make([]search.Result, 0, repeats)
	for i := 0; i < repeats; i++ {
		var r search.Result
		started := time.Now()
		for batch := 0; batch < batchSize; batch++ {
			var err error
			r, err = runOne(parent, g, q, alg, timeout)
			if err != nil {
				return ResultZero(), err
			}
		}
		r.Stats.DurationNS = time.Since(started).Nanoseconds() / int64(batchSize)
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Stats.DurationNS < results[j].Stats.DurationNS })
	return results[len(results)/2], nil
}

// ResultZero keeps error paths explicit without exporting an implementation detail.
func ResultZero() search.Result { return search.Result{} }

func runOne(parent context.Context, g *graph.Graph, q Query, alg search.Algorithm, timeout time.Duration) (search.Result, error) {
	ctx := parent
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
	}
	defer cancel()
	return search.Run(ctx, g, q.Source, q.Target, alg)
}

func makePool(g *graph.Graph, mode string) []int {
	if mode == "all" {
		out := make([]int, len(g.Nodes))
		for i := range out {
			out[i] = i
		}
		return out
	}
	// Iterative Kosaraju: query selection only, never included in timings.
	n := len(g.Nodes)
	seen := make([]bool, n)
	order := make([]int, 0, n)
	type frame struct{ v, next int }
	for root := 0; root < n; root++ {
		if seen[root] {
			continue
		}
		seen[root] = true
		stack := []frame{{root, 0}}
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			if top.next < len(g.Adj[top.v]) {
				to := g.Adj[top.v][top.next].To
				top.next++
				if !seen[to] {
					seen[to] = true
					stack = append(stack, frame{to, 0})
				}
			} else {
				order = append(order, top.v)
				stack = stack[:len(stack)-1]
			}
		}
	}
	comp := make([]int, n)
	for i := range comp {
		comp[i] = -1
	}
	sizes := []int{}
	for oi := len(order) - 1; oi >= 0; oi-- {
		root := order[oi]
		if comp[root] >= 0 {
			continue
		}
		cid := len(sizes)
		size := 0
		stack := []int{root}
		comp[root] = cid
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			size++
			for _, e := range g.Rev[v] {
				if comp[e.To] < 0 {
					comp[e.To] = cid
					stack = append(stack, e.To)
				}
			}
		}
		sizes = append(sizes, size)
	}
	best := 0
	for i := 1; i < len(sizes); i++ {
		if sizes[i] > sizes[best] {
			best = i
		}
	}
	out := make([]int, 0, sizes[best])
	for v, c := range comp {
		if c == best {
			out = append(out, v)
		}
	}
	return out
}

func makeQueries(g *graph.Graph, pool []int, count int, seed uint64, suite string) []Query {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	n := len(pool)
	out := make([]Query, 0, count)
	for len(out) < count {
		s := pool[r.IntN(n)]
		class := suite
		if suite == "mixed" {
			switch len(out) % 3 {
			case 0:
				class = "local"
			case 1:
				class = "random"
			default:
				class = "regional"
			}
		}
		t := chooseTarget(g, pool, r, s, class)
		if s == t {
			continue
		}
		d := graph.HaversineMeters(g.Nodes[s].Lat, g.Nodes[s].Lon, g.Nodes[t].Lat, g.Nodes[t].Lon)
		out = append(out, Query{Source: s, Target: t, StraightLineMeters: d, Class: class})
	}
	return out
}

func chooseTarget(g *graph.Graph, pool []int, r *rand.Rand, source int, class string) int {
	n := len(pool)
	if class == "random" || n < 4 {
		for {
			t := pool[r.IntN(n)]
			if t != source {
				return t
			}
		}
	}
	best := -1
	bestD := 0.0
	if class == "local" {
		bestD = math.MaxFloat64
	}
	samples := 64
	if n < samples {
		samples = n
	}
	for i := 0; i < samples; i++ {
		t := pool[r.IntN(n)]
		if t == source {
			continue
		}
		d := graph.HaversineMeters(g.Nodes[source].Lat, g.Nodes[source].Lon, g.Nodes[t].Lat, g.Nodes[t].Lon)
		if (class == "local" && d < bestD) || (class == "regional" && d > bestD) {
			best, bestD = t, d
		}
	}
	if best < 0 {
		for _, v := range pool {
			if v != source {
				return v
			}
		}
		return source
	}
	return best
}

func summarize(samples []Sample, algs []search.Algorithm) []Summary {
	out := make([]Summary, 0, len(algs))
	for _, alg := range algs {
		durations := make([]int64, 0)
		relaxed := make([]uint64, 0)
		expanded := make([]uint64, 0)
		s := Summary{Algorithm: alg}
		for _, x := range samples {
			if x.Stats.Algorithm != alg {
				continue
			}
			s.Runs++
			if x.Stats.Reachable {
				s.Reachable++
			}
			if x.Correct {
				s.Correct++
			}
			durations = append(durations, x.Stats.DurationNS)
			relaxed = append(relaxed, x.Stats.Relaxed)
			expanded = append(expanded, x.Stats.Expanded)
		}
		if len(durations) > 0 {
			s.MedianNS = percentileInt64(durations, 0.5)
			s.P95NS = percentileInt64(durations, 0.95)
			s.MedianEdges = percentileUint64(relaxed, 0.5)
			s.MedianExpanded = percentileUint64(expanded, 0.5)
		}
		out = append(out, s)
	}
	var base *Summary
	for i := range out {
		if out[i].Algorithm == search.Dijkstra {
			base = &out[i]
			break
		}
	}
	if base != nil {
		for i := range out {
			if out[i].MedianNS > 0 {
				out[i].SpeedupVsDijkstra = float64(base.MedianNS) / float64(out[i].MedianNS)
			}
			if base.MedianEdges > 0 {
				out[i].RelaxedReductionPct = 100 * (1 - float64(out[i].MedianEdges)/float64(base.MedianEdges))
			}
		}
	}
	return out
}

func summarizeClasses(samples []Sample, algs []search.Algorithm) []ClassSummary {
	classes := []string{"local", "random", "regional"}
	out := make([]ClassSummary, 0, len(classes)*len(algs))
	for _, class := range classes {
		for _, alg := range algs {
			durations := []int64{}
			relaxed := []uint64{}
			for _, s := range samples {
				if s.QueryClass == class && s.Stats.Algorithm == alg {
					durations = append(durations, s.Stats.DurationNS)
					relaxed = append(relaxed, s.Stats.Relaxed)
				}
			}
			if len(durations) == 0 {
				continue
			}
			out = append(out, ClassSummary{Class: class, Algorithm: alg, Runs: len(durations), MedianNS: percentileInt64(durations, 0.5), P95NS: percentileInt64(durations, 0.95), MedianEdges: percentileUint64(relaxed, 0.5)})
		}
	}
	return out
}

func summarizeAegis(samples []Sample) AegisSummary {
	out := AegisSummary{DirectionByClass: map[string]DirectionTotals{}}
	byQuery := map[int][]Sample{}
	for _, s := range samples {
		byQuery[s.QueryIndex] = append(byQuery[s.QueryIndex], s)
	}
	runtimeRegrets := []float64{}
	speedups := []float64{}
	forwardShares := []float64{}
	switches := []uint64{}
	chunks := []uint64{}
	upperFractions := []float64{}
	tightness := []float64{}
	pruned := []uint64{}
	potentialEvals := []uint64{}
	upperUpdates := []uint64{}
	gaps := []uint64{}
	indices := make([]int, 0, len(byQuery))
	for i := range byQuery {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	for _, qi := range indices {
		group := byQuery[qi]
		core := map[search.Algorithm]Sample{}
		var aegis *Sample
		for i := range group {
			s := group[i]
			switch s.Stats.Algorithm {
			case search.Dijkstra, search.BiDijkstra, search.AStar:
				core[s.Stats.Algorithm] = s
			case search.Aegis:
				copy := s
				aegis = &copy
			}
		}
		if aegis == nil || aegis.Stats.DurationNS <= 0 {
			continue
		}
		fastest := search.Algorithm("")
		fastestNS := int64(math.MaxInt64)
		for _, alg := range []search.Algorithm{search.Dijkstra, search.BiDijkstra, search.AStar} {
			if sample, ok := core[alg]; ok && sample.Stats.DurationNS > 0 && sample.Stats.DurationNS < fastestNS {
				fastest, fastestNS = alg, sample.Stats.DurationNS
			}
		}
		if fastest == "" {
			continue
		}
		regret := float64(aegis.Stats.DurationNS) / float64(fastestNS)
		runtimeRegrets = append(runtimeRegrets, regret)
		out.Regrets = append(out.Regrets, RegretPoint{QueryIndex: qi, Class: aegis.QueryClass, Fastest: fastest, Regret: regret})
		if d, ok := core[search.Dijkstra]; ok && d.Stats.DurationNS > 0 {
			speedups = append(speedups, float64(d.Stats.DurationNS)/float64(aegis.Stats.DurationNS))
		}
		totalDirections := aegis.Stats.ForwardExpanded + aegis.Stats.BackwardExpanded
		if totalDirections > 0 {
			forwardShares = append(forwardShares, float64(aegis.Stats.ForwardExpanded)/float64(totalDirections))
		}
		switches = append(switches, aegis.Stats.DirectionSwitches)
		chunks = append(chunks, aegis.Stats.Chunks)
		if aegis.Stats.Expanded > 0 && aegis.Stats.FirstUpperBoundExpanded > 0 {
			upperFractions = append(upperFractions, float64(aegis.Stats.FirstUpperBoundExpanded)/float64(aegis.Stats.Expanded))
		}
		if aegis.Stats.Distance > 0 && aegis.Stats.TerminationLowerBound > 0 {
			tightness = append(tightness, float64(aegis.Stats.TerminationLowerBound)/float64(aegis.Stats.Distance))
		}
		pruned = append(pruned, aegis.Stats.BoundPruned)
		potentialEvals = append(potentialEvals, aegis.Stats.PotentialEvaluations)
		upperUpdates = append(upperUpdates, aegis.Stats.UpperBoundUpdates)
		gaps = append(gaps, aegis.Stats.OptimalityGap)
		totals := out.DirectionByClass[aegis.QueryClass]
		totals.ForwardExpanded += aegis.Stats.ForwardExpanded
		totals.BackwardExpanded += aegis.Stats.BackwardExpanded
		out.DirectionByClass[aegis.QueryClass] = totals
		out.Comparisons++
	}
	if len(runtimeRegrets) > 0 {
		out.MedianRuntimeRegret = percentileFloat64(runtimeRegrets, 0.5)
		out.P95RuntimeRegret = percentileFloat64(runtimeRegrets, 0.95)
		out.MaxRuntimeRegret = percentileFloat64(runtimeRegrets, 1)
	}
	if len(speedups) > 0 {
		out.MedianSpeedupVsDijkstra = percentileFloat64(speedups, 0.5)
	}
	if len(forwardShares) > 0 {
		out.MedianForwardShare = percentileFloat64(forwardShares, 0.5)
		out.P95ForwardShare = percentileFloat64(forwardShares, 0.95)
	}
	if len(switches) > 0 {
		out.MedianDirectionSwitches = percentileUint64(switches, 0.5)
	}
	if len(chunks) > 0 {
		out.MedianChunks = percentileUint64(chunks, 0.5)
	}
	if len(upperFractions) > 0 {
		out.MedianFirstUpperBoundFraction = percentileFloat64(upperFractions, 0.5)
	}
	if len(tightness) > 0 {
		out.MedianTerminationTightness = percentileFloat64(tightness, 0.5)
	}
	if len(pruned) > 0 {
		out.MedianBoundPruned = percentileUint64(pruned, 0.5)
	}
	if len(potentialEvals) > 0 {
		out.MedianPotentialEvaluations = percentileUint64(potentialEvals, 0.5)
	}
	if len(upperUpdates) > 0 {
		out.MedianUpperBoundUpdates = percentileUint64(upperUpdates, 0.5)
	}
	if len(gaps) > 0 {
		out.MedianOptimalityGap = percentileUint64(gaps, 0.5)
	}
	return out
}

func percentileInt64(values []int64, p float64) int64 {
	v := append([]int64(nil), values...)
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	return v[percentileIndex(len(v), p)]
}

func percentileUint64(values []uint64, p float64) uint64 {
	v := append([]uint64(nil), values...)
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	return v[percentileIndex(len(v), p)]
}

func percentileFloat64(values []float64, p float64) float64 {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	return v[percentileIndex(len(v), p)]
}

func percentileIndex(n int, p float64) int {
	if n <= 1 {
		return 0
	}
	if p <= 0 {
		return 0
	}
	if p >= 1 {
		return n - 1
	}
	return int(math.Ceil(float64(n)*p)) - 1
}

func WriteJSON(path string, report Report) error {
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}

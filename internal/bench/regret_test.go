package bench

import (
	"testing"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
	"github.com/lasder-ca/aegis-acbs/internal/search"
)

func TestAnalyzeRegretSeparatesRatioNoiseFromMeaningfulPenalty(t *testing.T) {
	report := Report{
		Version:        "test",
		GraphName:      "time-road",
		Metric:         graph.MetricTime,
		DiameterMeters: 10_000,
		Samples: []Sample{
			{QueryIndex: 0, QueryClass: "local", StraightLineMeters: 100, Stats: search.Stats{Algorithm: search.Dijkstra, DurationNS: 1_000}, Correct: true},
			{QueryIndex: 0, QueryClass: "local", StraightLineMeters: 100, Stats: search.Stats{Algorithm: search.BiDijkstra, DurationNS: 2_000}, Correct: true},
			{QueryIndex: 0, QueryClass: "local", StraightLineMeters: 100, Stats: search.Stats{Algorithm: search.Aegis, DurationNS: 2_000, Expanded: 10}, Correct: true},
			{QueryIndex: 1, QueryClass: "regional", StraightLineMeters: 8_000, Stats: search.Stats{Algorithm: search.Dijkstra, DurationNS: 10_000_000}, Correct: true},
			{QueryIndex: 1, QueryClass: "regional", StraightLineMeters: 8_000, Stats: search.Stats{Algorithm: search.AStar, DurationNS: 8_000_000}, Correct: true},
			{QueryIndex: 1, QueryClass: "regional", StraightLineMeters: 8_000, Stats: search.Stats{Algorithm: search.Aegis, DurationNS: 12_000_000, Expanded: 100, ForwardExpanded: 70, BackwardExpanded: 30, Chunks: 10, DirectionSwitches: 4}, Correct: true},
		},
	}
	got, err := AnalyzeRegret(report, RegretConfig{Algorithm: search.Aegis, RatioThreshold: 1.25, PenaltyFloorNS: 1_000_000, Top: 5})
	if err != nil {
		t.Fatal(err)
	}
	if got.Queries != 2 || got.MeaningfulQueries != 1 {
		t.Fatalf("unexpected counts: %+v", got)
	}
	if got.TopByRatio[0].QueryIndex != 0 || got.TopByRatio[0].Meaningful {
		t.Fatalf("ratio-only microbenchmark noise should not be meaningful: %+v", got.TopByRatio[0])
	}
	if got.TopByPenalty[0].QueryIndex != 1 || !got.TopByPenalty[0].Meaningful {
		t.Fatalf("large absolute penalty should be meaningful: %+v", got.TopByPenalty[0])
	}
	if got.TopByPenalty[0].DistanceRatio < .79 || got.TopByPenalty[0].DistanceRatio > .81 {
		t.Fatalf("distance ratio fallback failed: %+v", got.TopByPenalty[0])
	}
}

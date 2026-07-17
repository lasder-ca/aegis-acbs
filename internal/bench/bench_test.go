package bench

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
	"github.com/lasder-ca/aegis-acbs/internal/search"
)

func TestReportIncludesACBSMetricsAndHTML(t *testing.T) {
	g := graph.New("tiny-road", "fixture", "car", graph.MetricDistance)
	g.Nodes = []graph.Node{
		{ID: 1, Lat: 35.000, Lon: 139.000},
		{ID: 2, Lat: 35.001, Lon: 139.001},
		{ID: 3, Lat: 35.002, Lon: 139.002},
		{ID: 4, Lat: 35.003, Lon: 139.003},
	}
	g.Adj = [][]graph.Edge{
		{{To: 1, Cost: 150_000}},
		{{To: 0, Cost: 150_000}, {To: 2, Cost: 150_000}},
		{{To: 1, Cost: 150_000}, {To: 3, Cost: 150_000}},
		{{To: 2, Cost: 150_000}},
	}
	if err := g.Finalize(); err != nil {
		t.Fatal(err)
	}
	r, err := Run(context.Background(), g, Config{
		Queries: 9, Repeats: 3, BatchSize: 2, Seed: 1010,
		Algorithms: []search.Algorithm{search.Dijkstra, search.BiDijkstra, search.AStar, search.Aegis},
		Suite:      "mixed", PairMode: "strongly-connected",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !r.AllCorrect {
		t.Fatal("expected all algorithms to match")
	}
	if r.Aegis.Comparisons != 9 {
		t.Fatalf("unexpected ACBS summary: %+v", r.Aegis)
	}
	if len(r.ClassSummary) == 0 || len(r.Aegis.Regrets) != 9 || len(r.Aegis.DirectionByClass) == 0 {
		t.Fatal("missing visual summary data")
	}
	path := filepath.Join(t.TempDir(), "report.html")
	if err := WriteHTML(path, r); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	if !strings.Contains(html, "ACBS benchmark") || strings.Contains(html, "__AEGIS_REPORT_BASE64__") {
		t.Fatal("standalone report was not rendered")
	}
}

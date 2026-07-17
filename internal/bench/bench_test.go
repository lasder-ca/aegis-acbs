package bench

import (
	"context"
	"fmt"
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
	if len(r.ClassSummary) == 0 || len(r.Aegis.RuntimeComparisons) != 9 || len(r.Aegis.DirectionByClass) == 0 {
		t.Fatal("missing visual summary data")
	}
	if r.Aegis.MedianOracleRegret < 1 || r.Aegis.RatioOfMediansVsDijkstra <= 0 {
		t.Fatalf("invalid benchmark semantics: %+v", r.Aegis)
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

func TestAggregateDirectoryBuildsMultiSeedMatrix(t *testing.T) {
	g := graph.New("matrix-road", "fixture", "car", graph.MetricDistance)
	g.Nodes = []graph.Node{
		{ID: 1, Lat: 35.000, Lon: 139.000},
		{ID: 2, Lat: 35.001, Lon: 139.001},
		{ID: 3, Lat: 35.002, Lon: 139.002},
		{ID: 4, Lat: 35.003, Lon: 139.003},
	}
	g.Adj = [][]graph.Edge{
		{{To: 1, Cost: 150_000}, {To: 2, Cost: 340_000}},
		{{To: 0, Cost: 150_000}, {To: 2, Cost: 150_000}},
		{{To: 1, Cost: 150_000}, {To: 3, Cost: 150_000}},
		{{To: 2, Cost: 150_000}},
	}
	if err := g.Finalize(); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, seed := range []uint64{1010, 20260717} {
		report, err := Run(context.Background(), g, Config{
			Queries: 9, Repeats: 3, BatchSize: 2, Seed: seed,
			Algorithms: []search.Algorithm{search.Dijkstra, search.BiDijkstra, search.AStar, search.Aegis},
			Suite:      "mixed", PairMode: "strongly-connected",
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := WriteJSON(filepath.Join(dir, fmt.Sprintf("seed-%d.json", seed)), report); err != nil {
			t.Fatal(err)
		}
	}
	matrix, err := AggregateDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if matrix.ReportCount != 2 || len(matrix.Groups) != 1 || !matrix.AllCorrect {
		t.Fatalf("unexpected matrix: %+v", matrix)
	}
	if matrix.Groups[0].Runs != 2 || len(matrix.Groups[0].Seeds) != 2 || matrix.Groups[0].WorstP95OracleRegret < 1 {
		t.Fatalf("unexpected matrix group: %+v", matrix.Groups[0])
	}
	if err := WriteMatrixJSON(filepath.Join(dir, "matrix.json"), matrix); err != nil {
		t.Fatal(err)
	}
	if err := WriteMatrixCSV(filepath.Join(dir, "matrix.csv"), matrix); err != nil {
		t.Fatal(err)
	}
	htmlPath := filepath.Join(dir, "matrix.html")
	if err := WriteMatrixHTML(htmlPath, matrix); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Benchmark Matrix") || strings.Contains(string(data), "__AEGIS_MATRIX_BASE64__") {
		t.Fatal("matrix HTML was not rendered")
	}
}

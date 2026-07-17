package graph

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestOSMImportAndBinaryRoundTrip(t *testing.T) {
	xml := `<?xml version="1.0"?><osm version="0.6"><node id="1" lat="35" lon="139"/><node id="2" lat="35.001" lon="139"/><node id="3" lat="35.002" lon="139"/><way id="9"><nd ref="1"/><nd ref="2"/><nd ref="3"/><tag k="highway" v="residential"/><tag k="oneway" v="yes"/><tag k="maxspeed" v="30"/></way></osm>`
	dir := t.TempDir()
	in := filepath.Join(dir, "x.osm")
	out := filepath.Join(dir, "x.aegis")
	if err := os.WriteFile(in, []byte(xml), 0644); err != nil {
		t.Fatal(err)
	}
	g, err := ImportOSMXML(in, OSMImportOptions{Name: "x", Profile: "car", Metric: MetricTime})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 3 || g.EdgeCount != 2 || !g.Directed {
		t.Fatalf("unexpected graph: nodes=%d edges=%d directed=%v", len(g.Nodes), g.EdgeCount, g.Directed)
	}
	if err := Save(out, g); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.EdgeCount != g.EdgeCount || len(loaded.Nodes) != len(g.Nodes) || loaded.Metric != MetricTime {
		t.Fatal("round trip mismatch")
	}
}

func TestProfiles(t *testing.T) {
	if isRoutable(map[string]string{"highway": "footway"}, "car") {
		t.Fatal("car must not use footway")
	}
	if !isRoutable(map[string]string{"highway": "footway"}, "walk") {
		t.Fatal("walk should use footway")
	}
	if onewayDirection(map[string]string{"oneway": "-1"}, "car") != -1 {
		t.Fatal("reverse oneway")
	}
	if parseMaxspeed("30 mph") < 48 || parseMaxspeed("30 mph") > 49 {
		t.Fatal("mph conversion")
	}
}

func TestRealHatfieldFixture(t *testing.T) {
	path := filepath.Join("..", "..", "benchdata", "hatfield-uk.osm")
	g, err := ImportOSMXML(path, OSMImportOptions{Name: "hatfield", Profile: "car", Metric: MetricDistance})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) < 100 || g.EdgeCount < 180 {
		t.Fatalf("fixture unexpectedly small: %d nodes %d edges", len(g.Nodes), g.EdgeCount)
	}
	if g.MinCostPerMeter <= 0 {
		t.Fatal("missing admissible heuristic bound")
	}
}

func TestUnitVectorIsPrecomputedAndNormalized(t *testing.T) {
	g := New("coords", "", "car", MetricDistance)
	g.Nodes = []Node{{ID: 1, Lat: 35.0, Lon: 139.0}, {ID: 2, Lat: 35.01, Lon: 139.01}}
	g.Adj = [][]Edge{{{To: 1, Cost: 2_000_000}}, {{To: 0, Cost: 2_000_000}}}
	if err := g.Finalize(); err != nil {
		t.Fatal(err)
	}
	for i := range g.Nodes {
		x, y, z := g.UnitVector(i)
		norm := math.Sqrt(x*x + y*y + z*z)
		if math.Abs(norm-1) > 1e-12 {
			t.Fatalf("node %d unit-vector norm=%g", i, norm)
		}
	}
}

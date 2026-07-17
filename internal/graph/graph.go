package graph

import (
	"errors"
	"math"
	"sort"
)

type Metric string

const (
	MetricDistance Metric = "distance"
	MetricTime     Metric = "time"
)

type Node struct {
	ID  int64   `json:"id"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Edge struct {
	To   int    `json:"to"`
	Cost uint64 `json:"cost"`
}

type Graph struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Profile string `json:"profile"`
	Metric  Metric `json:"metric"`
	Nodes   []Node `json:"nodes"`
	// Adj is a construction-time adjacency list. Finalize compacts it into
	// cache-friendly CSR storage and releases the per-node slice headers.
	Adj               [][]Edge `json:"-"`
	outOffsets        []uint32
	outEdges          []Edge
	inOffsets         []uint32
	inEdges           []Edge
	EdgeCount         int     `json:"edgeCount"`
	MinCostPerMeter   float64 `json:"minCostPerMeter"`
	MeanCostPerMeter  float64 `json:"meanCostPerMeter"`
	HeuristicStrength float64 `json:"heuristicStrength"`
	AverageDegree     float64 `json:"averageDegree"`
	DiameterMeters    float64 `json:"diameterMeters"`
	Directed          bool    `json:"directed"`
	idToIndex         map[int64]int
	unitX             []float64
	unitY             []float64
	unitZ             []float64
}

func New(name, source, profile string, metric Metric) *Graph {
	return &Graph{Name: name, Source: source, Profile: profile, Metric: metric, MinCostPerMeter: math.Inf(1)}
}

func (g *Graph) Finalize() error {
	if len(g.Nodes) == 0 {
		return errors.New("graph has no nodes")
	}
	if len(g.Adj) != len(g.Nodes) {
		return errors.New("adjacency length does not match node count")
	}
	g.EdgeCount = 0
	g.MeanCostPerMeter = 0
	g.HeuristicStrength = 0
	g.AverageDegree = 0
	g.DiameterMeters = 0
	g.idToIndex = nil
	g.unitX = make([]float64, len(g.Nodes))
	g.unitY = make([]float64, len(g.Nodes))
	g.unitZ = make([]float64, len(g.Nodes))

	// Validate node IDs without retaining a large hash map on the routing hot
	// path. IndexByID builds the map lazily only for clients that need it.
	ids := make([]int64, len(g.Nodes))
	for i, n := range g.Nodes {
		ids[i] = n.ID
		g.unitX[i], g.unitY[i], g.unitZ[i] = EarthUnit(n.Lat, n.Lon)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1] {
			return errors.New("duplicate node id")
		}
	}

	var ratioSum float64
	var ratioCount uint64
	for from, edges := range g.Adj {
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].To == edges[j].To {
				return edges[i].Cost < edges[j].Cost
			}
			return edges[i].To < edges[j].To
		})
		dedup := edges[:0]
		for _, e := range edges {
			if e.To < 0 || e.To >= len(g.Nodes) || e.Cost == 0 {
				return errors.New("invalid edge")
			}
			if len(dedup) > 0 && dedup[len(dedup)-1].To == e.To {
				if e.Cost < dedup[len(dedup)-1].Cost {
					dedup[len(dedup)-1] = e
				}
				continue
			}
			dedup = append(dedup, e)
		}
		g.Adj[from] = dedup
		g.EdgeCount += len(dedup)
		if uint64(g.EdgeCount) > uint64(^uint32(0)) {
			return errors.New("graph has too many edges for compact storage")
		}
		for _, e := range dedup {
			meters := HaversineMeters(g.Nodes[from].Lat, g.Nodes[from].Lon, g.Nodes[e.To].Lat, g.Nodes[e.To].Lon)
			if meters > 0 {
				ratio := float64(e.Cost) / meters
				if ratio < g.MinCostPerMeter {
					g.MinCostPerMeter = ratio
				}
				ratioSum += ratio
				ratioCount++
			}
		}
	}

	g.outOffsets = make([]uint32, len(g.Nodes)+1)
	g.inOffsets = make([]uint32, len(g.Nodes)+1)
	for from, edges := range g.Adj {
		g.outOffsets[from+1] = g.outOffsets[from] + uint32(len(edges))
		for _, e := range edges {
			g.inOffsets[e.To+1]++
		}
	}
	for i := 1; i < len(g.inOffsets); i++ {
		g.inOffsets[i] += g.inOffsets[i-1]
	}
	g.outEdges = make([]Edge, g.EdgeCount)
	g.inEdges = make([]Edge, g.EdgeCount)
	inCursor := append([]uint32(nil), g.inOffsets[:len(g.Nodes)]...)
	for from, edges := range g.Adj {
		copy(g.outEdges[g.outOffsets[from]:g.outOffsets[from+1]], edges)
		for _, e := range edges {
			at := inCursor[e.To]
			g.inEdges[at] = Edge{To: from, Cost: e.Cost}
			inCursor[e.To]++
		}
	}
	// Drop the n slice headers and their individually allocated backing arrays.
	g.Adj = nil

	if math.IsInf(g.MinCostPerMeter, 1) || g.MinCostPerMeter <= 0 {
		g.MinCostPerMeter = 0
	}
	if ratioCount > 0 {
		g.MeanCostPerMeter = ratioSum / float64(ratioCount)
	}
	if g.MinCostPerMeter > 0 && g.MeanCostPerMeter > 0 {
		g.HeuristicStrength = math.Min(1, g.MinCostPerMeter/g.MeanCostPerMeter)
	}
	g.AverageDegree = float64(g.EdgeCount) / math.Max(1, float64(len(g.Nodes)))
	minLat, minLon, maxLat, maxLon := g.BoundingBox()
	g.DiameterMeters = HaversineMeters(minLat, minLon, maxLat, maxLon)
	g.Directed = detectDirected(g)
	return nil
}

// OutEdges returns the outgoing edges of v from compact CSR storage.
func (g *Graph) OutEdges(v int) []Edge {
	return g.outEdges[g.outOffsets[v]:g.outOffsets[v+1]]
}

// InEdges returns the incoming edges of v from compact reverse CSR storage.
func (g *Graph) InEdges(v int) []Edge {
	return g.inEdges[g.inOffsets[v]:g.inOffsets[v+1]]
}

func (g *Graph) OutDegree(v int) int { return int(g.outOffsets[v+1] - g.outOffsets[v]) }
func (g *Graph) InDegree(v int) int  { return int(g.inOffsets[v+1] - g.inOffsets[v]) }

func detectDirected(g *Graph) bool {
	for from := range g.Nodes {
		for _, e := range g.OutEdges(from) {
			found := false
			for _, back := range g.OutEdges(e.To) {
				if back.To == from {
					found = true
					break
				}
			}
			if !found {
				return true
			}
		}
	}
	return false
}

func (g *Graph) IndexByID(id int64) (int, bool) {
	if g.idToIndex == nil {
		g.idToIndex = make(map[int64]int, len(g.Nodes))
		for i, n := range g.Nodes {
			g.idToIndex[n.ID] = i
		}
	}
	i, ok := g.idToIndex[id]
	return i, ok
}

func (g *Graph) Nearest(lat, lon float64) (int, float64) {
	best := -1
	bestD := math.Inf(1)
	for i, n := range g.Nodes {
		d := HaversineMeters(lat, lon, n.Lat, n.Lon)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best, bestD
}

// EarthUnit converts latitude/longitude to a unit vector on the Earth sphere.
// It is exported so exact search heuristics can reuse the graph's precomputed
// coordinates without repeating trigonometric work for every query.
func EarthUnit(latDeg, lonDeg float64) (x, y, z float64) {
	lat := latDeg * math.Pi / 180
	lon := lonDeg * math.Pi / 180
	cosLat := math.Cos(lat)
	return cosLat * math.Cos(lon), cosLat * math.Sin(lon), math.Sin(lat)
}

// UnitVector returns the precomputed unit-sphere coordinate for node i.
func (g *Graph) UnitVector(i int) (x, y, z float64) {
	if len(g.unitX) != len(g.Nodes) {
		g.unitX = make([]float64, len(g.Nodes))
		g.unitY = make([]float64, len(g.Nodes))
		g.unitZ = make([]float64, len(g.Nodes))
		for j, n := range g.Nodes {
			g.unitX[j], g.unitY[j], g.unitZ[j] = EarthUnit(n.Lat, n.Lon)
		}
	}
	return g.unitX[i], g.unitY[i], g.unitZ[i]
}

func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earth = 6371008.8
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dp := (lat2 - lat1) * math.Pi / 180
	dl := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return earth * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func (g *Graph) BoundingBox() (minLat, minLon, maxLat, maxLon float64) {
	minLat, minLon = math.Inf(1), math.Inf(1)
	maxLat, maxLon = math.Inf(-1), math.Inf(-1)
	for _, n := range g.Nodes {
		if n.Lat < minLat {
			minLat = n.Lat
		}
		if n.Lat > maxLat {
			maxLat = n.Lat
		}
		if n.Lon < minLon {
			minLon = n.Lon
		}
		if n.Lon > maxLon {
			maxLon = n.Lon
		}
	}
	return
}

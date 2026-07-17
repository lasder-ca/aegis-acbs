package graph

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type OSMImportOptions struct {
	Name    string
	Profile string
	Metric  Metric
}

type osmNode struct {
	ID  int64   `xml:"id,attr"`
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

type osmND struct {
	Ref int64 `xml:"ref,attr"`
}
type osmTag struct {
	K string `xml:"k,attr"`
	V string `xml:"v,attr"`
}
type osmWay struct {
	ID   int64    `xml:"id,attr"`
	Refs []osmND  `xml:"nd"`
	Tags []osmTag `xml:"tag"`
}

func ImportOSMXML(path string, opts OSMImportOptions) (*Graph, error) {
	if opts.Profile == "" {
		opts.Profile = "car"
	}
	if opts.Metric == "" {
		opts.Metric = MetricDistance
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := xml.NewDecoder(bufio.NewReaderSize(f, 1<<20))
	nodes := make(map[int64]Node)
	accepted := make([]osmWay, 0, 4096)
	used := make(map[int64]struct{})
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse OSM XML: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "node":
			var n osmNode
			if err := dec.DecodeElement(&n, &se); err != nil {
				return nil, err
			}
			nodes[n.ID] = Node{ID: n.ID, Lat: n.Lat, Lon: n.Lon}
		case "way":
			var w osmWay
			if err := dec.DecodeElement(&w, &se); err != nil {
				return nil, err
			}
			tags := tagMap(w.Tags)
			if isRoutable(tags, opts.Profile) && len(w.Refs) >= 2 {
				accepted = append(accepted, w)
				for _, nd := range w.Refs {
					used[nd.Ref] = struct{}{}
				}
			}
		}
	}
	if len(nodes) == 0 {
		return nil, errors.New("OSM file has no nodes")
	}

	if len(accepted) == 0 {
		return nil, errors.New("OSM file contains no routable ways for selected profile")
	}

	g := New(opts.Name, path, opts.Profile, opts.Metric)
	if g.Name == "" {
		g.Name = strings.TrimSuffix(strings.TrimSuffix(filepathBase(path), ".osm"), ".xml")
	}
	index := make(map[int64]int, len(used))
	ids := make([]int64, 0, len(used))
	for id := range used {
		if _, ok := nodes[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		index[id] = len(g.Nodes)
		g.Nodes = append(g.Nodes, nodes[id])
	}
	g.Adj = make([][]Edge, len(g.Nodes))
	for _, w := range accepted {
		tags := tagMap(w.Tags)
		direction := onewayDirection(tags, opts.Profile)
		speed := speedMetersPerSecond(tags, opts.Profile)
		for i := 0; i+1 < len(w.Refs); i++ {
			a, okA := index[w.Refs[i].Ref]
			b, okB := index[w.Refs[i+1].Ref]
			if !okA || !okB || a == b {
				continue
			}
			meters := HaversineMeters(g.Nodes[a].Lat, g.Nodes[a].Lon, g.Nodes[b].Lat, g.Nodes[b].Lon)
			if meters < 0.01 {
				continue
			}
			var cost uint64
			if opts.Metric == MetricTime {
				cost = uint64(math.Ceil((meters / speed) * 1000))
			} else {
				cost = uint64(math.Ceil(meters * 1000))
			}
			if cost == 0 {
				cost = 1
			}
			switch direction {
			case -1:
				g.Adj[b] = append(g.Adj[b], Edge{To: a, Cost: cost})
			case 1:
				g.Adj[a] = append(g.Adj[a], Edge{To: b, Cost: cost})
			default:
				g.Adj[a] = append(g.Adj[a], Edge{To: b, Cost: cost})
				g.Adj[b] = append(g.Adj[b], Edge{To: a, Cost: cost})
			}
		}
	}
	if err := g.Finalize(); err != nil {
		return nil, err
	}
	return g, nil
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func tagMap(tags []osmTag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[t.K] = t.V
	}
	return m
}

func isRoutable(t map[string]string, profile string) bool {
	h := t["highway"]
	if h == "" || t["area"] == "yes" || t["access"] == "no" || t["access"] == "private" {
		return false
	}
	switch profile {
	case "walk":
		if t["foot"] == "no" {
			return false
		}
		switch h {
		case "motorway", "motorway_link", "trunk", "trunk_link", "construction", "proposed":
			return false
		default:
			return true
		}
	case "bike":
		if t["bicycle"] == "no" {
			return false
		}
		switch h {
		case "motorway", "motorway_link", "construction", "proposed", "steps":
			return false
		default:
			return true
		}
	default:
		if t["motor_vehicle"] == "no" || t["motorcar"] == "no" {
			return false
		}
		switch h {
		case "motorway", "motorway_link", "trunk", "trunk_link", "primary", "primary_link", "secondary", "secondary_link", "tertiary", "tertiary_link", "unclassified", "residential", "living_street", "service", "road":
			return true
		default:
			return false
		}
	}
}

func onewayDirection(t map[string]string, profile string) int {
	if profile == "walk" {
		v := strings.ToLower(t["oneway:foot"])
		if v != "yes" && v != "1" && v != "true" && v != "-1" {
			return 0
		}
	}
	if profile == "bike" && strings.ToLower(t["oneway:bicycle"]) == "no" {
		return 0
	}
	v := strings.ToLower(t["oneway"])
	if v == "-1" {
		return -1
	}
	if v == "yes" || v == "1" || v == "true" || t["junction"] == "roundabout" {
		return 1
	}
	return 0
}

func speedMetersPerSecond(t map[string]string, profile string) float64 {
	if profile == "walk" {
		return 1.4
	}
	if profile == "bike" {
		return 4.5
	}
	if v := parseMaxspeed(t["maxspeed"]); v > 0 {
		return v / 3.6
	}
	kmh := map[string]float64{
		"motorway": 100, "motorway_link": 60, "trunk": 90, "trunk_link": 50,
		"primary": 70, "primary_link": 45, "secondary": 60, "secondary_link": 40,
		"tertiary": 50, "tertiary_link": 35, "unclassified": 40, "residential": 30,
		"living_street": 10, "service": 15, "road": 30,
	}[t["highway"]]
	if kmh == 0 {
		kmh = 30
	}
	return kmh / 3.6
}

func parseMaxspeed(s string) float64 {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "signals" || s == "none" || s == "national" || s == "walk" {
		return 0
	}
	s = strings.Split(s, ";")[0]
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	v, err := strconv.ParseFloat(strings.TrimSuffix(fields[0], ";"), 64)
	if err != nil {
		return 0
	}
	if strings.Contains(s, "mph") {
		return v * 1.609344
	}
	return v
}

package graph

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ImportDIMACS(graphPath, coordPath, name string) (*Graph, error) {
	f, err := os.Open(graphPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	type arc struct {
		from, to int
		cost     uint64
	}
	arcs := make([]arc, 0, 1024)
	n := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 4<<20)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == 'c' {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		switch fs[0] {
		case "p":
			if len(fs) < 4 {
				return nil, fmt.Errorf("invalid DIMACS problem line %d", lineNo)
			}
			n, err = strconv.Atoi(fs[2])
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid node count on line %d", lineNo)
			}
		case "a":
			if len(fs) < 4 {
				return nil, fmt.Errorf("invalid DIMACS arc line %d", lineNo)
			}
			u, e1 := strconv.Atoi(fs[1])
			v, e2 := strconv.Atoi(fs[2])
			c, e3 := strconv.ParseUint(fs[3], 10, 64)
			if e1 != nil || e2 != nil || e3 != nil || c == 0 {
				return nil, fmt.Errorf("invalid arc on line %d", lineNo)
			}
			arcs = append(arcs, arc{u - 1, v - 1, c})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("DIMACS file has no problem line")
	}
	g := New(name, graphPath, "road", MetricDistance)
	if g.Name == "" {
		g.Name = filepathBase(graphPath)
	}
	g.Nodes = make([]Node, n)
	g.Adj = make([][]Edge, n)
	for i := range g.Nodes {
		g.Nodes[i].ID = int64(i + 1)
	}
	for _, a := range arcs {
		if a.from < 0 || a.from >= n || a.to < 0 || a.to >= n {
			return nil, errors.New("arc references out-of-range node")
		}
		g.Adj[a.from] = append(g.Adj[a.from], Edge{To: a.to, Cost: a.cost})
	}
	if coordPath != "" {
		if err := loadDIMACSCoords(coordPath, g.Nodes); err != nil {
			return nil, err
		}
	}
	if err := g.Finalize(); err != nil {
		return nil, err
	}
	return g, nil
}

func loadDIMACSCoords(path string, nodes []Node) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 4<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == 'c' || line[0] == 'p' {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) < 4 || fs[0] != "v" {
			continue
		}
		id, e1 := strconv.Atoi(fs[1])
		x, e2 := strconv.ParseFloat(fs[2], 64)
		y, e3 := strconv.ParseFloat(fs[3], 64)
		if e1 != nil || e2 != nil || e3 != nil || id < 1 || id > len(nodes) {
			return errors.New("invalid DIMACS coordinate")
		}
		if x > 180 || x < -180 || y > 90 || y < -90 {
			x /= 1e6
			y /= 1e6
		}
		nodes[id-1].Lon, nodes[id-1].Lat = x, y
	}
	return sc.Err()
}

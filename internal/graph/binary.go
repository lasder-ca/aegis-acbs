package graph

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var magic = [8]byte{'A', 'E', 'G', 'I', 'S', '1', '2', 0}

func Save(path string, g *Graph) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	defer w.Flush()
	if _, err := w.Write(magic[:]); err != nil {
		return err
	}
	if err := writeString(w, g.Name); err != nil {
		return err
	}
	if err := writeString(w, g.Source); err != nil {
		return err
	}
	if err := writeString(w, g.Profile); err != nil {
		return err
	}
	if err := writeString(w, string(g.Metric)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint64(len(g.Nodes))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint64(g.EdgeCount)); err != nil {
		return err
	}
	for _, n := range g.Nodes {
		if err := binary.Write(w, binary.LittleEndian, n.ID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, n.Lat); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, n.Lon); err != nil {
			return err
		}
	}
	for _, edges := range g.Adj {
		if err := binary.Write(w, binary.LittleEndian, uint32(len(edges))); err != nil {
			return err
		}
		for _, e := range edges {
			if err := binary.Write(w, binary.LittleEndian, uint32(e.To)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, e.Cost); err != nil {
				return err
			}
		}
	}
	return nil
}

func Load(path string) (*Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 1<<20)
	var got [8]byte
	if _, err := io.ReadFull(r, got[:]); err != nil {
		return nil, err
	}
	if got != magic {
		return nil, errors.New("not a supported Aegis graph")
	}
	name, err := readString(r)
	if err != nil {
		return nil, err
	}
	source, err := readString(r)
	if err != nil {
		return nil, err
	}
	profile, err := readString(r)
	if err != nil {
		return nil, err
	}
	metricS, err := readString(r)
	if err != nil {
		return nil, err
	}
	var n, m uint64
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &m); err != nil {
		return nil, err
	}
	if n > 1_000_000_000 || m > 20_000_000_000 {
		return nil, errors.New("graph dimensions are unreasonable")
	}
	g := New(name, source, profile, Metric(metricS))
	g.Nodes = make([]Node, int(n))
	g.Adj = make([][]Edge, int(n))
	for i := range g.Nodes {
		if err := binary.Read(r, binary.LittleEndian, &g.Nodes[i].ID); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &g.Nodes[i].Lat); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &g.Nodes[i].Lon); err != nil {
			return nil, err
		}
	}
	for i := range g.Adj {
		var c uint32
		if err := binary.Read(r, binary.LittleEndian, &c); err != nil {
			return nil, err
		}
		g.Adj[i] = make([]Edge, int(c))
		for j := range g.Adj[i] {
			var to uint32
			if err := binary.Read(r, binary.LittleEndian, &to); err != nil {
				return nil, err
			}
			if err := binary.Read(r, binary.LittleEndian, &g.Adj[i][j].Cost); err != nil {
				return nil, err
			}
			g.Adj[i][j].To = int(to)
		}
	}
	if err := g.Finalize(); err != nil {
		return nil, err
	}
	if uint64(g.EdgeCount) != m {
		return nil, errors.New("edge count mismatch")
	}
	return g, nil
}

func writeString(w io.Writer, s string) error {
	if len(s) > 1<<24 {
		return errors.New("string too large")
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := io.WriteString(w, s)
	return err
}

func readString(r io.Reader) (string, error) {
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	if n > 1<<24 {
		return "", errors.New("string too large")
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

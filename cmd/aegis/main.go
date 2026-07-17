package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/lasder-ca/aegis-acbs/internal/bench"
	"github.com/lasder-ca/aegis-acbs/internal/graph"
	"github.com/lasder-ca/aegis-acbs/internal/i18n"
	"github.com/lasder-ca/aegis-acbs/internal/search"
	"github.com/lasder-ca/aegis-acbs/internal/server"
	"github.com/lasder-ca/aegis-acbs/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "import-osm":
		err = importOSM(os.Args[2:])
	case "import-dimacs":
		err = importDIMACS(os.Args[2:])
	case "route":
		err = route(os.Args[2:])
	case "benchmark":
		err = benchmark(os.Args[2:])
	case "aggregate":
		err = aggregate(os.Args[2:])
	case "serve":
		err = serve(os.Args[2:])
	case "inspect":
		err = inspect(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("%s v%s\n", version.Name, version.Version)
	case "help", "--help", "-h":
		usage()
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf(`%s v%s

Usage:
  aegis import-osm --input city.osm --output city.aegis [--profile car|bike|walk] [--metric distance|time]
  aegis import-dimacs --graph USA-road-d.NY.gr --coords USA-road-d.NY.co --output ny.aegis
  aegis route --graph city.aegis --source ID|lat,lon --target ID|lat,lon [--algorithm aegis]
  aegis benchmark --graph city.aegis --queries 100 --repeats 9 --output report.json --html report.html
  aegis aggregate --input-dir artifacts/benchmarks --output matrix.json --csv matrix.csv --html matrix.html
  aegis serve --graph city.aegis --listen 127.0.0.1:8787
  aegis inspect --graph city.aegis
`, version.Name, version.Version)
}

func importOSM(args []string) error {
	fs := flag.NewFlagSet("import-osm", flag.ContinueOnError)
	in := fs.String("input", "", "OSM XML input")
	out := fs.String("output", "", "Aegis graph output")
	name := fs.String("name", "", "graph name")
	profile := fs.String("profile", "car", "car, bike, or walk")
	metric := fs.String("metric", "distance", "distance or time")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" {
		return errors.New("--input and --output are required")
	}
	if *profile != "car" && *profile != "bike" && *profile != "walk" {
		return errors.New("invalid profile")
	}
	m := graph.Metric(*metric)
	if m != graph.MetricDistance && m != graph.MetricTime {
		return errors.New("invalid metric")
	}
	g, err := graph.ImportOSMXML(*in, graph.OSMImportOptions{Name: *name, Profile: *profile, Metric: m})
	if err != nil {
		return err
	}
	if err := graph.Save(*out, g); err != nil {
		return err
	}
	fmt.Printf("created %s: %d nodes, %d edges\n", *out, len(g.Nodes), g.EdgeCount)
	return nil
}

func importDIMACS(args []string) error {
	fs := flag.NewFlagSet("import-dimacs", flag.ContinueOnError)
	in := fs.String("graph", "", "DIMACS .gr input")
	coords := fs.String("coords", "", "DIMACS .co input")
	out := fs.String("output", "", "Aegis graph output")
	name := fs.String("name", "", "graph name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" {
		return errors.New("--graph and --output are required")
	}
	g, err := graph.ImportDIMACS(*in, *coords, *name)
	if err != nil {
		return err
	}
	if err := graph.Save(*out, g); err != nil {
		return err
	}
	fmt.Printf("created %s: %d nodes, %d edges\n", *out, len(g.Nodes), g.EdgeCount)
	return nil
}

func route(args []string) error {
	fs := flag.NewFlagSet("route", flag.ContinueOnError)
	path := fs.String("graph", "", "Aegis graph")
	src := fs.String("source", "", "node ID or lat,lon")
	dst := fs.String("target", "", "node ID or lat,lon")
	alg := fs.String("algorithm", "aegis", "algorithm")
	langS := fs.String("lang", "en", "ja, en, zh-CN, ko, fr")
	timeout := fs.Duration("timeout", 30*time.Second, "query timeout")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" || *src == "" || *dst == "" {
		return errors.New("--graph, --source, and --target are required")
	}
	g, err := graph.Load(*path)
	if err != nil {
		return err
	}
	s, err := resolve(g, *src)
	if err != nil {
		return err
	}
	t, err := resolve(g, *dst)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	r, err := search.Run(ctx, g, s, t, search.Algorithm(*alg))
	if err != nil {
		return err
	}
	if *jsonOut {
		return json.NewEncoder(os.Stdout).Encode(routeJSON(g, r))
	}
	lang := i18n.Normalize(*langS)
	fmt.Printf("%s: %v\n", i18n.T(lang, "reachable"), r.Stats.Reachable)
	fmt.Printf("%s: %d\n", i18n.T(lang, "distance"), r.Stats.Distance)
	fmt.Printf("%s: %d\n", i18n.T(lang, "expanded"), r.Stats.Expanded)
	fmt.Printf("%s: %d\n", i18n.T(lang, "relaxed"), r.Stats.Relaxed)
	fmt.Printf("%s: %.3f ms\n", i18n.T(lang, "time"), float64(r.Stats.DurationNS)/1e6)
	return nil
}

func routeJSON(g *graph.Graph, r search.Result) map[string]any {
	ids := make([]int64, len(r.Path))
	coords := make([][2]float64, len(r.Path))
	for i, v := range r.Path {
		ids[i] = g.Nodes[v].ID
		coords[i] = [2]float64{g.Nodes[v].Lat, g.Nodes[v].Lon}
	}
	return map[string]any{"stats": r.Stats, "pathNodeIds": ids, "coordinates": coords}
}

func benchmark(args []string) error {
	fs := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	path := fs.String("graph", "", "Aegis graph")
	queries := fs.Int("queries", 100, "query count")
	seed := fs.Uint64("seed", 1010, "deterministic seed")
	out := fs.String("output", "benchmark.json", "JSON report")
	htmlOut := fs.String("html", "benchmark.html", "self-contained visual HTML report; empty disables")
	repeats := fs.Int("repeats", 9, "odd repeated measurements per query and algorithm")
	batchSize := fs.Int("batch", 0, "executions per measurement; 0 selects by graph size")
	order := fs.String("order", "interleaved", "measurement order: interleaved or rotated")
	measureMemory := fs.Bool("measure-memory", false, "run an untimed allocation pass per query and algorithm")
	algs := fs.String("algorithms", "", "comma-separated algorithms; default chooses valid exact algorithms")
	research := fs.Bool("research", false, "include ACBS static-scheduler and no-pruning ablations")
	timeout := fs.Duration("timeout", 30*time.Second, "per-query timeout")
	suite := fs.String("suite", "mixed", "mixed, local, regional, or random")
	pairMode := fs.String("pair-mode", "strongly-connected", "strongly-connected or all")
	cpu := fs.String("cpuprofile", "", "write CPU profile")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("--graph is required")
	}
	if *cpu != "" {
		f, err := os.Create(*cpu)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}
	g, err := graph.Load(*path)
	if err != nil {
		return err
	}
	list := []search.Algorithm{}
	for _, s := range strings.Split(*algs, ",") {
		if s = strings.TrimSpace(s); s != "" {
			list = append(list, search.Algorithm(s))
		}
	}
	if *research && len(list) == 0 {
		list = []search.Algorithm{search.Dijkstra, search.BiDijkstra}
		if g.MinCostPerMeter > 0 {
			list = append(list, search.AStar)
		}
		list = append(list, search.AegisStatic, search.AegisNoPrune, search.Aegis)
	}
	report, err := bench.Run(context.Background(), g, bench.Config{Queries: *queries, Seed: *seed, Algorithms: list, Warmup: 3, Repeats: *repeats, BatchSize: *batchSize, Order: *order, MeasureMemory: *measureMemory, Timeout: *timeout, Suite: *suite, PairMode: *pairMode})
	if err != nil {
		return err
	}
	if err := bench.WriteJSON(*out, report); err != nil {
		return err
	}
	if *htmlOut != "" {
		if err := bench.WriteHTML(*htmlOut, report); err != nil {
			return err
		}
	}
	for _, s := range report.Summary {
		fmt.Printf("%-14s mean=%8.3fms median=%8.3fms best=%8.3fms worst=%8.3fms p95=%8.3fms p99=%8.3fms relaxed=%d expanded=%d alloc=%dB correct=%d/%d\n",
			s.Algorithm, float64(s.MeanNS)/1e6, float64(s.MedianNS)/1e6, float64(s.MinNS)/1e6, float64(s.MaxNS)/1e6, float64(s.P95NS)/1e6, float64(s.P99NS)/1e6,
			s.MedianEdges, s.MedianExpanded, s.MedianAllocBytes, s.Correct, s.Runs)
	}
	if report.Aegis.Comparisons > 0 {
		fmt.Printf("acbs           ratio-of-medians-vs-dijkstra=%.3fx geomean-speedup=%.3fx relative-runtime-to-fastest(p50/p95)=%.3fx/%.3fx oracle-regret(p50/p95)=%.3fx/%.3fx\n",
			report.Aegis.RatioOfMediansVsDijkstra, report.Aegis.GeomeanPerQuerySpeedupVsDijkstra,
			report.Aegis.MedianRelativeRuntimeToFastestBaseline, report.Aegis.P95RelativeRuntimeToFastestBaseline,
			report.Aegis.MedianOracleRegret, report.Aegis.P95OracleRegret)
		fmt.Printf("acbs-work      forward-share=%.1f%% switches=%d chunks=%d pushes=%d pops=%d stale=%d pruned(pop/relax)=%d/%d connections=%d finite-meetings=%d upper-updates=%d\n",
			100*report.Aegis.MedianForwardShare, report.Aegis.MedianDirectionSwitches, report.Aegis.MedianChunks,
			report.Aegis.MedianQueuePushes, report.Aegis.MedianQueuePops, report.Aegis.MedianStalePops,
			report.Aegis.MedianPrunedAtPop, report.Aegis.MedianPrunedAtRelax, report.Aegis.MedianConnectionChecks,
			report.Aegis.MedianFiniteMeetings, report.Aegis.MedianUpperBoundUpdates)
	}
	fmt.Printf("memory         peak-rss=%.2fMiB go-heap=%.2fMiB go-heap-sys=%.2fMiB total-alloc=%.2fMiB gc=%d\n",
		float64(report.Memory.PeakRSSBytes)/(1024*1024), float64(report.Memory.GoHeapAllocBytes)/(1024*1024),
		float64(report.Memory.GoHeapSysBytes)/(1024*1024), float64(report.Memory.GoTotalAllocBytes)/(1024*1024), report.Memory.GoNumGC)
	if !report.AllCorrect {
		return errors.New("correctness mismatch detected")
	}
	fmt.Println("report:", *out)
	if *htmlOut != "" {
		fmt.Println("visual report:", *htmlOut)
	}
	return nil
}

func aggregate(args []string) error {
	fs := flag.NewFlagSet("aggregate", flag.ContinueOnError)
	inputDir := fs.String("input-dir", "", "directory containing benchmark JSON reports")
	output := fs.String("output", "benchmark-matrix.json", "aggregate JSON output")
	csvOut := fs.String("csv", "benchmark-matrix.csv", "aggregate CSV output; empty disables")
	htmlOut := fs.String("html", "benchmark-matrix.html", "self-contained aggregate HTML output; empty disables")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *inputDir == "" {
		return errors.New("--input-dir is required")
	}
	report, err := bench.AggregateDirectory(*inputDir)
	if err != nil {
		return err
	}
	if err := bench.WriteMatrixJSON(*output, report); err != nil {
		return err
	}
	if *csvOut != "" {
		if err := bench.WriteMatrixCSV(*csvOut, report); err != nil {
			return err
		}
	}
	if *htmlOut != "" {
		if err := bench.WriteMatrixHTML(*htmlOut, report); err != nil {
			return err
		}
	}
	for _, group := range report.Groups {
		fmt.Printf("%-20s %-8s runs=%d p50=%8.3fms median-p95=%8.3fms speedup=%.3fx worst-p95-regret=%.3fx correct=%v\n",
			group.GraphName, group.Metric, group.Runs, float64(group.MedianOfAegisMediansNS)/1e6,
			float64(group.MedianOfAegisP95NS)/1e6, group.GeomeanPerQuerySpeedupVsDijkstra,
			group.WorstP95OracleRegret, group.AllCorrect)
	}
	fmt.Println("matrix report:", *output)
	if *csvOut != "" {
		fmt.Println("matrix csv:", *csvOut)
	}
	if *htmlOut != "" {
		fmt.Println("matrix visual report:", *htmlOut)
	}
	if !report.AllCorrect {
		return errors.New("one or more benchmark reports contain correctness mismatches")
	}
	return nil
}

func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	path := fs.String("graph", "", "Aegis graph")
	listen := fs.String("listen", "127.0.0.1:8787", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("--graph is required")
	}
	g, err := graph.Load(*path)
	if err != nil {
		return err
	}
	srv := &http.Server{Addr: *listen, Handler: server.App{Graph: g}.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 5 * time.Minute, IdleTimeout: 60 * time.Second}
	log.Printf("Aegis ACBS v%s: http://%s", version.Version, *listen)
	return srv.ListenAndServe()
}

func inspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	path := fs.String("graph", "", "Aegis graph")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("--graph is required")
	}
	g, err := graph.Load(*path)
	if err != nil {
		return err
	}
	minLat, minLon, maxLat, maxLon := g.BoundingBox()
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"name": g.Name, "source": g.Source, "profile": g.Profile, "metric": g.Metric, "nodes": len(g.Nodes), "edges": g.EdgeCount, "directed": g.Directed, "minCostPerMeter": g.MinCostPerMeter, "meanCostPerMeter": g.MeanCostPerMeter, "heuristicStrength": g.HeuristicStrength, "averageDegree": g.AverageDegree, "diameterMeters": g.DiameterMeters, "bbox": []float64{minLat, minLon, maxLat, maxLon}})
}

func resolve(g *graph.Graph, s string) (int, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") {
		p := strings.Split(s, ",")
		if len(p) != 2 {
			return -1, errors.New("invalid coordinates")
		}
		lat, e1 := strconv.ParseFloat(strings.TrimSpace(p[0]), 64)
		lon, e2 := strconv.ParseFloat(strings.TrimSpace(p[1]), 64)
		if e1 != nil || e2 != nil {
			return -1, errors.New("invalid coordinates")
		}
		i, _ := g.Nearest(lat, lon)
		return i, nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1, errors.New("expected node ID or lat,lon")
	}
	if i, ok := g.IndexByID(id); ok {
		return i, nil
	}
	return -1, errors.New("node not found")
}

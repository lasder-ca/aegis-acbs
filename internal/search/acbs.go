package search

import (
	"context"
	"math"

	"github.com/lasder-ca/aegis-acbs/internal/graph"
)

const (
	acbsSchedulerVersion    = "edge-efficiency-v3"
	acbsChordPotentialModel = "balanced-chord-v3"
	acbsProjectionModel     = "balanced-projection-v1"
)

type acbsOptions struct {
	algorithm  Algorithm
	adaptive   bool
	pruning    bool
	projection bool
}

func acbs(ctx context.Context, g *graph.Graph, source, target int) (Result, error) {
	// The incumbent-bound pruning ablation was inactive on the large road
	// benchmarks and added extra bound evaluation work. The production ACBS
	// path therefore keeps the exact coupled-bound termination rule but does
	// not run the optional per-node incumbent pruning experiment.
	return acbsWithOptions(ctx, g, source, target, acbsOptions{algorithm: Aegis, adaptive: true, pruning: false})
}

func acbsStatic(ctx context.Context, g *graph.Graph, source, target int) (Result, error) {
	// Match production ACBS except for the scheduler, so this ablation isolates
	// the adaptive edge-efficiency scheduler without a pruning confounder.
	return acbsWithOptions(ctx, g, source, target, acbsOptions{algorithm: AegisStatic, adaptive: false, pruning: false})
}

func acbsPrune(ctx context.Context, g *graph.Graph, source, target int) (Result, error) {
	return acbsWithOptions(ctx, g, source, target, acbsOptions{algorithm: AegisPrune, adaptive: true, pruning: true})
}

func acbsNoPrune(ctx context.Context, g *graph.Graph, source, target int) (Result, error) {
	// Backward-compatible alias retained for old benchmark command lines.
	r, err := acbsWithOptions(ctx, g, source, target, acbsOptions{algorithm: AegisNoPrune, adaptive: true, pruning: false})
	return r, err
}

func acbsProjection(ctx context.Context, g *graph.Graph, source, target int) (Result, error) {
	// Experimental potential ablation only. It intentionally shares the same
	// no-pruning production search path so the potential is the sole variable.
	return acbsWithOptions(ctx, g, source, target, acbsOptions{algorithm: AegisProjection, adaptive: true, pruning: false, projection: true})
}

// ACBS implements Aegis Coupled-Bound Search.
//
// ACBS is one exact bidirectional search. Both directions run on non-negative
// reduced costs induced by a balanced feasible potential. The scheduler only
// decides which frontier receives the next edge-work budget; correctness is
// controlled exclusively by the incumbent upper bound and the coupled lower
// bound. Optional per-node incumbent pruning is retained only as an explicit
// experimental ablation because it did not activate on the large road runs.
func acbsWithOptions(ctx context.Context, g *graph.Graph, source, target int, opts acbsOptions) (Result, error) {
	modelName := acbsChordPotentialModel
	if opts.projection {
		modelName = acbsProjectionModel
	}
	if source == target {
		return Result{Path: []int{source}, Stats: Stats{
			Algorithm: opts.algorithm, Reachable: true, PathNodes: 1,
			SchedulerVersion: schedulerName(opts), PotentialModel: modelName,
		}}, nil
	}

	n := len(g.Nodes)
	w := acquireBiWorkspace(n)
	defer releaseBiWorkspace(w)
	df, db := w.df, w.db
	pf, pb := w.pf, w.pb
	settledF, settledB := w.settledF, w.settledB

	potential := newACBSPotential(g, source, target, opts.projection)
	phiS, freshS := w.potential(g, potential, source)
	phiT, freshT := w.potential(g, potential, target)

	w.touchForward(source)
	w.touchBackward(target)
	df[source], db[target] = 0, 0
	qf, qb := &w.qf, &w.qb
	radixPush(qf, item{node: source, distance: 0, priority: reducedForwardKey(0, phiS, phiS)})
	radixPush(qb, item{node: target, distance: 0, priority: reducedBackwardKey(0, phiT, phiT)})

	stats := Stats{
		Algorithm: opts.algorithm, QueuePushes: 2,
		SchedulerVersion: schedulerName(opts), PotentialModel: modelName,
	}
	if freshS {
		stats.PotentialEvaluations++
	}
	if freshT && target != source {
		stats.PotentialEvaluations++
	}

	best, meet := inf, -1
	bestReduced := inf
	trace := acbsTraceFromContext(ctx)
	var scoreF, scoreB uint64
	var sampledF, sampledB bool
	lastDirection := byte(0)
	consecutive := 0
	terminatedByBound := false

	for {
		frontF, okF := peekValid(qf, df, settledF, &stats)
		frontB, okB := peekValid(qb, db, settledB, &stats)
		if !okF || !okB {
			break
		}
		lowerBound := saturatingAdd(frontF.priority, frontB.priority)
		if bestReduced != inf && lowerBound >= bestReduced {
			stats.TerminationLowerBound = reducedToOriginalLowerBound(lowerBound, phiS, phiT)
			if stats.TerminationLowerBound > best {
				stats.TerminationLowerBound = best
			}
			terminatedByBound = true
			break
		}
		if stats.Expanded&1023 == 0 {
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			default:
			}
		}

		direction := byte(0)
		if opts.adaptive {
			direction = chooseACBSDirection(
				g, frontF, frontB, qf.Len(), qb.Len(), scoreF, scoreB,
				sampledF, sampledB, lastDirection, consecutive,
			)
		} else {
			direction = chooseACBSStaticDirection(g, frontF, frontB, qf.Len(), qb.Len())
		}
		if direction != lastDirection {
			if lastDirection != 0 {
				stats.DirectionSwitches++
			}
			lastDirection = direction
			consecutive = 1
		} else {
			consecutive++
		}

		budget := acbsEdgeBudget(g.EdgeCount, scoreF, scoreB, direction, bestReduced != inf)
		if !opts.adaptive {
			budget = acbsBaseEdgeBudget(g.EdgeCount)
		}
		beforeLB := lowerBound
		beforeRelaxed := stats.Relaxed
		beforeExpanded := stats.Expanded
		beforeQueues := qf.Len() + qb.Len()
		beforeQF, beforeQB := qf.Len(), qb.Len()
		beforeScoreF, beforeScoreB := scoreF, scoreB
		beforeBest := best
		stats.Chunks++

		for used := 0; used < budget; {
			frontF, okF = peekValid(qf, df, settledF, &stats)
			frontB, okB = peekValid(qb, db, settledB, &stats)
			if !okF || !okB {
				break
			}
			lowerBound = saturatingAdd(frontF.priority, frontB.priority)
			if bestReduced != inf && lowerBound >= bestReduced {
				break
			}

			if direction == 'F' {
				cur := radixPop(qf)
				stats.QueuePops++
				if cur.distance != df[cur.node] || settledF[cur.node] {
					stats.StalePops++
					continue
				}
				settledF[cur.node] = true
				updateACBSBest(cur.node, df, db, &best, &bestReduced, &meet, &stats, phiS, phiT)
				hForward := uint64(0)
				if opts.pruning && best != inf {
					var fresh bool
					hForward, _, fresh = w.heuristicBounds(g, potential, cur.node)
					if fresh {
						stats.BoundEvaluations++
					}
				}
				if opts.pruning && best != inf && boundCannotImprove(df[cur.node], hForward, best) {
					stats.PrunedAtPop++
					stats.BoundPruned++
					used++
					continue
				}
				edges := g.OutEdges(cur.node)
				used += max(1, len(edges))
				stats.Expanded++
				stats.ForwardExpanded++
				for _, e := range edges {
					stats.Relaxed++
					if df[cur.node] > inf-e.Cost {
						continue
					}
					nd := df[cur.node] + e.Cost
					if nd < df[e.To] {
						phi, fresh := w.potential(g, potential, e.To)
						if fresh {
							stats.PotentialEvaluations++
						}
						hf := uint64(0)
						if opts.pruning && best != inf {
							var boundFresh bool
							hf, _, boundFresh = w.heuristicBounds(g, potential, e.To)
							if boundFresh {
								stats.BoundEvaluations++
							}
						}
						if opts.pruning && best != inf && boundCannotImprove(nd, hf, best) {
							stats.PrunedAtRelax++
							stats.BoundPruned++
						} else {
							w.touchForward(e.To)
							df[e.To] = nd
							pf[e.To] = int32(cur.node)
							radixPush(qf, item{node: e.To, distance: nd, priority: reducedForwardKey(nd, phi, phiS)})
							stats.QueuePushes++
						}
					}
					if db[e.To] != inf {
						updateACBSBest(e.To, df, db, &best, &bestReduced, &meet, &stats, phiS, phiT)
					}
				}
			} else {
				cur := radixPop(qb)
				stats.QueuePops++
				if cur.distance != db[cur.node] || settledB[cur.node] {
					stats.StalePops++
					continue
				}
				settledB[cur.node] = true
				updateACBSBest(cur.node, df, db, &best, &bestReduced, &meet, &stats, phiS, phiT)
				hBackward := uint64(0)
				if opts.pruning && best != inf {
					var fresh bool
					_, hBackward, fresh = w.heuristicBounds(g, potential, cur.node)
					if fresh {
						stats.BoundEvaluations++
					}
				}
				if opts.pruning && best != inf && boundCannotImprove(db[cur.node], hBackward, best) {
					stats.PrunedAtPop++
					stats.BoundPruned++
					used++
					continue
				}
				edges := g.InEdges(cur.node)
				used += max(1, len(edges))
				stats.Expanded++
				stats.BackwardExpanded++
				for _, e := range edges {
					stats.Relaxed++
					if db[cur.node] > inf-e.Cost {
						continue
					}
					nd := db[cur.node] + e.Cost
					if nd < db[e.To] {
						phi, fresh := w.potential(g, potential, e.To)
						if fresh {
							stats.PotentialEvaluations++
						}
						hb := uint64(0)
						if opts.pruning && best != inf {
							var boundFresh bool
							_, hb, boundFresh = w.heuristicBounds(g, potential, e.To)
							if boundFresh {
								stats.BoundEvaluations++
							}
						}
						if opts.pruning && best != inf && boundCannotImprove(nd, hb, best) {
							stats.PrunedAtRelax++
							stats.BoundPruned++
						} else {
							w.touchBackward(e.To)
							db[e.To] = nd
							pb[e.To] = int32(cur.node)
							radixPush(qb, item{node: e.To, distance: nd, priority: reducedBackwardKey(nd, phi, phiT)})
							stats.QueuePushes++
						}
					}
					if df[e.To] != inf {
						updateACBSBest(e.To, df, db, &best, &bestReduced, &meet, &stats, phiS, phiT)
					}
				}
			}
		}

		frontF, okF = peekValid(qf, df, settledF, &stats)
		frontB, okB = peekValid(qb, db, settledB, &stats)
		afterLB := beforeLB
		if okF && okB {
			afterLB = saturatingAdd(frontF.priority, frontB.priority)
		}
		gain := uint64(0)
		if afterLB > beforeLB {
			gain = afterLB - beforeLB
		}
		work := schedulerWork(
			stats.Relaxed-beforeRelaxed,
			stats.Expanded-beforeExpanded,
			qf.Len()+qb.Len()-beforeQueues,
		)
		if opts.adaptive {
			instant := efficiencyScore(gain, work)
			if direction == 'F' {
				scoreF = emaScore(scoreF, instant, sampledF)
				sampledF = true
			} else {
				scoreB = emaScore(scoreB, instant, sampledB)
				sampledB = true
			}
		}
		if trace != nil {
			event := ACBSTraceEvent{
				Chunk: stats.Chunks, Direction: string([]byte{direction}), Budget: budget,
				BeforeLowerBound: beforeLB, AfterLowerBound: afterLB, LowerBoundGain: gain, Work: work,
				RelaxedDelta: stats.Relaxed - beforeRelaxed, ExpandedDelta: stats.Expanded - beforeExpanded,
				ForwardQueueBefore: beforeQF, BackwardQueueBefore: beforeQB,
				ForwardQueueAfter: qf.Len(), BackwardQueueAfter: qb.Len(),
				ForwardScoreBefore:  float64(beforeScoreF) / 1_000_000.0,
				BackwardScoreBefore: float64(beforeScoreB) / 1_000_000.0,
				ForwardScoreAfter:   float64(scoreF) / 1_000_000.0,
				BackwardScoreAfter:  float64(scoreB) / 1_000_000.0,
				HadUpperBoundBefore: beforeBest != inf, HadUpperBoundAfter: best != inf,
			}
			if beforeBest != inf {
				event.UpperBoundBefore = beforeBest
			}
			if best != inf {
				event.UpperBoundAfter = best
			}
			trace(event)
		}
	}

	stats.ForwardEfficiency = float64(scoreF) / 1_000_000.0
	stats.BackwardEfficiency = float64(scoreB) / 1_000_000.0
	if best == inf || meet < 0 {
		return Result{Stats: stats}, nil
	}
	if !terminatedByBound {
		stats.TerminationLowerBound = best
	}
	stats.UpperBound = best
	stats.LowerBound = stats.TerminationLowerBound
	if stats.LowerBound < stats.UpperBound {
		stats.OptimalityGap = stats.UpperBound - stats.LowerBound
	}

	path := reconstructBidirectional(pf, pb, source, meet, target)
	if len(path) == 0 {
		return Result{Stats: stats}, nil
	}
	stats.Distance = best
	stats.Reachable = true
	stats.PathNodes = len(path)
	return Result{Path: path, Stats: stats}, nil
}

type acbsPotential struct {
	sourceX, sourceY, sourceZ float64
	targetX, targetY, targetZ float64
	dirX, dirY, dirZ          float64
	projectionScale           float64
	costPerMeter              float64
	projection                bool
	enabled                   bool
}

func newACBSPotential(g *graph.Graph, source, target int, projection bool) acbsPotential {
	if g.MinCostPerMeter <= 0 {
		return acbsPotential{}
	}
	sx, sy, sz := g.UnitVector(source)
	tx, ty, tz := g.UnitVector(target)
	p := acbsPotential{
		sourceX: sx, sourceY: sy, sourceZ: sz,
		targetX: tx, targetY: ty, targetZ: tz,
		costPerMeter: g.MinCostPerMeter * (1 - 1e-12),
		projection:   projection,
		enabled:      true,
	}
	if !projection {
		return p
	}
	dx, dy, dz := tx-sx, ty-sy, tz-sz
	norm := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if norm <= 0 {
		return acbsPotential{}
	}
	const earthRadiusMeters = 6371008.8
	p.dirX, p.dirY, p.dirZ = dx/norm, dy/norm, dz/norm
	p.costPerMeter = g.MinCostPerMeter * (1 - 1e-9)
	p.projectionScale = 2 * earthRadiusMeters * p.costPerMeter
	return p
}

func (p acbsPotential) phi(g *graph.Graph, v int) int64 {
	if !p.enabled {
		return 0
	}
	x, y, z := g.UnitVector(v)
	if p.projection {
		value := p.projectionScale * (x*p.dirX + y*p.dirY + z*p.dirZ)
		const limit = float64(math.MaxInt64 / 4)
		if value >= limit {
			return math.MaxInt64 / 4
		}
		if value <= -limit {
			return -math.MaxInt64 / 4
		}
		return int64(value)
	}
	forward := lowerBoundCost(chordUnitMeters(x, y, z, p.targetX, p.targetY, p.targetZ), p.costPerMeter)
	backward := lowerBoundCost(chordUnitMeters(x, y, z, p.sourceX, p.sourceY, p.sourceZ), p.costPerMeter)
	return signedDifference(forward, backward)
}

func (p acbsPotential) bounds(g *graph.Graph, v int) (forward, backward uint64) {
	if !p.enabled {
		return 0, 0
	}
	x, y, z := g.UnitVector(v)
	forward = lowerBoundCost(chordUnitMeters(x, y, z, p.targetX, p.targetY, p.targetZ), p.costPerMeter)
	backward = lowerBoundCost(chordUnitMeters(x, y, z, p.sourceX, p.sourceY, p.sourceZ), p.costPerMeter)
	return forward, backward
}

func chordUnitMeters(ax, ay, az, bx, by, bz float64) float64 {
	const earthRadiusMeters = 6371008.8
	dx, dy, dz := ax-bx, ay-by, az-bz
	return earthRadiusMeters * math.Sqrt(dx*dx+dy*dy+dz*dz)
}

func lowerBoundCost(meters, costPerMeter float64) uint64 {
	if meters <= 0 || costPerMeter <= 0 {
		return 0
	}
	v := math.Floor(meters * costPerMeter)
	if v >= float64(math.MaxInt64/4) {
		return uint64(math.MaxInt64 / 4)
	}
	return uint64(v)
}

func signedDifference(a, b uint64) int64 {
	const limit = uint64(math.MaxInt64 / 4)
	if a >= b {
		d := a - b
		if d > limit {
			return math.MaxInt64 / 4
		}
		return int64(d)
	}
	d := b - a
	if d > limit {
		return -math.MaxInt64 / 4
	}
	return -int64(d)
}

func updateACBSBest(node int, df, db []uint64, best, bestReduced *uint64, meet *int, stats *Stats, phiS, phiT int64) {
	stats.ConnectionChecks++
	if df[node] == inf || db[node] == inf || df[node] > inf-db[node] {
		return
	}
	stats.FiniteMeetings++
	// MeetingChecks is retained for JSON compatibility with v0.3 reports.
	// From v0.4 onward it means a finite forward/backward overlap, while
	// ConnectionChecks records every attempted connection test.
	stats.MeetingChecks++
	candidate := df[node] + db[node]
	if candidate >= *best {
		return
	}
	*best = candidate
	*meet = node
	*bestReduced = originalToReducedUpperBound(candidate, phiS, phiT)
	if stats.UpperBoundUpdates == 0 {
		stats.FirstUpperBoundExpanded = stats.Expanded
	}
	stats.UpperBoundUpdates++
}

func boundCannotImprove(gCost, heuristic, incumbent uint64) bool {
	if gCost >= incumbent {
		return true
	}
	return heuristic >= incumbent-gCost
}

func peekValid(q *radixHeap, dist []uint64, settled []bool, stats *Stats) (item, bool) {
	for q.Len() > 0 {
		cur, _ := radixPeek(q)
		if cur.distance == dist[cur.node] && !settled[cur.node] {
			return cur, true
		}
		radixPop(q)
		stats.QueuePops++
		stats.StalePops++
	}
	return item{}, false
}

func schedulerName(opts acbsOptions) string {
	if !opts.adaptive {
		return "lower-key-static-v2"
	}
	if opts.pruning {
		return acbsSchedulerVersion + "-incumbent-prune"
	}
	return acbsSchedulerVersion
}

func chooseACBSStaticDirection(g *graph.Graph, frontF, frontB item, lenF, lenB int) byte {
	if frontF.priority < frontB.priority {
		return 'F'
	}
	if frontB.priority < frontF.priority {
		return 'B'
	}
	if g.OutDegree(frontF.node) < g.InDegree(frontB.node) {
		return 'F'
	}
	if g.InDegree(frontB.node) < g.OutDegree(frontF.node) {
		return 'B'
	}
	if lenF <= lenB {
		return 'F'
	}
	return 'B'
}

func chooseACBSDirection(g *graph.Graph, frontF, frontB item, lenF, lenB int, scoreF, scoreB uint64, sampledF, sampledB bool, last byte, consecutive int) byte {
	if !sampledF {
		return 'F'
	}
	if !sampledB {
		return 'B'
	}
	// Periodically refresh both empirical efficiency estimates. This affects
	// scheduling only; it is not required for exactness.
	if consecutive >= 6 {
		if last == 'F' {
			return 'B'
		}
		return 'F'
	}
	if scoreF > scoreB+scoreB/10 {
		return 'F'
	}
	if scoreB > scoreF+scoreF/10 {
		return 'B'
	}
	if frontF.priority < frontB.priority {
		return 'F'
	}
	if frontB.priority < frontF.priority {
		return 'B'
	}
	degreeF := g.OutDegree(frontF.node)
	degreeB := g.InDegree(frontB.node)
	if degreeF < degreeB {
		return 'F'
	}
	if degreeB < degreeF {
		return 'B'
	}
	if lenF <= lenB {
		return 'F'
	}
	return 'B'
}

func acbsBaseEdgeBudget(edgeCount int) int {
	if edgeCount < 10_000 {
		return 256
	}
	if edgeCount < 100_000 {
		return 1024
	}
	return 2048
}

func acbsEdgeBudget(edgeCount int, scoreF, scoreB uint64, direction byte, hasUpperBound bool) int {
	base := acbsBaseEdgeBudget(edgeCount)
	chosen, other := scoreF, scoreB
	if direction == 'B' {
		chosen, other = scoreB, scoreF
	}
	budget := base
	if other > 0 && chosen >= other*2 {
		budget = base * 2
	}
	if other > 0 && chosen >= other*4 {
		budget = base * 4
	}
	if hasUpperBound && budget > base*2 {
		budget = base * 2
	}
	return budget
}

func schedulerWork(relaxed, expanded uint64, queueGrowth int) uint64 {
	work := relaxed + expanded*4
	if queueGrowth > 0 {
		work = saturatingAdd(work, uint64(queueGrowth)*2)
	}
	if work == 0 {
		return 1
	}
	return work
}

func efficiencyScore(gain, work uint64) uint64 {
	if work == 0 {
		return 0
	}
	if gain > math.MaxUint64/1_000_000 {
		return math.MaxUint64
	}
	return gain * 1_000_000 / work
}

func emaScore(previous, current uint64, initialized bool) uint64 {
	if !initialized {
		return current
	}
	return previous - previous/4 + current/4
}

func reducedForwardKey(distance uint64, phi, phiSource int64) uint64 {
	return doubledPlusSigned(distance, phi-phiSource)
}

func reducedBackwardKey(distance uint64, phi, phiTarget int64) uint64 {
	return doubledPlusSigned(distance, phiTarget-phi)
}

func doubledPlusSigned(distance uint64, delta int64) uint64 {
	if distance > math.MaxUint64/2 {
		return inf
	}
	base := distance * 2
	if delta >= 0 {
		return saturatingAdd(base, uint64(delta))
	}
	negative := uint64(-(delta + 1)) + 1
	if negative >= base {
		return 0
	}
	return base - negative
}

func originalToReducedUpperBound(original uint64, phiSource, phiTarget int64) uint64 {
	return doubledPlusSigned(original, phiTarget-phiSource)
}

func reducedToOriginalLowerBound(reduced uint64, phiSource, phiTarget int64) uint64 {
	shift := phiTarget - phiSource
	adjusted := reduced
	if shift >= 0 {
		s := uint64(shift)
		if s >= adjusted {
			return 0
		}
		adjusted -= s
	} else {
		adjusted = saturatingAdd(adjusted, uint64(-(shift+1))+1)
	}
	return adjusted / 2
}

func saturatingAdd(a, b uint64) uint64 {
	if a > math.MaxUint64-b {
		return math.MaxUint64
	}
	return a + b
}

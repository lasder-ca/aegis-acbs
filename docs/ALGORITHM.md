# Aegis Coupled-Bound Search

## Scope

Aegis Coupled-Bound Search (ACBS) computes an exact point-to-point shortest path on a finite directed graph with non-negative integer edge weights. Road graphs use both the original adjacency structure and a reverse adjacency structure.

The implementation advances two search frontiers inside one algorithm:

- a forward frontier from the source,
- a reverse frontier from the target,
- a shared incumbent upper bound,
- and a shared admissible lower bound used for termination.

The adaptive component controls exploration order only. Exactness depends on the feasible potential, the lower and upper bounds, and the stopping rule.

## State

For source `s`, target `t`, and node `v`:

- `OPEN_F`: forward priority queue,
- `OPEN_B`: reverse priority queue,
- `g_F(v)`: best known cost from `s` to `v`,
- `g_B(v)`: best known cost from `v` to `t`,
- `U`: cost of the best complete path found so far,
- `L`: lower bound on every not-yet-certified path.

## Geographic lower bounds

At graph finalization, each latitude/longitude coordinate is mapped to a point on the unit sphere:

```text
q(v) = (x(v), y(v), z(v))
```

The Earth-chord distance is:

```text
chord(u, v) = R ||q(u) - q(v)||₂
```

Chord distance is no greater than great-circle distance and satisfies the triangle inequality. Multiplying it by a conservatively derived minimum cost per meter gives admissible directional bounds:

```text
h_F(v) = floor(chord(v, t) × min_cost_per_meter)
h_B(v) = floor(chord(s, v) × min_cost_per_meter)
```

When coordinates or a safe cost-per-meter lower bound are unavailable, both heuristic terms become zero.

## Balanced feasible potential

ACBS uses the doubled potential:

```text
φ₂(v) = h_F(v) - h_B(v)
```

Forward and reverse queue keys are:

```text
k_F(v) = 2g_F(v) + φ₂(v) - φ₂(s)
k_B(v) = 2g_B(v) + φ₂(t) - φ₂(v)
```

Consistency of the directional bounds keeps reduced edge costs non-negative in both directions. This allows each frontier to use Dijkstra-style label setting on its reduced-cost graph.

## Coupled lower bound and incumbent

Let `min_F` and `min_B` be the valid minimum keys in the two queues. The coupled lower bound is:

```text
L₂ = min_F + min_B
```

Whenever both searches have finite labels for a connection node `m`, the incumbent is updated:

```text
U = min(U, g_F(m) + g_B(m))
U₂ = 2U + φ₂(t) - φ₂(s)
```

The search terminates when:

```text
L₂ >= U₂
```

At that point no undiscovered path can improve the incumbent. Reports expose `upperBound`, `lowerBound`, and `optimalityGap`; a certified result has an optimality gap of zero.

## Adaptive edge-work scheduling

Each direction is processed in chunks measured by examined edges rather than only by expanded nodes. The scheduler estimates how efficiently a direction has raised the coupled lower bound:

```text
efficiency = lower_bound_gain
             / (relaxed_edges
                + 4 × expanded_nodes
                + 2 × positive_queue_growth)
```

The estimates are smoothed and used with the current queue minima, next-node degree, and queue size. The implementation also forces periodic sampling of the opposite direction to avoid starving one frontier.

Typical edge budgets range from 256 to 8192 examined edges. After an incumbent is found, the maximum chunk is reduced so the stopping condition is checked more frequently.

The scheduler never rewrites:

- reduced edge costs,
- `g` labels,
- the incumbent `U`,
- the coupled lower bound `L`,
- or the termination condition.

## Queue and graph representation

The finalized graph uses CSR-style arrays for both directions:

```text
outOffsets, outEdges
inOffsets,  inEdges
```

The two priority queues use monotone radix heaps because valid reduced keys are non-decreasing. Search workspaces and queue backing arrays are pooled between requests; returned paths are exact-sized and owned by the caller.

## Variants used in evaluation

- `aegis`: adaptive edge-work scheduler with the balanced chord potential.
- `aegis-static`: the same search and stopping proof with a fixed direction scheduler.
- `aegis-prune`: an experimental incumbent-bound pruning variant.
- `aegis-projection`: an experimental linear-projection potential.

Previously rejected scheduler guards remain in the codebase only so published experiments can be reproduced. They are excluded from the normal benchmark set.

## Complexity

The worst-case time complexity is:

```text
O((V + E) log V)
```

The per-query search workspace is `O(V)` excluding the graph. Geographic unit vectors are prepared once when the graph is finalized; directional heuristic values are cached only for touched nodes during a query.

## Status

ACBS combines established shortest-path components with an adaptive coupled-bound scheduling strategy. The repository provides implementation details, differential tests, and reproducible experiments. Independent review is still required to determine novelty and performance generalization.

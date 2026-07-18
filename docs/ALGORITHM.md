# Aegis Coupled-Bound Search v3

## Scope

ACBS v2は、有限・非負整数重みの有向グラフ上で、厳密な1対1最短経路を求める。道路グラフでは元の隣接リストと、各有向辺を反転した逆隣接リストを使用する。

ACBSはA*、Dijkstra、双方向Dijkstraのいずれかを選択するポートフォリオではない。前向きと後ろ向きの探索状態を1つの停止証明で結合する、単一の双方向探索である。

## State

- `OPEN_F`: 始点から進む前向きfrontier
- `OPEN_B`: 終点から逆辺を進む後ろ向きfrontier
- `g_F(v)`: 始点から`v`までの既知最小コスト
- `g_B(v)`: `v`から終点までの既知最小コスト
- `U`: 発見済み完全経路の最小コスト
- `L`: 未発見経路に対する共通下界

## Chord lower bounds

各緯度経度を地球単位球面上の3次元ベクトルへ、グラフ読み込み時に一度だけ変換する。

```text
q(v) = (x(v), y(v), z(v))
```

2点間の弦距離を次で求める。

```text
chord(u,v) = R ||q(u)-q(v)||₂
```

弦距離は大円距離以下であり、3次元ユークリッド距離なので三角不等式を満たす。グラフ全辺から求めた安全な`min_cost_per_meter`を掛け、切り捨てる。

```text
h_F(v) = floor(chord(v,t) × min_cost_per_meter)
h_B(v) = floor(chord(s,v) × min_cost_per_meter)
```

これにより、三角関数を各展開ノードで繰り返さず、前向き・後ろ向きの整合下界を得る。

## Balanced feasible potential

```text
φ₂(v) = h_F(v) - h_B(v)
```

前向きreduced key:

```text
k_F(v) = 2g_F(v) + φ₂(v) - φ₂(s)
```

後ろ向きreduced key:

```text
k_B(v) = 2g_B(v) + φ₂(t) - φ₂(v)
```

元の各辺コストを`c`とすると、整合性により両方向のreduced edge costは非負となる。座標または安全なコスト/距離下限がない場合、potentialは0となり、厳密な双方向Dijkstra型探索へ退化する。

## Coupled lower bound and incumbent

各queueの有効最小keyを`min_F`、`min_B`とする。

```text
L₂ = min_F + min_B
```

前後両方向で同じノード`m`への経路が得られたとき、完全経路上界を更新する。

```text
U = min(U, g_F(m) + g_B(m))
U₂ = 2U + φ₂(t) - φ₂(s)
```

停止条件:

```text
L₂ >= U₂
```

返却JSONには`upperBound`、`lowerBound`、`optimalityGap`を証明用メトリクスとして記録する。厳密解では`optimalityGap = 0`となる。

## Optional incumbent-bound pruning experiment

上界`U`が得られた後、次を満たす前向き状態は、より短い経路へつながらない。

```text
g_F(v) + h_F(v) >= U
```

後ろ向きも同様である。

```text
g_B(v) + h_B(v) >= U
```

該当状態はqueueへの追加または展開を省略できる。ただし東京道路網ではこの条件がほぼ発火せず、追加のbound評価だけが残った。v0.7の既定`aegis`では無効化し、`aegis-prune`でのみ評価する。安全性の補題は実験変種のために維持する。

## Edge-work adaptive scheduling

v1はchunkサイズを展開頂点数で測っていた。v2は方向ごとの実コスト差を反映するため、確認辺数を予算単位とする。

```text
efficiency = coupled_lower_bound_gain
             / (relaxed_edges + 4×expanded_nodes + 2×positive_queue_growth)
```

方向別の効率は指数移動平均で平滑化する。

- 最初に前後を最低1回ずつ測定
- 効率差が10%を超えた場合は高効率側を優先
- 差が小さい場合は下界key、次ノード次数、queue長で決定
- 同一方向が6chunk続いたら反対側を再測定
- 辺予算はグラフ規模と効率比により256〜8192辺
- 上界発見後は最大予算を抑え、停止判定を細かく行う

スケジューラはどちらの有効最小keyを先に処理するかだけを決める。下界、上界、枝刈り条件を書き換えない。

## Ablation variants

研究用に次を実装する。

- `aegis`: edge-work scheduler、incumbent枝刈りなし
- `aegis-static`: 固定scheduler、incumbent枝刈りなし
- `aegis-prune`: edge-work scheduler + incumbent枝刈り
- `aegis-projection`: edge-work scheduler + 線形射影potential
- `aegis-no-prune`: 旧版互換alias

```bash
aegis benchmark --graph city.aegis --research
aegis benchmark --graph city.aegis --experimental
```

## Complexity

最悪時間計算量は`O((V+E) log V)`、グラフを除く探索作業領域は`O(V)`である。座標の単位ベクトルはグラフ読み込み時に`O(V)`で計算し、各クエリでは訪問ノードの弦距離下界だけをキャッシュする。

## Status

ACBS v2は検証可能な研究プロトタイプである。構成要素の多くは既存概念であり、組み合わせとedge-work schedulerの学術的新規性は未確定である。

## v0.5 implementation note: queue ownership

The ACBS mathematical algorithm is unchanged. The forward and backward priority queues are now owned by the pooled bidirectional workspace. A specialized binary heap mutates these slices directly, avoiding `container/heap` interface boxing and repeated backing-array growth. Queue state is reset to length zero when the workspace is released; capacity is deliberately retained for reuse. Returned paths are exact-sized and remain owned by the caller.

## v0.6 queue and graph representation

Each directional reduced key is monotone because every reduced edge cost is non-negative. v0.6 therefore stores `OPEN_F` and `OPEN_B` in independent monotone radix heaps. This changes queue complexity constants, not the expansion rule or termination proof.

The finalized graph uses CSR arrays:

```text
outOffsets, outEdges
inOffsets,  inEdges
```

Construction-time per-node adjacency slices are discarded after validation and deduplication. The graph format on disk remains compatible with previous `AEGIS12` files.

## Linear-projection potential ablation

`aegis-projection` uses a source-to-target unit direction `d` on the Earth chord embedding:

```text
p(v) = 2 R min_ratio <q(v), d>
```

For every edge `u→v`, Cauchy–Schwarz gives:

```text
|p(v)-p(u)| <= 2 min_ratio chord(u,v) <= 2 c(u,v)
```

After conservative scaling and integer truncation, both reduced edge directions remain non-negative. This is the same ACBS scheduler and termination rule with a cheaper, potentially weaker feasible potential. The default `aegis` continues to use the chord-difference potential.

## v0.7 ablation isolation

The production and static variants now share the same chord potential, radix queues, CSR graph, coupled termination test, and no incumbent-pruning pass. Therefore `aegis` versus `aegis-static` isolates only the direction scheduler. `aegis-prune` and `aegis-projection` are deliberately excluded from the default research set because they change separate mechanisms.

## Experimental late-upper-bound guard (v0.11)

`aegis-late-guard` is not part of the default ACBS algorithm. It is a bounded scheduler experiment for the single reproduced case where the adaptive scheduler found the first upper bound at the penultimate chunk and static scheduling was materially faster.

The guard is eligible only when all of the following hold:

- the graph metric is travel time;
- at least 48 chunks have completed;
- no finite upper bound has been discovered;
- both frontier-efficiency estimates have been sampled;
- at least half of completed chunks caused a direction switch;
- the two efficiency estimates differ by no more than 25%.

When eligible, the search uses the static lower-key scheduler and base edge budget for at most eight chunks. It does not change reduced costs, the incumbent, the coupled lower bound, or the stopping condition, so exactness is unchanged. The experiment remains separate from `aegis` until the release gates in `docs/RELEASE_PLAN.md` pass.


## Experimental connection guards (v0.11.1)

The v0.11.0 late guard did not materially improve the reproduced scheduler tail. v0.11.1 therefore evaluates three earlier or longer balanced-scheduling intervals. They modify scheduling only; feasible potentials, upper/lower bounds, termination, and exactness are unchanged. None is part of the default `aegis` algorithm until the Tokyo release gate selects it.

## Whole-suite trigger profiling (v0.11.2)

The v0.11.1 connection guards demonstrated that a condition broad enough to improve the reproduced Tokyo scheduler tail can activate on too many normal queries. v0.11.2 therefore adds diagnosis rather than another scheduler mutation.

`profile-trigger` runs the unchanged default ACBS once for every query referenced by a validation report. At configured chunk checkpoints it stores only deterministic scheduler state:

- cumulative and recent lower-bound gain per normalized work,
- direction-switch rate,
- forward/backward score, queue, priority, and directional-work imbalance,
- frontier growth,
- stale-pop rate,
- finite-meeting rate,
- queue population and cumulative work,
- whether a finite upper bound already exists.

Replay-confirmed `adaptive-scheduler-tail` cases are positive labels. The profiler enumerates interpretable one- and two-threshold rules, always requiring the upper bound to remain missing at the checkpoint. Rules are ranked by complete positive recall, total matches, false positives, and simplicity. The default eligibility cap is five matches across the whole validation suite.

This process does not alter search order, bounds, potentials, termination, or exactness. A selected rule is diagnostic evidence only and must be validated on independent seeds and cities before being implemented as a scheduler condition.

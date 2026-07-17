# Aegis Coupled-Bound Search v2

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

## Incumbent-bound pruning

上界`U`が得られた後、次を満たす前向き状態は、より短い経路へつながらない。

```text
g_F(v) + h_F(v) >= U
```

後ろ向きも同様である。

```text
g_B(v) + h_B(v) >= U
```

該当状態はqueueへの追加または展開を省略する。これは実行順序だけではなく探索量も減らすが、使用する値は証明可能な下界と実在する完全経路上界だけである。

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

- `aegis`: edge-work scheduler + incumbent pruning
- `aegis-static`: 下界key中心の固定scheduler + incumbent pruning
- `aegis-no-prune`: edge-work scheduler、上界枝刈りなし

```bash
aegis benchmark --graph city.aegis --research
```

## Complexity

最悪時間計算量は`O((V+E) log V)`、グラフを除く探索作業領域は`O(V)`である。座標の単位ベクトルはグラフ読み込み時に`O(V)`で計算し、各クエリでは訪問ノードの弦距離下界だけをキャッシュする。

## Status

ACBS v2は検証可能な研究プロトタイプである。構成要素の多くは既存概念であり、組み合わせとedge-work schedulerの学術的新規性は未確定である。

## v0.5 implementation note: queue ownership

The ACBS mathematical algorithm is unchanged. The forward and backward priority queues are now owned by the pooled bidirectional workspace. A specialized binary heap mutates these slices directly, avoiding `container/heap` interface boxing and repeated backing-array growth. Queue state is reset to length zero when the workspace is released; capacity is deliberately retained for reuse. Returned paths are exact-sized and remain owned by the caller.

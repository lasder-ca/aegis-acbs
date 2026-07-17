# Correctness sketch — ACBS v2

## Assumptions

1. グラフは有限。
2. 全辺コストは正の整数。
3. reverse adjacencyは元の有向辺を正確に反転する。
4. `min_cost_per_meter`は全地理的辺について`cost/大円距離`以下である。

## Lemma 1: chord heuristic is admissible and consistent

地球単位球面上の弦距離は大円距離以下である。また3次元ユークリッド距離なので三角不等式を満たす。

元の辺`u→v`について:

```text
chord(u,t) <= chord(u,v) + chord(v,t)
min_ratio × chord(u,v) <= cost(u,v)
```

したがって:

```text
h_F(u) <= cost(u,v) + h_F(v)
```

同様に`h_B`もreverse traversalに対して整合的である。浮動小数誤差に対して係数をわずかに縮小し、最後に切り捨てる。

## Lemma 2: reduced edge costs are non-negative

```text
φ₂(v) = h_F(v) - h_B(v)
```

整合性から:

```text
h_F(u) - h_F(v) <= c(u,v)
h_B(v) - h_B(u) <= c(u,v)
```

2式を加えると:

```text
φ₂(u) - φ₂(v) <= 2c(u,v)
```

よって前向きreduced cost:

```text
2c(u,v) + φ₂(v) - φ₂(u) >= 0
```

後ろ向きも対称に非負となる。

## Lemma 3: reduced complete-path costs differ by a constant

任意の`s→t`経路`P`について、前後reduced costの和はtelescopingにより:

```text
2 cost(P) + φ₂(t) - φ₂(s)
```

となる。定数項は全`s→t`経路で同じなので、元グラフとreduced graphの最短経路は一致する。

## Lemma 4: coupled lower bound is safe

各方向は非負reduced cost上のDijkstra探索であるため、queue最小keyはその方向の未確定部分経路に対する下界である。未証明の完全経路は、前向き未確定部分と後ろ向き未確定部分の組を含む。したがって:

```text
L₂ = min_F + min_B
```

は未証明完全経路のreduced cost下界となる。

## Lemma 5: incumbent pruning is safe

`U`を実際に発見済みの完全経路コストとする。前向き状態`v`について、任意の`v→t`経路コストは`h_F(v)`以上である。

```text
g_F(v) + h_F(v) >= U
```

なら、`v`を経由する経路は`U`未満にならない。すでに`U`の経路を保持しているため、その状態を展開しなくても最適解を失わない。後ろ向きも同様である。

## Theorem: ACBS v2 returns an optimal path

ACBSは実在する完全経路上界`U₂`を維持し、`L₂ >= U₂`の場合にのみ正常停止する。Lemma 4より、それ以降に`U₂`未満のreduced-cost経路は存在しない。Lemma 3よりreduced graphと元グラフの最短経路は一致する。Lemma 5の枝刈りも`U`未満になり得ない状態だけを除く。

したがって、到達可能な場合に返す経路は元グラフの厳密最短経路である。

edge-work schedulerは有効なqueue最小状態をどちらから先に処理するかだけを変えるため、正確性に影響しない。

## Machine checks

- 4頂点以下の全自己ループなし有向グラフ総当たり
- 250種類×40クエリのランダム有向時間道路グラフ
- ランダム距離道路グラフ
- chord距離が大円距離を超えないことを10,000組で確認
- 全返却辺の存在と経路コスト再計算
- Dijkstraとの到達可能性・距離一致
- reduced edge非負性
- `upperBound = lowerBound = distance`
- `optimalityGap = 0`
- 適応、固定scheduler、枝刈りなしの全variant
- 作業領域の連続再利用

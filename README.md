# Aegis ACBS v0.10.0-experimental

**Aegis Coupled-Bound Search（ACBS）は、前向き・後ろ向き探索の共通下界を結合し、探索中に方向別の辺予算を適応配分する、実験的な厳密最短経路アルゴリズムです。**

A*、双方向Dijkstra、Dijkstraのどれかを選ぶメタ選択器ではありません。`aegis`は1回の探索の中で2つのfrontierを同時に維持し、同じ停止証明を共有します。

> 研究上の新規性は未確定です。NBS、MM、DVCBS、BAE*、lower-bound propagationなどとの重複調査と第三者レビューが終わるまでは、`experimental`として扱います。

## 目的

- 重み付き・有向道路グラフ上で厳密な最短経路を返す
- 前向きと後ろ向きの探索量を、固定交互ではなく下界進行効率で配分する
- 適応処理が失敗しても、正確性へ影響させない
- Dijkstra、双方向Dijkstra、A*と同じ入力・前処理なしで比較する
- 実道路データ、再現可能なJSON、オフラインHTMLで検証する

## ACBSの核

1. 始点側と終点側に2つの優先キューを持つ。
2. 地理的な整合ヒューリスティックからbalanced potentialを作る。
3. 元の辺を非負のreduced-costへ変換し、両方向をDijkstraとして扱う。
4. `min_forward + min_backward`を安全な共通下界として維持する。
5. 完全経路の最良値を上界として維持する。
6. 下界の増加量を確認辺・展開頂点・queue成長で正規化し、次の辺予算chunkを配分する。
7. 共通下界が上界へ到達した時だけ停止し、gap 0の証明値を出力する。
8. 効果が未確認の`g+h >= U`枝刈りは`aegis-prune`実験だけで評価する。

適応制御は探索順序だけを変えます。既定`aegis`の停止条件には証明可能な共通下界と実在する上界だけを使います。

## 対応

- OSM XML
- DIMACS第9回Shortest Paths Challenge形式
- Aegisバイナリグラフ
- 距離・推定移動時間
- 車・自転車・徒歩
- 日本語・英語・簡体字中国語・韓国語・フランス語
- Linux・Windows・macOS

Wikipedia、Webクローリング、ハイパーリンク探索は含みません。

## WSL Ubuntuで開始

```bash
cd /mnt/c/Users/danda/Downloads/aegis-acbs-0.10.0-experimental
go test ./...
go build -o bin/aegis ./cmd/aegis

bin/aegis import-osm \
  --input benchdata/hatfield-uk.osm \
  --output artifacts/hatfield-distance.aegis \
  --profile car \
  --metric distance

bin/aegis benchmark \
  --graph artifacts/hatfield-distance.aegis \
  --queries 1000 \
  --repeats 9 \
  --order interleaved \
  --measure-memory \
  --suite mixed \
  --seed 1010 \
  --output artifacts/hatfield.json \
  --html artifacts/hatfield.html
```

## 比較アルゴリズム

- `dijkstra`: 正確性基準
- `bidijkstra`: 双方向Dijkstra
- `astar`: 地理ヒューリスティックA*
- `aegis`: 強いbalanced chord potentialと適応schedulerを使うACBS本体
- `aegis-static`: 適応schedulerなし。通常の研究比較で使うアブレーション
- `aegis-prune`: incumbent枝刈りを有効化する実験用変種
- `aegis-projection`: 平方根を避ける線形射影potentialの実験用変種
- `aegis-no-prune`: 旧コマンド互換用。現在は`aegis`と同じ探索
- `portfolio`: v12系の旧メタ選択器。比較用のみ
- `aegis-race`: A*と双方向Dijkstraを2コアで競争。比較条件が違うため標準ベンチから除外

## 研究アブレーション

```bash
bin/aegis benchmark \
  --graph artifacts/hatfield-distance.aegis \
  --queries 300 \
  --repeats 7 \
  --research \
  --output artifacts/research.json \
  --html artifacts/research.html
```

`--research`は通常の比較に`aegis-static`だけを追加し、適応schedulerの効果を単独で測ります。枝刈り・Projectionも含める場合は`--experimental`を使います。

```bash
bin/aegis benchmark --graph artifacts/hatfield-distance.aegis --experimental
```

## 複数seed・複数都市の再現試験

東京・横浜・大阪・名古屋について、距離・時間グラフを準備します。

```bash
scripts/prepare-japan-regional-benchmarks.sh
```

全グラフを複数seedで測定し、個別JSON/HTMLと集約JSON/CSV/HTMLを生成します。反復内では各方式の順序を決定論的にシャッフルして交互測定します。

```bash
scripts/benchmark-matrix.sh
```

準備から集約まで一括実行する場合:

```bash
scripts/run-japan-matrix.sh
```

既存レポートだけを再集約する場合:

```bash
bin/aegis aggregate   --input-dir artifacts/matrix   --output artifacts/matrix/benchmark-matrix.json   --csv artifacts/matrix/benchmark-matrix.csv   --html artifacts/matrix/benchmark-matrix.html
```

速度は次の3種類を区別します。

- Dijkstraと候補方式の中央値同士の比
- クエリ単位speedupの中央値
- クエリ単位speedupの幾何平均

また、`runtime vs fastest classical baseline`はDijkstra・双方向Dijkstra・A*の最速値との比なので1未満になり得ます。`classical oracle regret = max(1, runtime ratio)`は必ず1以上です。

## 大量tail検証

時間メトリックなどで、倍率だけでなく1ms以上の実損を伴うslowdownが再現するかを複数seedで検証します。完了済みseedは再利用されます。

```bash
AEGIS_QUERIES=1000 \
AEGIS_SEEDS="1010 20260717 424242 8675309 123456789 314159265 271828182 161803398 141421356 173205080" \
scripts/validate-tail.sh path/to/tokyo-time.aegis artifacts/tokyo-tail
```

集約だけを再実行する場合:

```bash
bin/aegis validate-regret \
  --input-dir artifacts/tokyo-tail \
  --min-queries 10000 \
  --max-meaningful-rate 0 \
  --output artifacts/tokyo-tail/regret-validation.json \
  --csv artifacts/tokyo-tail/regret-validation.csv \
  --html artifacts/tokyo-tail/regret-validation.html
```

0件だった場合も真の発生率が厳密に0と証明されるわけではありません。v0.9以降はWilson 95%区間と、ゼロ事象に対する片側95%上限を表示します。

## meaningful tailの隔離再実行

`validate-regret`で保持された問題クエリだけを繰り返し再測定します。通常のACBSは変更せず、固定scheduler版との比較とchunk単位の追跡を行います。

```bash
bin/aegis replay-regret \
  --graph path/to/tokyo-time.aegis \
  --validation artifacts/tokyo-tail/regret-validation.json \
  --input-root artifacts/tokyo-tail \
  --runs 31 \
  --warmup 5 \
  --output artifacts/tokyo-tail/regret-replay.json \
  --csv artifacts/tokyo-tail/regret-replay.csv \
  --html artifacts/tokyo-tail/regret-replay.html
```

分類は次の3種類です。

- `not-reproduced`: 隔離再測定ではmeaningful slowdownが消えた
- `adaptive-scheduler-tail`: 固定schedulerが適応schedulerを実時間で上回った
- `persistent-classical-tail`: 既存方式が速いがscheduler差だけでは説明できない

通常経路ではtraceを記録しません。`replay-regret`だけが、方向、辺budget、下界上昇、正規化work、frontierサイズ、効率score、上界発見chunkを保存します。

## 大規模な研究検証

4都市・距離/時間・10 seed・各1,000クエリの検証を実行します。

```bash
scripts/validate-research.sh
```

ACBS単体のプロセス最大RSSと割り当て量を測る場合:

```bash
scripts/memory-profile.sh path/to/graph.aegis
```

`--measure-memory`の割り当て測定は時間計測とは別の実行で行われるため、主要レイテンシを汚染しません。`peakRssBytes`はグラフ、Goランタイム、ベンチマーク基盤を含むプロセス全体の値です。

## 視覚レポート

HTMLレポートには以下を表示します。

- 平均・最良・最悪・p50 / p95 / p99レイテンシ
- 展開頂点・確認辺数
- queue push / pop / stale pop
- 実験用`aegis-prune`の展開時枝刈り・緩和時枝刈り
- ACBSの前向き・後ろ向き比率
- 距離クラス別の方向配分
- 最速の古典的ベースラインに対する実行時間比と、1以上に補正したclassical oracle regret
- 最初の上界発見位置
- 方向切替回数とchunk数
- 終了時下界・上界・optimality gap
- 上界枝刈り数、全接続検査、有限交差、上界更新数、potential評価数
- 方式別割り当てバイト/オブジェクト、プロセス最大RSS、Goヒープ
- 正確性比較表

## 検証

```bash
go test ./...
go vet ./...
go test -race ./...
```

テストには次を含みます。

- 4頂点以下の全有向グラフ総当たり
- ランダム道路グラフ
- 座標あり・なし
- 距離・時間メトリック
- 作業領域の連続再利用
- chord下界が大円距離を超えないこと
- reduced-costの非負性
- 前後両方向が実際に使われること
- Dijkstraとの距離一致と経路検証

## 文書

- [`docs/ALGORITHM.md`](docs/ALGORITHM.md): アルゴリズム仕様
- [`docs/CORRECTNESS.md`](docs/CORRECTNESS.md): 正確性の証明スケッチ
- [`docs/RELATED_WORK.md`](docs/RELATED_WORK.md): 既存研究との位置関係
- [`docs/NOVELTY_CHECKLIST.md`](docs/NOVELTY_CHECKLIST.md): 新規性を主張する前の必須確認
- [`docs/BENCHMARKING.md`](docs/BENCHMARKING.md): 測定方法
- [`docs/DATA.md`](docs/DATA.md): 道路データ

## ライセンス

コードはMITです。OpenStreetMap由来データには`© OpenStreetMap contributors`の表示が必要です。

## v0.6 search-core and graph-storage changes

v0.6 keeps the ACBS scheduler and exact coupled-bound termination rule. It changes the implementation and adds one potential ablation:

- monotone radix heaps for both bidirectional frontiers
- compact CSR forward/reverse adjacency storage
- lazy node-ID index construction
- 32-bit bidirectional parent and touched-node indexes
- `aegis-projection`, an exact linear-projection feasible-potential variant

The default `aegis` continues to use the stronger chord-difference potential. `aegis-projection` is not a portfolio selector; it is the same ACBS search with a cheaper feasible potential and must be measured separately on each domain.

Synthetic 180×180 road-grid measurements are implementation regression tests only. Tokyo and other OSM-derived city graphs remain the publication-quality evaluation.

## v0.7 production cleanup and stress validation

v0.7 keeps the same exact coupled-bound termination rule and the same chord potential. The default `aegis` no longer evaluates the inactive incumbent-pruning ablation. `aegis-static` now differs from production only in scheduler choice, making the scheduler ablation cleaner.

Experimental variants are opt-in:

```bash
aegis benchmark --graph city.aegis --experimental
```

Concurrent workspace and long-run behavior can be tested in one process:

```bash
aegis stress --graph city.aegis --queries 10000 --workers 8 --verify-every 100

# Separate ratio noise from material tail slowdowns
aegis diagnose --input benchmark.json --output regret.json --csv regret.csv --html regret.html
scripts/stress-matrix.sh city.aegis
scripts/soak.sh city.aegis
```

`stress` verifies sampled queries against Dijkstra while reporting throughput, p50/p95/p99/worst latency, peak RSS, Go heap, and error counts.

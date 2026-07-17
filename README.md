# Aegis ACBS v0.6.0-experimental

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
7. 上界発見後は`g+h >= U`の状態を安全に枝刈りする。
8. 共通下界が上界へ到達した時だけ停止し、gap 0の証明値を出力する。

適応制御は探索順序だけを変えます。停止条件と枝刈りには証明可能な値しか使いません。

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
cd /mnt/c/Users/danda/Downloads/aegis-acbs-0.6.0-experimental
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
- `aegis`: 強いbalanced chord potentialを使うACBS本体
- `aegis-projection`: 平方根を避ける線形射影potentialの研究用変種
- `aegis-static`: 適応schedulerなしの研究用アブレーション
- `aegis-no-prune`: 上界枝刈りなしの研究用アブレーション
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

`--research`は通常の比較に`aegis-static`、`aegis-no-prune`、`aegis-projection`を追加します。

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

また、`relative runtime`は1未満になり得ますが、`oracle regret = max(1, relative runtime)`は必ず1以上です。

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
- 展開時枝刈り・緩和時枝刈り
- ACBSの前向き・後ろ向き比率
- 距離クラス別の方向配分
- 最速ベースラインに対する相対実行時間と、1以上に補正したoracle regret
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

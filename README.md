# Aegis ACBS v0.2.0-experimental

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
cd /mnt/c/Users/danda/Downloads/aegis-acbs-0.2.0-experimental
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
  --suite mixed \
  --seed 1010 \
  --output artifacts/hatfield.json \
  --html artifacts/hatfield.html
```

## 比較アルゴリズム

- `dijkstra`: 正確性基準
- `bidijkstra`: 双方向Dijkstra
- `astar`: 地理ヒューリスティックA*
- `aegis`: ACBS v2本体
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

`--research`は通常の比較に`aegis-static`と`aegis-no-prune`を追加します。

## 視覚レポート

HTMLレポートには以下を表示します。

- p50 / p95レイテンシ
- 展開頂点・確認辺数
- ACBSの前向き・後ろ向き比率
- 距離クラス別の方向配分
- 最速既存方式に対するruntime regret
- 最初の上界発見位置
- 方向切替回数とchunk数
- 終了時下界・上界・optimality gap
- 上界枝刈り数とpotential評価数
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

# Aegis ACBS

[![CI](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml/badge.svg)](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Aegis Coupled-Bound Search（ACBS）**は、重み付き有向道路グラフ上で1対1の厳密最短経路を求める実装です。

始点側と終点側のfrontierを同時に進め、両方向で共有する下界と、発見済み経路の上界を維持します。探索中は、下界を効率よく押し上げている方向へ辺処理の予算を多く配分します。共有下界が最良の完全経路へ到達した時点で停止するため、schedulerは探索順序を変えても最短性の判定には影響しません。

> **研究段階:** 再現可能な研究プロトタイプとして公開しています。既存の双方向探索との関係は文書化していますが、学術的新規性や他の道路網への性能一般化は第三者検証前です。

English: [README.md](README.md)

## 特徴

- **厳密な経路探索:** 有限・非負重みの有向グラフで最短経路を返します。
- **証明用メトリクス:** 上界、停止時下界、optimality gapを出力します。
- **適応的な双方向探索:** 下界進行量を基に、前後frontierの辺処理量を調整します。
- **道路グラフ対応:** OSM XMLとDIMACSを取り込み、コンパクトなバイナリ形式へ変換します。
- **再現可能な評価:** JSON、CSV、単体で開けるHTMLレポートを生成します。
- **複数OS:** Linux、Windows、macOSでビルドとテストを行います。

## 現在の検証結果

2026年7月18日に、東京の時間重み付き道路グラフで10,000クエリを検証しました。グラフ規模は611,846ノード、1,235,323有向辺です。

| 検証項目 | 観測結果 |
|---|---:|
| Dijkstraと最短距離が一致 | **10,000 / 10,000** |
| 初回測定で検出したmeaningful slowdown | 11 / 10,000 |
| 隔離再測定でも再現 | 2 / 11 |
| 適応scheduler由来の再現tail | 1 / 10,000 |
| 既存方式が継続的に有利な再現tail | 1 / 10,000 |
| 事前定義ゲートを通過したguard候補 | **0 / 3** |
| 同一suite内で見つかった診断ルール | checkpoint 48、1件一致 |

この結果は、特定のグラフ、クエリ生成方法、実行環境に対する観測です。すべての道路網で高速であることを示すものではありません。生データ、判定基準、不採用実験は[東京検証の記録](docs/TOKYO_EVIDENCE.md)にまとめています。

## セットアップ

必要環境はGo 1.23以降です。

```bash
git clone https://github.com/lasder-ca/aegis-acbs.git
cd aegis-acbs

go test ./...
go build -o bin/aegis ./cmd/aegis
```

同梱のOSM fixtureを取り込みます。

```bash
bin/aegis import-osm \
  --input benchdata/hatfield-uk.osm \
  --output /tmp/hatfield-distance.aegis \
  --profile car \
  --metric distance
```

比較ベンチマークを実行します。

```bash
bin/aegis benchmark \
  --graph /tmp/hatfield-distance.aegis \
  --queries 1000 \
  --repeats 9 \
  --order interleaved \
  --measure-memory \
  --suite mixed \
  --seed 1010 \
  --output /tmp/hatfield.json \
  --html /tmp/hatfield.html
```

## 探索の構成

```text
始点  ->  前向きfrontier  ->  接続候補
                                  ^
終点  <-  後ろ向きfrontier  <-  接続候補

共有状態: 許容下界 L、発見済み完全経路の上界 U
停止条件: L >= U
```

適応schedulerが決めるのは、次の辺処理chunkをどちらのfrontierへ渡すかだけです。potential、上下界、最短性の停止条件は変更しません。

数式と状態遷移は[アルゴリズム](docs/ALGORITHM.md)、正確性の根拠は[Correctness](docs/CORRECTNESS.md)に分けています。

## 主なコマンド

| コマンド | 用途 |
|---|---|
| `import-osm` | OSM XMLをAegisグラフへ変換 |
| `import-dimacs` | DIMACS形式を取り込み |
| `route` | 1件の経路を計算 |
| `benchmark` | 複数方式を反復・交互測定 |
| `stress` | 並行実行とDijkstraによる標本検証 |
| `diagnose` | クエリ単位の性能tailを抽出 |
| `replay-regret` | 抽出したtailを隔離再測定 |
| `profile-trigger` | scheduler特徴量をcheckpointごとに記録 |
| `aggregate` | 複数seedの結果を集約 |

通常の比較対象はDijkstra、双方向Dijkstra、地理A*、固定scheduler版ACBS、適応scheduler版ACBSです。不採用になった実験変種は、結果の再現用としてのみ残しています。

## 研究結果の再現

```bash
# 複数seedでtailを検証
scripts/validate-tail.sh path/to/time-graph.aegis artifacts/tail

# 検出したtailを隔離再測定
bin/aegis replay-regret \
  --graph path/to/time-graph.aegis \
  --validation artifacts/tail/regret-validation.json \
  --input-root artifacts/tail \
  --runs 31 \
  --warmup 5 \
  --output artifacts/replay.json \
  --csv artifacts/replay.csv \
  --html artifacts/replay.html

# 全クエリのscheduler特徴量を記録
bin/aegis profile-trigger \
  --graph path/to/time-graph.aegis \
  --validation artifacts/tail/regret-validation.json \
  --replay artifacts/replay.json \
  --input-root artifacts/tail \
  --checkpoints 24,32,40,48 \
  --max-matches 5 \
  --output artifacts/trigger-profile.json \
  --csv artifacts/trigger-profile.csv \
  --html artifacts/trigger-profile.html
```

## ドキュメント

- [アルゴリズム](docs/ALGORITHM.md)
- [正確性](docs/CORRECTNESS.md)
- [ベンチマーク方法](docs/BENCHMARKING.md)
- [東京検証の記録](docs/TOKYO_EVIDENCE.md)
- [関連研究](docs/RELATED_WORK.md)
- [データ形式](docs/DATA.md)
- [セキュリティポリシー](SECURITY.md)
- [コントリビューション](CONTRIBUTING.md)

## 制約

- 性能はグラフ、重み、経路長、実行環境によって変わります。
- 公開済みの大規模検証は、現時点では東京の時間グラフが中心です。
- checkpoint 48のルールは同じsuite内で発見・評価しているため、診断結果としてのみ保持しています。
- contraction hierarchies、landmarksなどのグラフ固有前処理は使っていません。
- 学術的新規性と一般化性能には、独立したレビューと追試が必要です。

## リリース

`v0.1.0`が最初の公開版です。CHANGELOGにあるそれ以前の番号は、公開前の研究反復を示します。

## ライセンス

MIT Licenseです。詳細は[LICENSE](LICENSE)を参照してください。

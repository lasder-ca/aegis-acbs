# Aegis ACBS v0.2.0-experimental

ACBS v2は、v1の単一双方向探索を維持しながら、道路グラフで重かったpotential計算とchunk配分を作り直した研究版です。

## Main changes

- 緯度経度をグラフ読み込み時に3次元単位ベクトルへ前計算
- 大円距離以下で三角不等式を満たす弦距離potential
- 頂点数ではなく確認辺数を基準にしたadaptive edge budget
- 上界発見後の`g+h >= U`安全枝刈り
- `upperBound / lowerBound / optimalityGap`証明値
- `aegis-static`と`aegis-no-prune`アブレーション
- `benchmark --research`
- 日英中韓仏の視覚レポートに枝刈りKPIを追加

## Validation

- 4頂点以下の全有向グラフ総当たり
- 250種類×40クエリのランダム有向時間道路グラフ
- chord vs great-circle 10,000組
- Dijkstra距離・到達可能性・経路連続性一致
- gap 0証明値
- `go test`, `go vet`, race detector

## Builds

- Linux amd64
- Linux arm64
- Windows amd64
- macOS amd64

## Research status

研究上の新規性は未確定です。MM、NBS、DVCBS、BAE*、2025年のtight termination methodとの比較が完了するまで`experimental`を維持します。

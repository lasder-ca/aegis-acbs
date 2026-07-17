# Related work and novelty boundary

## Directly related families

- bidirectional Dijkstra / bidirectional A*
- MM and MMe: meet-in-the-middle priority and termination
- NBS: must-expand pairs and near-optimal expansion guarantees
- DVCBS and lower-bound propagation
- BAE* and individual heuristic-error bounds
- tighter exact bidirectional termination conditions

## Components ACBS does not claim as new

- two-frontier exact search
- admissible and consistent heuristics
- feasible/balanced potentials and reduced costs
- incumbent upper bounds
- lower-bound termination
- A*-style incumbent pruning
- workspace reuse
- OSM or DIMACS road-graph processing

## Provisional ACBS contribution

ACBSの独自候補は、次を1つの道路向け厳密探索として統合した点にある。

1. balanced chord potentialによる安価な双方向reduced-cost探索
2. 共通下界の上昇量を、展開頂点数ではなく実際の確認辺・queue成長で割るオンライン効率
3. その効率に応じた辺予算chunkの動的配分
4. 同一の上界を用いた両方向の安全な枝刈り
5. schedulerと枝刈りを個別に外せるアブレーション

この組み合わせが既存研究と実質的に異なるかは未確認である。特に2025年のtight termination、NBS/DVCBSのpair lower bound、BAE*のindividual boundsとの詳細比較が必要である。

## Required comparisons before publication

1. A*
2. bidirectional Dijkstra
3. MM / MMe
4. NBS
5. DVCBS
6. BAE*
7. 2025 tight-termination method
8. `aegis-static`
9. `aegis-no-prune`
10. ACBS full

報告項目:

- p50 / p95 / p99実行時間
- 展開頂点、確認辺、queue push
- 最大常駐メモリ
- 最初の上界発見時点
- 上界更新回数
- 枝刈り状態数
- potential評価数
- 方向切替とchunk数
- 終了下界、上界、optimality gap
- 距離クラス別結果

## Publication rule

第三者レビューと既存実装比較が終わるまで、「世界初」「最先端」「既存方式を上回る」とは表現しない。

## Primary references

- Holte et al., “MM: A bidirectional search algorithm that is guaranteed to meet in the middle,” *Artificial Intelligence* 252, 2017. DOI: `10.1016/j.artint.2017.05.004`.
- Chen et al., “Front-to-End Bidirectional Heuristic Search with Near-Optimal Node Expansions,” 2017. arXiv: `1703.03868`.
- Shperberg et al., “Bidirectional Heuristic Search: Expanding Nodes by a Lower Bound,” IJCAI 2020. DOI: `10.24963/ijcai.2020/664`.
- Alcázar, Riddle, Barley, “A Unifying View on Individual Bounds and Heuristic Inaccuracies in Bidirectional Search,” AAAI 2020. DOI: `10.1609/aaai.v34i03.5611`.
- Wang et al., “Bidirectional Search while Ensuring Meet-In-The-Middle via Effective and Efficient-to-Compute Termination Conditions,” IJCAI 2025. DOI: `10.24963/ijcai.2025/999`.

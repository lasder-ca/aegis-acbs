# Novelty checklist

ACBSを論文・技術記事で新規アルゴリズムとして主張する前に、以下をすべて完了する。

- [ ] MM / MMeのpriorityと停止条件を数式単位で比較
- [ ] NBSのmust-expand pair lower boundと比較
- [ ] DVCBSのpair選択とlb-propagationを比較
- [ ] BAE*のindividual boundsと比較
- [ ] 2025 tight termination methodを実装または公式実装で比較
- [ ] edge-work schedulerと同等の既存方式を文献検索
- [ ] full / static / no-pruneのアブレーションを5都市で実施
- [ ] 距離・時間メトリックを分離
- [ ] 同一OS・CPU・Go版・固定GOMAXPROCSで測定
- [ ] bootstrap confidence intervalを算出
- [ ] 全生データ、seed、地図SHA-256を公開
- [ ] 第三者が再現

未完了項目がある間は`experimental`を外さない。

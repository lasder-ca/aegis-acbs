package i18n

import "strings"

type Language string

const (
	JA Language = "ja"
	EN Language = "en"
	ZH Language = "zh-CN"
	KO Language = "ko"
	FR Language = "fr"
)

var supported = []Language{JA, EN, ZH, KO, FR}

var messages = map[Language]map[string]string{
	JA: {
		"app.name": "Aegis ACBS", "app.tagline": "前後の下界を結合する実験的な厳密最短経路アルゴリズム",
		"graph": "グラフ", "nodes": "頂点", "edges": "辺", "profile": "移動手段", "metric": "重み",
		"source": "出発地点", "target": "到着地点", "algorithm": "アルゴリズム", "route": "経路を検索",
		"benchmark": "ベンチマーク", "queries": "クエリ数", "run": "実行", "result": "結果", "language": "言語",
		"distance": "距離", "time": "時間", "reachable": "到達可能", "expanded": "展開頂点", "relaxed": "確認した辺",
		"median": "中央値", "p95": "95パーセンタイル", "correct": "正確性", "loading": "処理中…",
		"error": "エラー", "source_help": "ノードIDまたは緯度,経度", "target_help": "ノードIDまたは緯度,経度",
		"no_graph": "グラフが読み込まれていません", "attribution": "地図データ © OpenStreetMap contributors",
		"dashboard": "可視化ダッシュボード", "routeSearch": "経路探索", "benchmarkLab": "ベンチマーク分析", "repeats": "反復回数",
		"latency": "レイテンシ", "searchWork": "探索量", "selection": "前後探索比率", "regret": "最速基準比", "oracleHit": "基準比較",
		"speedup": "対Dijkstra速度", "allCorrect": "全件一致", "decision": "Aegisの判断", "reason": "選択理由",
		"heuristicStrength": "ヒューリスティック強度", "distanceRatio": "地図に対する距離比", "predictedWork": "予測作業量", "visualSummary": "視覚サマリー", "forward": "前向き", "backward": "後ろ向き", "switches": "方向切替", "bound": "終了下界/解", "reason.small_graph": "小規模グラフではDijkstraの固定費が最小", "reason.local_query": "近距離クエリでは単方向探索が有利", "reason.strong_geographic_heuristic": "地理ヒューリスティックが十分に強い", "reason.coordinates_unavailable": "座標ヒューリスティックを利用できない", "reason.weak_geographic_heuristic": "地理ヒューリスティックが弱い", "reason.balanced_frontiers": "双方向frontierの予測作業量が最小", "reason.lowest_predicted_work": "予測作業量が最小",
	},
	EN: {
		"app.name": "Aegis ACBS", "app.tagline": "Experimental exact shortest-path search with coupled forward/backward bounds",
		"graph": "Graph", "nodes": "Nodes", "edges": "Edges", "profile": "Profile", "metric": "Metric",
		"source": "Source", "target": "Target", "algorithm": "Algorithm", "route": "Find route",
		"benchmark": "Benchmark", "queries": "Queries", "run": "Run", "result": "Result", "language": "Language",
		"distance": "Distance", "time": "Time", "reachable": "Reachable", "expanded": "Expanded nodes", "relaxed": "Relaxed edges",
		"median": "Median", "p95": "95th percentile", "correct": "Correctness", "loading": "Working…",
		"error": "Error", "source_help": "Node ID or latitude,longitude", "target_help": "Node ID or latitude,longitude",
		"no_graph": "No graph loaded", "attribution": "Map data © OpenStreetMap contributors",
		"dashboard": "Visual dashboard", "routeSearch": "Route search", "benchmarkLab": "Benchmark analysis", "repeats": "Repeats",
		"latency": "Latency", "searchWork": "Search work", "selection": "Direction balance", "regret": "Relative runtime", "oracleHit": "Baseline comparison",
		"speedup": "Speed vs Dijkstra", "allCorrect": "All matched", "decision": "Aegis decision", "reason": "Selection reason",
		"heuristicStrength": "Heuristic strength", "distanceRatio": "Map distance ratio", "predictedWork": "Predicted work", "visualSummary": "Visual summary", "forward": "Forward", "backward": "Backward", "switches": "Direction switches", "bound": "Termination bound / path", "reason.small_graph": "Dijkstra has the lowest setup cost on this small graph", "reason.local_query": "One-direction search is cheaper for this local query", "reason.strong_geographic_heuristic": "The geographic heuristic is strong enough", "reason.coordinates_unavailable": "No usable coordinate heuristic is available", "reason.weak_geographic_heuristic": "The geographic heuristic is too weak", "reason.balanced_frontiers": "Bidirectional frontiers have the lowest predicted work", "reason.lowest_predicted_work": "Lowest predicted work",
	},
	ZH: {
		"app.name": "Aegis ACBS", "app.tagline": "结合前后向下界的实验性精确最短路径算法",
		"graph": "图", "nodes": "节点", "edges": "边", "profile": "出行方式", "metric": "权重",
		"source": "起点", "target": "终点", "algorithm": "算法", "route": "查找路线",
		"benchmark": "基准测试", "queries": "查询数", "run": "运行", "result": "结果", "language": "语言",
		"distance": "距离", "time": "时间", "reachable": "可达", "expanded": "扩展节点", "relaxed": "检查边数",
		"median": "中位数", "p95": "第95百分位", "correct": "正确性", "loading": "处理中…",
		"error": "错误", "source_help": "节点ID或纬度,经度", "target_help": "节点ID或纬度,经度",
		"no_graph": "未加载图", "attribution": "地图数据 © OpenStreetMap contributors",
		"dashboard": "可视化仪表板", "routeSearch": "路线搜索", "benchmarkLab": "基准分析", "repeats": "重复次数",
		"latency": "延迟", "searchWork": "搜索工作量", "selection": "前后向比例", "regret": "相对运行时间", "oracleHit": "基线比较",
		"speedup": "相对Dijkstra速度", "allCorrect": "全部一致", "decision": "Aegis决策", "reason": "选择原因",
		"heuristicStrength": "启发式强度", "distanceRatio": "地图距离比", "predictedWork": "预测工作量", "visualSummary": "可视化摘要", "forward": "前向", "backward": "后向", "switches": "方向切换", "bound": "终止下界/路径", "reason.small_graph": "小型图中Dijkstra的启动成本最低", "reason.local_query": "短距离查询使用单向搜索更合适", "reason.strong_geographic_heuristic": "地理启发式足够强", "reason.coordinates_unavailable": "无法使用坐标启发式", "reason.weak_geographic_heuristic": "地理启发式过弱", "reason.balanced_frontiers": "双向前沿的预测工作量最低", "reason.lowest_predicted_work": "预测工作量最低",
	},
	KO: {
		"app.name": "Aegis ACBS", "app.tagline": "정방향·역방향 하한을 결합하는 실험적 정확 최단 경로 알고리즘",
		"graph": "그래프", "nodes": "노드", "edges": "간선", "profile": "이동 수단", "metric": "가중치",
		"source": "출발지", "target": "도착지", "algorithm": "알고리즘", "route": "경로 찾기",
		"benchmark": "벤치마크", "queries": "쿼리 수", "run": "실행", "result": "결과", "language": "언어",
		"distance": "거리", "time": "시간", "reachable": "도달 가능", "expanded": "확장 노드", "relaxed": "검사한 간선",
		"median": "중앙값", "p95": "95백분위수", "correct": "정확성", "loading": "처리 중…",
		"error": "오류", "source_help": "노드 ID 또는 위도,경도", "target_help": "노드 ID 또는 위도,경도",
		"no_graph": "그래프가 로드되지 않았습니다", "attribution": "지도 데이터 © OpenStreetMap contributors",
		"dashboard": "시각화 대시보드", "routeSearch": "경로 탐색", "benchmarkLab": "벤치마크 분석", "repeats": "반복 횟수",
		"latency": "지연 시간", "searchWork": "탐색 작업량", "selection": "양방향 비율", "regret": "상대 실행 시간", "oracleHit": "기준 비교",
		"speedup": "Dijkstra 대비 속도", "allCorrect": "전부 일치", "decision": "Aegis 판단", "reason": "선택 이유",
		"heuristicStrength": "휴리스틱 강도", "distanceRatio": "지도 거리 비율", "predictedWork": "예측 작업량", "visualSummary": "시각 요약", "forward": "정방향", "backward": "역방향", "switches": "방향 전환", "bound": "종료 하한/경로", "reason.small_graph": "작은 그래프에서는 Dijkstra의 초기 비용이 가장 낮음", "reason.local_query": "근거리 쿼리에는 단방향 탐색이 유리함", "reason.strong_geographic_heuristic": "지리 휴리스틱이 충분히 강함", "reason.coordinates_unavailable": "좌표 휴리스틱을 사용할 수 없음", "reason.weak_geographic_heuristic": "지리 휴리스틱이 너무 약함", "reason.balanced_frontiers": "양방향 frontier의 예측 작업량이 가장 낮음", "reason.lowest_predicted_work": "예측 작업량이 가장 낮음",
	},
	FR: {
		"app.name": "Aegis ACBS", "app.tagline": "Algorithme expérimental exact avec bornes avant/arrière couplées",
		"graph": "Graphe", "nodes": "Sommets", "edges": "Arêtes", "profile": "Mode", "metric": "Métrique",
		"source": "Départ", "target": "Arrivée", "algorithm": "Algorithme", "route": "Calculer l’itinéraire",
		"benchmark": "Benchmark", "queries": "Requêtes", "run": "Exécuter", "result": "Résultat", "language": "Langue",
		"distance": "Distance", "time": "Temps", "reachable": "Accessible", "expanded": "Sommets explorés", "relaxed": "Arêtes examinées",
		"median": "Médiane", "p95": "95e percentile", "correct": "Exactitude", "loading": "Traitement…",
		"error": "Erreur", "source_help": "ID de sommet ou latitude,longitude", "target_help": "ID de sommet ou latitude,longitude",
		"no_graph": "Aucun graphe chargé", "attribution": "Données cartographiques © contributeurs OpenStreetMap",
		"dashboard": "Tableau de bord visuel", "routeSearch": "Recherche d’itinéraire", "benchmarkLab": "Analyse du benchmark", "repeats": "Répétitions",
		"latency": "Latence", "searchWork": "Travail de recherche", "selection": "Équilibre des directions", "regret": "Temps relatif", "oracleHit": "Comparaison de référence",
		"speedup": "Vitesse vs Dijkstra", "allCorrect": "Tous identiques", "decision": "Décision Aegis", "reason": "Raison du choix",
		"heuristicStrength": "Force heuristique", "distanceRatio": "Ratio de distance", "predictedWork": "Travail prédit", "visualSummary": "Résumé visuel", "forward": "Avant", "backward": "Arrière", "switches": "Changements de direction", "bound": "Borne finale / chemin", "reason.small_graph": "Dijkstra a le coût initial le plus faible sur ce petit graphe", "reason.local_query": "La recherche unidirectionnelle est moins coûteuse pour ce trajet local", "reason.strong_geographic_heuristic": "L’heuristique géographique est suffisamment forte", "reason.coordinates_unavailable": "Aucune heuristique de coordonnées exploitable", "reason.weak_geographic_heuristic": "L’heuristique géographique est trop faible", "reason.balanced_frontiers": "Les deux frontières ont le travail prédit le plus faible", "reason.lowest_predicted_work": "Travail prédit minimal",
	},
}

func Normalize(s string) Language {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "ja"):
		return JA
	case strings.HasPrefix(s, "zh"):
		return ZH
	case strings.HasPrefix(s, "ko"):
		return KO
	case strings.HasPrefix(s, "fr"):
		return FR
	default:
		return EN
	}
}

func T(lang Language, key string) string {
	if m, ok := messages[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := messages[EN][key]; ok {
		return v
	}
	return key
}

func Catalog(lang Language) map[string]string {
	out := map[string]string{}
	for k, v := range messages[EN] {
		out[k] = v
	}
	if m, ok := messages[lang]; ok {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func Supported() []Language { return append([]Language(nil), supported...) }

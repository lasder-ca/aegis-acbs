# Road data inputs

## OSM XML

Aegis streams nodes and ways from OSM XML. It filters routable `highway=*` ways according to the selected car, bike or walk profile. It handles:

- consecutive way-node edges;
- `oneway=yes`, `oneway=-1`, and roundabouts;
- access restrictions relevant to the profile;
- `maxspeed` in km/h and mph;
- conservative default speeds by highway type;
- distance or estimated travel-time costs.

OSM relations and turn restrictions are not implemented in the current importer. Therefore, the benchmark measures graph shortest paths rather than a production navigation engine with turn penalties.

## OSM PBF

Use `scripts/import-pbf.sh`. `osmium-tool` performs PBF decoding and emits OSM XML for the dependency-free Aegis importer.

## DIMACS

The importer accepts the Ninth DIMACS Shortest Paths Challenge `.gr` arc format and optional `.co` coordinates. DIMACS edge weights are preserved unchanged.

## Included fixture

The Hatfield fixture is a real OSM-derived extract for deterministic tests. It is intentionally small and must not be used as the only performance dataset.

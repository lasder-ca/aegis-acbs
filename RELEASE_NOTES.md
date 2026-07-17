# Aegis ACBS v0.6.0-experimental

v0.6.0 continues the exact ACBS research line. It does not reintroduce the old algorithm selector. The default `aegis` remains one coupled bidirectional search with the same adaptive edge-work scheduler and coupled lower-bound termination rule.

## Main changes

- Replace both bidirectional binary heaps with monotone radix heaps over `uint64` reduced keys.
- Compact forward and reverse road adjacency into CSR storage after import/load.
- Drop the always-resident node-ID hash index; build it lazily only for ID lookups.
- Store bidirectional parents and touched-node indexes in 32 bits.
- Keep `aegis` on the stronger balanced chord-difference potential.
- Add `aegis-projection`, the same exact ACBS search using a cheaper 3D linear-projection feasible potential.
- Add tests for radix ordering, non-negative reduced costs under both potentials, exhaustive small graphs, random directed time graphs, path validity, and zero optimality gaps.

## Local implementation results

On the generated 180×180 road-grid regression fixture:

- the isolated release comparison measured v0.5 default ACBS at 8.687 ms/query and v0.6 default chord ACBS at 6.356 ms/query, a 26.8% reduction;
- a separate five-run diagnostic measured the v0.6 projection variant at a median of approximately 5.11 ms/query;
- steady-state allocation remains one exact-sized returned path.

On a 250,000-node / 998,000-edge generated graph after one warmed query:

- v0.5 retained Go allocation: approximately 82.04 MiB.
- v0.6 retained Go allocation: approximately 59.37 MiB.
- retained heap objects fell from roughly 500,000 to fewer than 700.

These generated fixtures are implementation diagnostics, not evidence of superiority on real road networks. Tokyo distance/time results must be measured independently.

## Potential interpretation

`aegis-projection` can be faster per evaluated state because it avoids two square roots. Its lower bound can be weaker than the default chord-difference potential and may expand more states. It is included as an ablation, not selected automatically.

## Research status

Academic novelty remains unconfirmed. ACBS must still be compared with MM/MMe, NBS, DVCBS, BAE*, and the IJCAI-25 MEET termination method before publication claims are made.

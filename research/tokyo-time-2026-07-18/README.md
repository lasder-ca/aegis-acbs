# Raw Tokyo evidence directory

Run `scripts/import-tokyo-evidence.sh ~/aegis-benchmark/tokyo-v12.1.0` before public release.

The importer copies the user-run validation, replay, rejected-guard benchmark, release-gate, and trigger-profile JSON/CSV/HTML files into this directory and creates `MANIFEST.sha256`.

`observed-summary.json` is already tracked as a console-result transcription. Raw files remain authoritative.

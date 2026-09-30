# Cross-language contracts

Plain data files that several modules test against, so a change to one side
fails a test on the others instead of drifting silently.

| File | Meaning | Tested by |
|---|---|---|
| `telemetry-metrics.json` | The 20 telemetry metric names stored by the ingestor and referenced by the compliance rules | `elevator-simulator/tests/test_contract.py`, `telemetry-ingestor/internal/ingest/contract_test.go`, `compliance-engine/internal/rules/registry_test.go` |
| `simulator-snapshot.json` | One `GET /elevators` item exactly as the simulator emits it | `elevator-simulator/tests/test_contract.py` (same keys as `to_dict()`), `telemetry-ingestor/internal/ingest/contract_test.go` (decodes and maps to the 20 metrics, all `good`) |

Changing a metric name means changing these files and every side in the same
commit. See `docs/adr/0001-rule-catalogue-lives-in-compliance-engine.md`.

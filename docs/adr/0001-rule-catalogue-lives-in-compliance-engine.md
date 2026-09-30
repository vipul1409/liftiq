# ADR 0001 — The rule catalogue lives only in compliance-engine

**Status:** Accepted, 2026-09-30

## Context

Each rule's subsystem grouping, metric name and threshold semantics used to be
restated in several modules: the web and mobile checklists and the
report-generator each had a hand-copied rule ID → subsystem map; metric names
were listed in the simulator, the ingestor model, the ingestor mapper and a
compliance test. A new rule (ASME-021) or a renamed metric drifted silently, and a
renamed simulator key reached the rules as a fabricated `0` reading.

## Decision

- `compliance-engine/internal/rules/registry.go` is the only rule catalogue. Each
  rule states `Subsystem`, `Metric`, `Comparison` (`at_most` / `below` / `equals`)
  and `Threshold` as data; pass/fail follows from those, with no per-rule code.
- Every other module learns rule facts from the `Result` fields the engine returns
  (`subsystem`, `comparison`, `threshold`, `unit`). Clients and the report group by
  `subsystem` in result order; a result without one goes under "Other", and a
  rule is never dropped.
- Metric names cross the Python/Go seam through the plain files in `contracts/`,
  which the simulator, ingestor and compliance tests all read.

## Consequences

- Adding or regrouping a rule is a change to `registry.go` only. No client or report change is needed.
- Do not add an ID → subsystem map, a threshold, or a metric list to any client,
  the report-generator or a test. Read it from `Result`, or from `contracts/`.
- Renaming a metric fails the contract tests until the simulator, ingestor mapper,
  rules and `contracts/telemetry-metrics.json` are changed together.

# LiftIQ domain glossary

**Inspection** — one technician's certification of one elevator unit at a point in time: the telemetry rule results, the technician's Overrides, photo evidence and signature. Ends in a signed Report.

**Rule result** (telemetry status) — the compliance engine's deterministic pass / fail / unknown for one ASME A17.1 rule, computed from the latest telemetry. It is evidence and is never edited; an Override sits beside it.

**Rule catalogue** — the 20 ASME A17.1 rules as data, held only in `compliance-engine/internal/rules/registry.go`. For each rule it records the metric, subsystem, comparison and threshold. See `docs/adr/0001`.

**Subsystem** — the display group a rule belongs to, e.g. "Door Operator". It comes from the Rule catalogue on every rule result. Results with no subsystem group under "Other".

**Comparison** — how a metric value is checked against a threshold: `at_most` (≤), `below` (<, service due at N), or `equals` (boolean safety fields).

**Metric quality** — the state of a stored telemetry reading: `good` means a real reading, and `missing` means the simulator payload lacked the field. Only `good` readings are evaluated, so a latest reading that isn't good makes its rule unknown.

**Override** — the technician's manual call on a rule, `pass` or `fail` only. It can replace any telemetry status, including unknown. Clients send Overrides as a map of rule ID → status, separately from the rule results.

**Effective status** — the status the Report certifies for a rule: the Override if there is one, otherwise the telemetry status.

**Inspection outcome** — the effective status of every rule plus the summary derived from them (overall is fail if any rule fails, unknown if any is unknown, otherwise pass). The report-generator's `report.Resolve` is authoritative. Clients compute the same thing only for on-screen display.

**Report** — the signed PDF, the compliance record. Every overridden row shows both the technician's call and the telemetry status, so the audit trail keeps both.

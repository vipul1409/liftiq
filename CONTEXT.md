# LiftIQ domain glossary

**Inspection** — one technician's certification of one elevator unit at a point in time: the telemetry rule results, the technician's Overrides, photo evidence and signature. Ends in a signed Report.

**Rule result** (telemetry status) — the compliance engine's deterministic pass / fail / unknown for one ASME A17.1 rule, computed from the latest telemetry. It is evidence and is never edited; an Override sits beside it.

**Override** — the technician's manual call on a rule, `pass` or `fail` only. It can replace any telemetry status, including unknown. Clients send Overrides as a map of rule ID → status, separately from the rule results.

**Effective status** — the status the Report certifies for a rule: the Override if there is one, otherwise the telemetry status.

**Inspection outcome** — the effective status of every rule plus the summary derived from them (overall is fail if any rule fails, unknown if any is unknown, otherwise pass). The report-generator's `report.Resolve` is authoritative. Clients compute the same thing only for on-screen display.

**Report** — the signed PDF, the compliance record. Every overridden row shows both the technician's call and the telemetry status, so the audit trail keeps both.

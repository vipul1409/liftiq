# LiftIQ

LiftIQ turns live telemetry from vertical-transport equipment into ASME A17.1 compliance evidence that a technician reviews, corrects and signs on site.

## Language

**Unit**:
One piece of vertical-transport equipment LiftIQ monitors, identified by its unit tag (for example ELV-003).
_Avoid_: elevator, lift, device (when meaning the general concept)

### Rules and telemetry

**Rule catalogue**:
The fixed set of ASME A17.1 rules LiftIQ checks. Each rule names one Metric, a Subsystem, a Comparison and a threshold.
_Avoid_: rule registry, rule list

**Subsystem**:
The part of a Unit a rule concerns, such as Motor, Door Operator or Brake System.
_Avoid_: category, section, group

**Metric**:
One named, measurable quantity reported by a Unit, such as door close force or brake response time.
_Avoid_: data point, field, sensor

**Comparison**:
How a Metric value must relate to a rule's threshold to pass: at most the threshold, below it (service falls due at the threshold), or equal to it (safety circuits).
_Avoid_: operator, condition

**Metric quality**:
Whether a stored reading is a real measurement (good) or a placeholder for a value the Unit did not report (missing). Only good readings count as evidence.
_Avoid_: stale (staleness is about age, not quality — see Stale window)

**Stale window**:
How old the latest good reading of a Metric may be before its rule's result becomes unknown.

**Compliance snapshot**:
The Rule results for one Unit as of one moment, as LiftIQ evaluates them from telemetry. It is what a technician reviews; it is not a Report.
_Avoid_: compliance report, compliance check

### Inspection

**Inspection**:
One technician's certification of one Unit, covering the latest Compliance snapshot at signing, the technician's Overrides, photo evidence and signature. An Inspection ends in a Report.

**Rule result**:
The deterministic pass, fail or unknown that LiftIQ derives for one rule from the latest good telemetry. It is evidence and is never edited.
_Avoid_: telemetry status (when meaning the whole result), check

**Override**:
The technician's own pass or fail call on a rule, made against a specific Rule result and recorded beside it rather than replacing it. If that Rule result's status later changes to something other than the Override, the technician must confirm the Override again before it counts, and the Inspection cannot be signed until they do. It can be made for any rule, including one whose result is unknown.
_Avoid_: manual status, correction

**Effective status**:
The status an Inspection certifies for a rule: the Override if there is one, otherwise the Rule result.

**Inspection outcome**:
The Effective status of every rule, plus the overall verdict: fail if any rule fails, unknown if any is unknown, otherwise pass.
_Avoid_: summary (when meaning the verdict)

**Report**:
The signed record of an Inspection, and the only thing LiftIQ calls a report. For every overridden rule it shows both the technician's call and the Rule result.
_Avoid_: PDF, certificate

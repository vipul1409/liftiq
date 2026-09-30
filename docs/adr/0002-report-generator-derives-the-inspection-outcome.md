# The report-generator derives the Inspection outcome; clients never supply it

**Status:** accepted

The signed Report is the compliance record, so the Effective status of each rule and the overall verdict are computed in one deterministic, tested place: the report-generator. Clients send the untouched Rule results plus the technician's Overrides, and never a summary. When clients built the payload themselves, mobile silently dropped Overrides and the Report certified telemetry status the technician had rejected. Clients still compute the same outcome for on-screen display only, so it is duplicated for now; a shared client module would remove the duplication, but it must never become the source the Report trusts.

## Considered Options

- **Clients build the full payload, with a shared TypeScript module deriving the outcome.** Rejected: the audit record would depend on two JavaScript apps staying correct, and the server would still have to trust whatever it was sent.
- **The server validates a client-sent summary and rejects mismatches.** Rejected as unnecessary: once the server derives the summary itself, a client-sent one adds only a way to disagree.

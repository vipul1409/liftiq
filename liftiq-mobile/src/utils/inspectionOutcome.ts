import type { ComplianceSummary, RuleResult } from '../types/compliance';
import type { ReportOverride, ReportPhoto, ReportRequest } from '../types/report';

/** rule_id → technician override. Plain object so it survives navigation params. */
export type Overrides = Record<string, ReportOverride>;

/**
 * Returns results with each overridden rule's status replaced by the
 * technician's call. For on-screen display only — the report-generator
 * re-derives this authoritatively from the raw results + overrides.
 */
export function applyOverrides(results: RuleResult[], overrides: Overrides): RuleResult[] {
  return results.map((r) => {
    const override = overrides[r.rule_id];
    return override ? { ...r, status: override } : r;
  });
}

/**
 * Tallies statuses. Overall is fail if any rule failed, unknown if any is
 * unknown (and none failed), pass only when all pass — matching the Go engine.
 */
export function summarise(results: RuleResult[]): ComplianceSummary {
  let pass = 0;
  let fail = 0;
  let unknown = 0;
  for (const r of results) {
    if (r.status === 'pass') pass++;
    else if (r.status === 'fail') fail++;
    else unknown++;
  }
  const overall = fail > 0 ? 'fail' : unknown > 0 ? 'unknown' : 'pass';
  return { pass, fail, unknown, overall };
}

interface ReportInput {
  unitTag: string;
  inspectedAt: string;
  technician: string;
  /** Telemetry results as returned by the compliance engine — not overridden. */
  results: RuleResult[];
  overrides: Overrides;
  photos: ReportPhoto[];
  signatureDataURI: string;
}

/** Builds the POST /reports body: raw telemetry results plus overrides, no summary. */
export function buildReportRequest(input: ReportInput): ReportRequest {
  return {
    unit_tag: input.unitTag,
    inspected_at: input.inspectedAt,
    technician: input.technician,
    results: input.results.map((r) => ({
      rule_id: r.rule_id,
      description: r.description,
      asme_ref: r.asme_ref,
      metric: r.metric,
      value: r.value ?? 0,
      threshold: r.threshold,
      unit: r.unit,
      status: r.status,
      message: r.message,
    })),
    overrides: { ...input.overrides },
    photos: input.photos,
    signature_data_uri: input.signatureDataURI,
  };
}

export type RuleStatus = 'pass' | 'fail' | 'unknown';

/** How value is checked against threshold: ≤, <, or = (boolean safety fields). */
export type Comparison = 'at_most' | 'below' | 'equals';

export interface RuleResult {
  rule_id: string;
  description: string;
  asme_ref: string;
  /** Display group from the compliance engine's rule catalogue, e.g. "Door Operator". */
  subsystem: string;
  metric: string;
  value: number | null;
  threshold: number;
  comparison: Comparison;
  unit: string;
  status: RuleStatus;
  message: string;
}

export interface ComplianceSummary {
  pass: number;
  fail: number;
  unknown: number;
  overall: RuleStatus;
}

export interface ComplianceResponse {
  unit_tag: string;
  as_of: string;
  stale_window_minutes: number;
  results: RuleResult[];
  summary: ComplianceSummary;
}

export interface SummaryResponse {
  unit_tag: string;
  as_of: string;
  stale_window_minutes: number;
  summary: ComplianceSummary;
}

export interface UnitsResponse {
  units: string[];
}

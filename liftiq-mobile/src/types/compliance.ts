export type RuleStatus = 'pass' | 'fail' | 'unknown';

export interface RuleResult {
  rule_id: string;
  description: string;
  asme_ref: string;
  metric: string;
  value: number | null;
  threshold: number;
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

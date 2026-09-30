import type { Comparison, RuleStatus } from './compliance';

export interface ReportRuleResult {
  rule_id: string;
  description: string;
  asme_ref: string;
  subsystem: string;
  metric: string;
  value: number;
  threshold: number;
  comparison: Comparison;
  unit: string;
  status: RuleStatus;
  message: string;
}

/** Technician's call on a rule and the result status it was made against. */
export interface ReportOverride {
  status: 'pass' | 'fail';
  /** Rejected by the report-generator unless it equals the submitted result's status. */
  against: RuleStatus;
}

export interface ReportPhoto {
  rule_id: string;
  data_uri: string;
  timestamp: string;
  latitude: number | null;
  longitude: number | null;
}

export interface ReportRequest {
  unit_tag: string;
  inspected_at: string;
  technician: string;
  /** Telemetry results exactly as the compliance engine returned them. */
  results: ReportRuleResult[];
  /** rule_id → technician override. The server derives effective status + summary. */
  overrides: Record<string, ReportOverride>;
  photos: ReportPhoto[];
  signature_data_uri?: string;
}

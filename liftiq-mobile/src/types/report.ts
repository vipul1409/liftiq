import type { RuleStatus } from './compliance';

export interface ReportRuleResult {
  rule_id: string;
  description: string;
  asme_ref: string;
  metric: string;
  value: number;
  threshold: number;
  unit: string;
  status: RuleStatus;
  message: string;
  overridden?: boolean;
}

export interface ReportSummary {
  pass: number;
  fail: number;
  unknown: number;
  overall: RuleStatus;
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
  inspected_at: string;   // ISO-8601
  technician: string;
  results: ReportRuleResult[];
  summary: ReportSummary;
  photos: ReportPhoto[];
  signature_data_uri?: string;  // "data:image/svg+xml;base64,..."
}

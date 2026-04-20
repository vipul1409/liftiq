import { apiFetch } from './client';
import type {
  ComplianceResponse,
  SummaryResponse,
  UnitsResponse,
} from '../types/compliance';

export async function getUnits(): Promise<string[]> {
  const data = await apiFetch<UnitsResponse>('/units');
  return data.units ?? [];
}

export async function getCompliance(tag: string): Promise<ComplianceResponse> {
  return apiFetch<ComplianceResponse>(`/units/${encodeURIComponent(tag)}/compliance`);
}

export async function getComplianceSummary(tag: string): Promise<SummaryResponse> {
  return apiFetch<SummaryResponse>(`/units/${encodeURIComponent(tag)}/compliance/summary`);
}

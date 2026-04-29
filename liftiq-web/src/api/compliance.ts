import { apiFetch } from './client';
import type { ComplianceResponse, UnitsResponse } from '../types/compliance';

export async function getUnits(): Promise<string[]> {
  const data = await apiFetch<UnitsResponse>('/units');
  return data.units;
}

export async function getCompliance(tag: string): Promise<ComplianceResponse> {
  return apiFetch<ComplianceResponse>(`/units/${encodeURIComponent(tag)}/compliance`);
}

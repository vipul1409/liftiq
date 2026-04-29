import { ApiError } from './client';
import type { ReportRequest } from '../types/report';

export async function generateReport(req: ReportRequest): Promise<Blob> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 60_000);

  try {
    const res = await fetch('/api/reports/reports', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
      signal: controller.signal,
    });

    if (!res.ok) {
      let message = `HTTP ${res.status}`;
      try {
        const body = await res.json();
        if (body?.error) message = body.error;
      } catch {
        // ignore parse error
      }
      throw new ApiError(res.status, message);
    }

    return await res.blob();
  } finally {
    clearTimeout(timeout);
  }
}

export function downloadPDF(blob: Blob, unitTag: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `liftiq-report-${unitTag}-${Date.now()}.pdf`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

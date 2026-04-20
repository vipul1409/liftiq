import type { ReportRequest } from '../types/report';
import { ApiError } from './client';

const REPORT_API_URL =
  process.env.EXPO_PUBLIC_REPORT_API_URL ?? 'http://localhost:8082';

/**
 * POST /reports — returns the PDF as a base64-encoded string for local saving.
 */
export async function generateReport(req: ReportRequest): Promise<string> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 60_000); // PDF gen can take time

  try {
    const res = await fetch(`${REPORT_API_URL}/reports`, {
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

    // Read the PDF blob and convert to base64 for expo-file-system
    const blob = await res.blob();
    return await blobToBase64(blob);
  } finally {
    clearTimeout(timeout);
  }
}

function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      const result = reader.result as string;
      // Strip "data:application/pdf;base64," prefix
      const base64 = result.split(',')[1];
      resolve(base64);
    };
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}

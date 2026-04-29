import { ApiError } from './client';

export interface KnowledgeResult {
  id: string;
  controller_make: string;
  controller_model: string;
  section: string;
  content: string;
  page_number: number;
  document_name: string;
  similarity: number;
}

export interface KnowledgeResponse {
  query: string;
  answer: string;
  results: KnowledgeResult[];
}

export async function queryKnowledge(
  question: string,
  controllerMake?: string,
  controllerModel?: string,
): Promise<KnowledgeResponse> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 120_000); // LLM can be slow

  try {
    const body: Record<string, unknown> = { question, top_k: 3 };
    if (controllerMake) body.controller_make = controllerMake;
    if (controllerModel) body.controller_model = controllerModel;

    const res = await fetch('/api/knowledge/query', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: controller.signal,
    });

    if (!res.ok) {
      let message = `HTTP ${res.status}`;
      try {
        const err = await res.json();
        if (err?.error) message = err.error;
      } catch {
        // ignore
      }
      throw new ApiError(res.status, message);
    }

    return (await res.json()) as KnowledgeResponse;
  } finally {
    clearTimeout(timeout);
  }
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function apiFetch<T>(path: string): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10_000);

  try {
    const res = await fetch(`/api/compliance${path}`, {
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

    return (await res.json()) as T;
  } finally {
    clearTimeout(timeout);
  }
}

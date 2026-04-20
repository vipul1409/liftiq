import { useState, useEffect, useCallback } from 'react';
import { getCompliance } from '../api/compliance';
import type { ComplianceResponse } from '../types/compliance';

interface UseComplianceResult {
  data: ComplianceResponse | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useCompliance(unitTag: string | null): UseComplianceResult {
  const [data, setData] = useState<ComplianceResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tick, setTick] = useState(0);

  const refetch = useCallback(() => setTick((n) => n + 1), []);

  useEffect(() => {
    if (!unitTag) return;

    let cancelled = false;

    setLoading(true);
    setError(null);

    getCompliance(unitTag)
      .then((res) => {
        if (!cancelled) {
          setData(res);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load compliance data');
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [unitTag, tick]);

  return { data, loading, error, refetch };
}

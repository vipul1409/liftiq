import { useState, useEffect } from 'react';
import { getUnits } from '../api/compliance';

interface UseUnitsResult {
  units: string[];
  loading: boolean;
  error: string | null;
}

export function useUnits(): UseUnitsResult {
  const [units, setUnits] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    setLoading(true);
    setError(null);

    getUnits()
      .then((data) => {
        if (!cancelled) {
          setUnits(data);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load units');
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return { units, loading, error };
}

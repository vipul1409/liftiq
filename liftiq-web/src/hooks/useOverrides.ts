import { useState, useCallback } from 'react';
import type { RuleStatus } from '../types/compliance';

type OverrideStatus = 'pass' | 'fail';

interface UseOverridesResult {
  overrides: Map<string, OverrideStatus>;
  setOverride: (ruleId: string, status: OverrideStatus) => void;
  clearOverride: (ruleId: string) => void;
  clearAll: () => void;
}

export function useOverrides(): UseOverridesResult {
  const [overrides, setOverrides] = useState<Map<string, OverrideStatus>>(new Map());

  const setOverride = useCallback((ruleId: string, status: OverrideStatus) => {
    setOverrides((prev) => new Map(prev).set(ruleId, status));
  }, []);

  const clearOverride = useCallback((ruleId: string) => {
    setOverrides((prev) => {
      const next = new Map(prev);
      next.delete(ruleId);
      return next;
    });
  }, []);

  const clearAll = useCallback(() => {
    setOverrides(new Map());
  }, []);

  return { overrides, setOverride, clearOverride, clearAll };
}

export function effectiveStatus(
  telemetryStatus: RuleStatus,
  override: OverrideStatus | undefined,
): RuleStatus {
  return override ?? telemetryStatus;
}

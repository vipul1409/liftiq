import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import type { ComplianceResponse, ComplianceSummary, RuleResult } from '../types/compliance';
import type { CapturedPhoto } from '../hooks/usePhotos';
import {
  effectiveStatus,
  overridesNeedingReview,
  reconcileOverrides,
  type Overrides,
} from '../utils/overrides';

interface InspectionState {
  unitTag: string | null;
  complianceData: ComplianceResponse | null;
  overrides: Overrides;
  photos: Map<string, CapturedPhoto[]>;
  technician: string;
  signatureDataURI: string;
}

interface InspectionActions {
  setUnitTag: (tag: string) => void;
  setComplianceData: (data: ComplianceResponse) => void;
  /** Record a call against the rule's current result; also re-confirms. */
  setOverride: (ruleId: string, status: 'pass' | 'fail') => void;
  /** Re-confirm an Override whose Rule result has changed. */
  confirmOverride: (ruleId: string) => void;
  clearOverride: (ruleId: string) => void;
  addPhoto: (photo: CapturedPhoto) => void;
  removePhoto: (ruleId: string, uri: string) => void;
  getPhotos: (ruleId: string) => CapturedPhoto[];
  getAllPhotos: () => CapturedPhoto[];
  setSignatureDataURI: (uri: string) => void;
  setTechnician: (name: string) => void;
  getEffectiveResults: () => RuleResult[];
  getEffectiveSummary: () => ComplianceSummary;
  /** Rule IDs whose Override needs re-confirmation; signing is blocked while non-empty. */
  getOverridesNeedingReview: () => string[];
  reset: () => void;
}

type InspectionContextType = InspectionState & InspectionActions;

const InspectionContext = createContext<InspectionContextType | null>(null);

const INITIAL_STATE: InspectionState = {
  unitTag: null,
  complianceData: null,
  overrides: {},
  photos: new Map(),
  technician: 'Inspector',
  signatureDataURI: '',
};

export function InspectionProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<InspectionState>(INITIAL_STATE);

  const setUnitTag = useCallback((tag: string) => {
    setState((prev) => ({ ...prev, unitTag: tag }));
  }, []);

  // A new Compliance snapshot re-anchors Overrides it now agrees with; the rest
  // stay anchored to the result they were made against, so changes get flagged.
  const setComplianceData = useCallback((data: ComplianceResponse) => {
    setState((prev) => ({
      ...prev,
      complianceData: data,
      overrides: reconcileOverrides(prev.overrides, data.results),
    }));
  }, []);

  const setOverride = useCallback((ruleId: string, status: 'pass' | 'fail') => {
    setState((prev) => {
      const rule = prev.complianceData?.results.find((r) => r.rule_id === ruleId);
      if (!rule) return prev;
      return { ...prev, overrides: { ...prev.overrides, [ruleId]: { status, against: rule.status } } };
    });
  }, []);

  const confirmOverride = useCallback(
    (ruleId: string) => {
      const o = state.overrides[ruleId];
      if (o) setOverride(ruleId, o.status);
    },
    [state.overrides, setOverride],
  );

  const clearOverride = useCallback((ruleId: string) => {
    setState((prev) => {
      const next = { ...prev.overrides };
      delete next[ruleId];
      return { ...prev, overrides: next };
    });
  }, []);

  const addPhoto = useCallback((photo: CapturedPhoto) => {
    setState((prev) => {
      const next = new Map(prev.photos);
      const existing = next.get(photo.ruleId) ?? [];
      next.set(photo.ruleId, [...existing, photo]);
      return { ...prev, photos: next };
    });
  }, []);

  const removePhoto = useCallback((ruleId: string, uri: string) => {
    setState((prev) => {
      const next = new Map(prev.photos);
      const existing = next.get(ruleId) ?? [];
      next.set(ruleId, existing.filter((p) => p.uri !== uri));
      return { ...prev, photos: next };
    });
  }, []);

  const getPhotos = useCallback(
    (ruleId: string): CapturedPhoto[] => state.photos.get(ruleId) ?? [],
    [state.photos],
  );

  const getAllPhotos = useCallback(
    (): CapturedPhoto[] => Array.from(state.photos.values()).flat(),
    [state.photos],
  );

  const setSignatureDataURI = useCallback((uri: string) => {
    setState((prev) => ({ ...prev, signatureDataURI: uri }));
  }, []);

  const setTechnician = useCallback((name: string) => {
    setState((prev) => ({ ...prev, technician: name }));
  }, []);

  const getEffectiveResults = useCallback((): RuleResult[] => {
    if (!state.complianceData) return [];
    return state.complianceData.results.map((r) => {
      const status = effectiveStatus(r.status, state.overrides[r.rule_id]);
      return status === r.status ? r : { ...r, status };
    });
  }, [state.complianceData, state.overrides]);

  const getEffectiveSummary = useCallback((): ComplianceSummary => {
    const results = getEffectiveResults();
    let pass = 0, fail = 0, unknown = 0;
    for (const r of results) {
      if (r.status === 'pass') pass++;
      else if (r.status === 'fail') fail++;
      else unknown++;
    }
    const overall = fail > 0 ? 'fail' : unknown > 0 ? 'unknown' : 'pass';
    return { pass, fail, unknown, overall } as ComplianceSummary;
  }, [getEffectiveResults]);

  const getOverridesNeedingReview = useCallback(
    (): string[] =>
      state.complianceData ? overridesNeedingReview(state.overrides, state.complianceData.results) : [],
    [state.complianceData, state.overrides],
  );

  const reset = useCallback(() => {
    setState(INITIAL_STATE);
  }, []);

  const value: InspectionContextType = {
    ...state,
    setUnitTag,
    setComplianceData,
    setOverride,
    confirmOverride,
    clearOverride,
    addPhoto,
    removePhoto,
    getPhotos,
    getAllPhotos,
    setSignatureDataURI,
    setTechnician,
    getEffectiveResults,
    getEffectiveSummary,
    getOverridesNeedingReview,
    reset,
  };

  return (
    <InspectionContext.Provider value={value}>
      {children}
    </InspectionContext.Provider>
  );
}

export function useInspection(): InspectionContextType {
  const ctx = useContext(InspectionContext);
  if (!ctx) throw new Error('useInspection must be used within InspectionProvider');
  return ctx;
}

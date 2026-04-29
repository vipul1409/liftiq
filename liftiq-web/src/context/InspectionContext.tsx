import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import type { ComplianceResponse, ComplianceSummary, RuleResult } from '../types/compliance';
import type { CapturedPhoto } from '../hooks/usePhotos';

interface InspectionState {
  unitTag: string | null;
  complianceData: ComplianceResponse | null;
  overrides: Map<string, 'pass' | 'fail'>;
  photos: Map<string, CapturedPhoto[]>;
  technician: string;
  signatureDataURI: string;
}

interface InspectionActions {
  setUnitTag: (tag: string) => void;
  setComplianceData: (data: ComplianceResponse) => void;
  setOverride: (ruleId: string, status: 'pass' | 'fail') => void;
  clearOverride: (ruleId: string) => void;
  addPhoto: (photo: CapturedPhoto) => void;
  removePhoto: (ruleId: string, uri: string) => void;
  getPhotos: (ruleId: string) => CapturedPhoto[];
  getAllPhotos: () => CapturedPhoto[];
  setSignatureDataURI: (uri: string) => void;
  setTechnician: (name: string) => void;
  getEffectiveResults: () => RuleResult[];
  getEffectiveSummary: () => ComplianceSummary;
  reset: () => void;
}

type InspectionContextType = InspectionState & InspectionActions;

const InspectionContext = createContext<InspectionContextType | null>(null);

const INITIAL_STATE: InspectionState = {
  unitTag: null,
  complianceData: null,
  overrides: new Map(),
  photos: new Map(),
  technician: 'Inspector',
  signatureDataURI: '',
};

export function InspectionProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<InspectionState>(INITIAL_STATE);

  const setUnitTag = useCallback((tag: string) => {
    setState((prev) => ({ ...prev, unitTag: tag }));
  }, []);

  const setComplianceData = useCallback((data: ComplianceResponse) => {
    setState((prev) => ({ ...prev, complianceData: data }));
  }, []);

  const setOverride = useCallback((ruleId: string, status: 'pass' | 'fail') => {
    setState((prev) => ({
      ...prev,
      overrides: new Map(prev.overrides).set(ruleId, status),
    }));
  }, []);

  const clearOverride = useCallback((ruleId: string) => {
    setState((prev) => {
      const next = new Map(prev.overrides);
      next.delete(ruleId);
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
      const override = state.overrides.get(r.rule_id);
      if (override) return { ...r, status: override };
      return r;
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

  const reset = useCallback(() => {
    setState(INITIAL_STATE);
  }, []);

  const value: InspectionContextType = {
    ...state,
    setUnitTag,
    setComplianceData,
    setOverride,
    clearOverride,
    addPhoto,
    removePhoto,
    getPhotos,
    getAllPhotos,
    setSignatureDataURI,
    setTechnician,
    getEffectiveResults,
    getEffectiveSummary,
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

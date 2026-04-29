import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router';
import { RuleRow } from '../components/RuleRow';
import { StatusBadge } from '../components/StatusBadge';
import { VoiceBar } from '../components/VoiceBar';
import { KnowledgePanel } from '../components/KnowledgePanel';
import { useCompliance } from '../hooks/useCompliance';
import { useInspection } from '../context/InspectionContext';
import { effectiveStatus } from '../hooks/useOverrides';
import { fileToDataURI } from '../hooks/usePhotos';
import { useVoice } from '../hooks/useVoice';
import type { RuleResult } from '../types/compliance';
import type { VoiceIntent } from '../utils/intentParser';

interface Section {
  title: string;
  data: RuleResult[];
}

const SUBSYSTEM_MAP: { title: string; prefix: string[] }[] = [
  { title: 'Motor', prefix: ['ASME-001', 'ASME-002', 'ASME-003', 'ASME-004'] },
  { title: 'Trip / Usage', prefix: ['ASME-005'] },
  { title: 'Door Operator', prefix: ['ASME-006', 'ASME-007', 'ASME-008', 'ASME-009', 'ASME-010'] },
  { title: 'Brake System', prefix: ['ASME-011', 'ASME-012', 'ASME-013'] },
  { title: 'Ride Quality', prefix: ['ASME-014', 'ASME-015'] },
  { title: 'Safety Circuits', prefix: ['ASME-016', 'ASME-017', 'ASME-018', 'ASME-019', 'ASME-020'] },
];

function groupResults(results: RuleResult[]): Section[] {
  const byId = new Map(results.map((r) => [r.rule_id, r]));
  return SUBSYSTEM_MAP.map(({ title, prefix }) => ({
    title,
    data: prefix.map((id) => byId.get(id)).filter((r): r is RuleResult => r !== undefined),
  })).filter((s) => s.data.length > 0);
}

export function ComplianceScreen() {
  const { tag } = useParams<{ tag: string }>();
  const navigate = useNavigate();
  const { data, loading, error, refetch } = useCompliance(tag ?? null);
  const inspection = useInspection();

  // Sync compliance data to context whenever it changes
  useEffect(() => {
    if (data) inspection.setComplianceData(data);
  }, [data]); // eslint-disable-line react-hooks/exhaustive-deps

  const sections = useMemo(
    () => (data ? groupResults(data.results) : []),
    [data],
  );

  const flatRules = useMemo(() => sections.flatMap((s) => s.data), [sections]);

  const [activeRuleId, setActiveRuleId] = useState<string | null>(null);

  useEffect(() => {
    if (flatRules.length > 0 && activeRuleId === null) {
      setActiveRuleId(flatRules[0].rule_id);
    }
  }, [flatRules]); // eslint-disable-line react-hooks/exhaustive-deps

  const [readText, setReadText] = useState('');
  const advancedByVoice = useRef(false);

  const advanceCursor = useCallback(() => {
    setActiveRuleId((current) => {
      const idx = flatRules.findIndex((r) => r.rule_id === current);
      const next = flatRules[idx + 1];
      return next ? next.rule_id : current;
    });
    advancedByVoice.current = true;
  }, [flatRules]);

  useEffect(() => {
    if (!advancedByVoice.current || !activeRuleId) return;
    advancedByVoice.current = false;
    const rule = flatRules.find((r) => r.rule_id === activeRuleId);
    if (!rule) return;
    const status = effectiveStatus(rule.status, inspection.overrides.get(rule.rule_id));
    setReadText(`${rule.rule_id}: ${rule.description}. Status: ${status}.`);
  }, [activeRuleId]); // eslint-disable-line react-hooks/exhaustive-deps

  // Photo capture trigger from voice
  const capturePhotoRef = useRef<HTMLInputElement>(null);
  const activeRuleIdRef = useRef(activeRuleId);
  activeRuleIdRef.current = activeRuleId;

  const handleIntent = useCallback(
    (intent: VoiceIntent) => {
      const ruleId = activeRuleIdRef.current;
      if (!ruleId) return;
      switch (intent) {
        case 'pass':
          inspection.setOverride(ruleId, 'pass');
          setReadText('Marked pass.');
          break;
        case 'fail':
          inspection.setOverride(ruleId, 'fail');
          setReadText('Marked fail.');
          break;
        case 'skip':
          setReadText('Skipped.');
          advanceCursor();
          break;
        case 'next':
          advanceCursor();
          break;
        case 'photo':
          setReadText('Opening camera.');
          capturePhotoRef.current?.click();
          break;
        case 'stop':
          setReadText('Voice stopped.');
          break;
        case 'unknown':
          setReadText("Didn't catch that.");
          break;
      }
    },
    [inspection, advanceCursor],
  );

  const voice = useVoice({
    onIntent: handleIntent,
    readText,
    enabled: !loading && !!data,
  });

  const summary = inspection.getEffectiveSummary();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
        <p className="text-gray-500 mt-3">Loading compliance data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-8">
        <h2 className="text-lg font-bold text-gray-900 mb-2">Failed to load</h2>
        <p className="text-gray-500 text-center mb-4">{error}</p>
        <button
          onClick={refetch}
          className="px-8 py-3 bg-blue-600 text-white font-bold rounded-lg hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className="min-h-screen bg-gray-50 pb-40">
      {/* Header */}
      <div className="bg-gray-50 px-4 pt-4 pb-2 max-w-3xl mx-auto">
        <div className="flex justify-between items-center">
          <h1 className="text-[22px] font-extrabold text-blue-800">{tag}</h1>
          <button
            onClick={refetch}
            className="px-3 py-1.5 text-[13px] font-semibold text-indigo-800 bg-indigo-100 rounded-lg hover:bg-indigo-200"
          >
            Refresh
          </button>
        </div>
        <p className="text-xs text-gray-400 mt-0.5">
          As of {new Date(data.as_of).toLocaleString()}
        </p>
      </div>

      {/* Knowledge Base Search */}
      <KnowledgePanel />

      {/* Sections */}
      <div className="max-w-3xl mx-auto">
        {sections.map((section) => (
          <div key={section.title}>
            <div className="sticky top-0 z-[5] bg-gray-100 px-4 py-2 border-y border-gray-200">
              <h2 className="text-[13px] font-bold text-gray-600 uppercase tracking-wider">
                {section.title}
              </h2>
            </div>
            {section.data.map((rule, i) => (
              <div key={rule.rule_id}>
                {i > 0 && <div className="h-px bg-gray-100 ml-4" />}
                <RuleRow
                  result={rule}
                  override={inspection.overrides.get(rule.rule_id)}
                  onOverride={inspection.setOverride}
                  onClearOverride={inspection.clearOverride}
                  isActive={rule.rule_id === activeRuleId}
                  photos={inspection.getPhotos(rule.rule_id)}
                  onAddPhoto={inspection.addPhoto}
                  onRemovePhoto={inspection.removePhoto}
                />
              </div>
            ))}
          </div>
        ))}
      </div>

      {/* Hidden file input for voice-triggered photo capture */}
      <input
        ref={capturePhotoRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={async (e) => {
          const file = e.target.files?.[0];
          if (!file || !activeRuleId) return;
          const dataUri = await fileToDataURI(file);
          inspection.addPhoto({
            uri: dataUri,
            ruleId: activeRuleId,
            timestamp: new Date().toISOString(),
            latitude: null,
            longitude: null,
          });
          e.target.value = '';
        }}
      />

      {/* Voice Bar */}
      <VoiceBar
        listening={voice.listening}
        isSpeaking={voice.isSpeaking}
        transcript={voice.transcript}
        lastIntent={voice.lastIntent}
        hasPermission={voice.hasPermission}
        onStartListening={voice.startListening}
        onStopListening={voice.stopListening}
      />

      {/* Sticky footer */}
      <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-4 py-3 z-20">
        <div className="max-w-3xl mx-auto flex items-center justify-between">
          <div className="text-[13px]">
            <span className="text-green-600 font-bold">{summary.pass} pass</span>
            {'  '}
            <span className="text-red-600 font-bold">{summary.fail} fail</span>
            {'  '}
            <span className="text-gray-400 font-semibold">{summary.unknown} unknown</span>
          </div>
          <button
            onClick={() => navigate('/summary')}
            className="flex items-center gap-2 bg-blue-800 text-white px-4 py-2.5 rounded-lg font-bold text-sm hover:bg-blue-900 transition-colors"
          >
            <StatusBadge status={summary.overall} />
            <span>View Summary</span>
          </button>
        </div>
      </div>
    </div>
  );
}

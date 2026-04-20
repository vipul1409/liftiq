import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ActivityIndicator,
  SafeAreaView,
  SectionList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RuleRow } from '../components/RuleRow';
import { StatusBadge } from '../components/StatusBadge';
import { VoiceBar } from '../components/VoiceBar';
import { useCompliance } from '../hooks/useCompliance';
import { useOverrides } from '../store/overrides';
import { usePhotos } from '../store/photos';
import { useVoice } from '../hooks/useVoice';
import { useCamera } from '../hooks/useCamera';
import { effectiveStatus } from '../store/overrides';
import type { RootStackParamList } from '../navigation/AppNavigator';
import type { RuleResult } from '../types/compliance';
import type { VoiceIntent } from '../types/voice';

type Props = NativeStackScreenProps<RootStackParamList, 'Compliance'>;

interface Section {
  title: string;
  data: RuleResult[];
}

const SUBSYSTEM_MAP: { title: string; prefix: string[] }[] = [
  { title: 'Motor', prefix: ['ASME-001', 'ASME-002', 'ASME-003', 'ASME-004'] },
  { title: 'Trip / Usage', prefix: ['ASME-005'] },
  {
    title: 'Door Operator',
    prefix: ['ASME-006', 'ASME-007', 'ASME-008', 'ASME-009', 'ASME-010'],
  },
  { title: 'Brake System', prefix: ['ASME-011', 'ASME-012', 'ASME-013'] },
  { title: 'Ride Quality', prefix: ['ASME-014', 'ASME-015'] },
  {
    title: 'Safety Circuits',
    prefix: ['ASME-016', 'ASME-017', 'ASME-018', 'ASME-019', 'ASME-020'],
  },
];

function groupResults(results: RuleResult[]): Section[] {
  const byId = new Map(results.map((r) => [r.rule_id, r]));
  return SUBSYSTEM_MAP.map(({ title, prefix }) => ({
    title,
    data: prefix.map((id) => byId.get(id)).filter((r): r is RuleResult => r !== undefined),
  })).filter((s) => s.data.length > 0);
}

export function ComplianceScreen({ route, navigation }: Props) {
  const { unitTag } = route.params;
  const { data, loading, error, refetch } = useCompliance(unitTag);
  const { overrides, setOverride, clearOverride } = useOverrides();
  const { addPhoto, removePhoto, getPhotos } = usePhotos();
  const camera = useCamera();

  const sections = useMemo(
    () => (data ? groupResults(data.results) : []),
    [data],
  );

  // Flat ordered rule list for cursor navigation.
  const flatRules = useMemo(() => sections.flatMap((s) => s.data), [sections]);

  // Active rule cursor (voice navigation target).
  const [activeRuleId, setActiveRuleId] = useState<string | null>(null);

  // Seed cursor to first rule once data loads.
  useEffect(() => {
    if (flatRules.length > 0 && activeRuleId === null) {
      setActiveRuleId(flatRules[0].rule_id);
    }
  }, [flatRules]); // eslint-disable-line react-hooks/exhaustive-deps

  // Text for TTS readback. Only updated when voice navigation fires.
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

  // Speak the active rule whenever voice navigation moves the cursor.
  useEffect(() => {
    if (!advancedByVoice.current || !activeRuleId) return;
    advancedByVoice.current = false;
    const rule = flatRules.find((r) => r.rule_id === activeRuleId);
    if (!rule) return;
    const status = effectiveStatus(rule.status, overrides.get(rule.rule_id));
    setReadText(`${rule.rule_id}: ${rule.description}. Status: ${status}.`);
  }, [activeRuleId]); // eslint-disable-line react-hooks/exhaustive-deps

  // Capture a photo for a rule and store it.
  const capturePhoto = useCallback(async (ruleId: string) => {
    const photo = await camera.capture(ruleId);
    if (photo) addPhoto(photo);
  }, [camera, addPhoto]);

  const handleIntent = useCallback(
    (intent: VoiceIntent) => {
      if (!activeRuleId) return;
      switch (intent) {
        case 'pass':
          setOverride(activeRuleId, 'pass');
          setReadText('Marked pass.');
          break;
        case 'fail':
          setOverride(activeRuleId, 'fail');
          setReadText('Marked fail. Opening camera for evidence.');
          capturePhoto(activeRuleId);
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
          capturePhoto(activeRuleId);
          break;
        case 'stop':
          setReadText('Voice stopped.');
          break;
        case 'unknown':
          setReadText("Didn't catch that.");
          break;
      }
    },
    [activeRuleId, setOverride, advanceCursor, capturePhoto],
  );

  const voice = useVoice({
    onIntent: handleIntent,
    readText,
    enabled: !loading && !!data,
  });

  if (loading) {
    return (
      <SafeAreaView style={styles.centered}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.loadingText}>Loading compliance data…</Text>
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.centered}>
        <Text style={styles.errorTitle}>Failed to load</Text>
        <Text style={styles.errorDetail}>{error}</Text>
        <TouchableOpacity style={styles.retryButton} onPress={refetch}>
          <Text style={styles.retryButtonText}>Retry</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  if (!data) return null;

  return (
    <SafeAreaView style={styles.container}>
      <SectionList
        sections={sections}
        keyExtractor={(item) => item.rule_id}
        renderItem={({ item }) => (
          <RuleRow
            result={item}
            override={overrides.get(item.rule_id)}
            onOverride={setOverride}
            onClearOverride={clearOverride}
            isActive={item.rule_id === activeRuleId}
            photos={getPhotos(item.rule_id)}
            onCapturePhoto={camera.hasPermission ? capturePhoto : undefined}
            onRemovePhoto={removePhoto}
          />
        )}
        renderSectionHeader={({ section }) => (
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>{section.title}</Text>
          </View>
        )}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
        ListHeaderComponent={
          <View style={styles.listHeader}>
            <View style={styles.listHeaderRow}>
              <Text style={styles.unitTag}>{unitTag}</Text>
              <TouchableOpacity onPress={refetch} style={styles.refreshBtn}>
                <Text style={styles.refreshBtnText}>Refresh</Text>
              </TouchableOpacity>
            </View>
            <Text style={styles.asOf}>
              As of {new Date(data.as_of).toLocaleString()}
            </Text>
          </View>
        }
        ListFooterComponent={<View style={{ height: 160 }} />}
        stickySectionHeadersEnabled
      />

      <VoiceBar
        listening={voice.listening}
        isSpeaking={voice.isSpeaking}
        transcript={voice.transcript}
        lastIntent={voice.lastIntent}
        hasPermission={voice.hasPermission}
        onStartListening={voice.startListening}
        onStopListening={voice.stopListening}
      />

      {/* Sticky footer with summary + view button */}
      <View style={styles.footer}>
        <View style={styles.footerCounts}>
          <Text style={styles.footerCount}>
            <Text style={styles.countPass}>{data.summary.pass} pass</Text>
            {'  '}
            <Text style={styles.countFail}>{data.summary.fail} fail</Text>
            {'  '}
            <Text style={styles.countUnknown}>{data.summary.unknown} unknown</Text>
          </Text>
        </View>
        <TouchableOpacity
          style={styles.summaryBtn}
          onPress={() =>
            navigation.navigate('Summary', {
              unitTag,
              summary: data.summary,
              asOf: data.as_of,
            })
          }
        >
          <StatusBadge status={data.summary.overall} />
          <Text style={styles.summaryBtnText}>View Summary</Text>
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 32,
    backgroundColor: '#f9fafb',
  },
  loadingText: {
    marginTop: 12,
    color: '#6b7280',
    fontSize: 15,
  },
  errorTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 8,
  },
  errorDetail: {
    fontSize: 14,
    color: '#6b7280',
    textAlign: 'center',
    marginBottom: 20,
  },
  retryButton: {
    backgroundColor: '#2563eb',
    paddingVertical: 12,
    paddingHorizontal: 32,
    borderRadius: 8,
  },
  retryButtonText: {
    color: '#fff',
    fontWeight: '700',
    fontSize: 15,
  },
  listHeader: {
    padding: 16,
    paddingBottom: 8,
    backgroundColor: '#f9fafb',
  },
  listHeaderRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  unitTag: {
    fontSize: 22,
    fontWeight: '800',
    color: '#1e40af',
  },
  refreshBtn: {
    paddingVertical: 6,
    paddingHorizontal: 12,
    backgroundColor: '#e0e7ff',
    borderRadius: 8,
  },
  refreshBtnText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#3730a3',
  },
  asOf: {
    fontSize: 12,
    color: '#9ca3af',
    marginTop: 2,
  },
  sectionHeader: {
    backgroundColor: '#f3f4f6',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderTopWidth: 1,
    borderBottomWidth: 1,
    borderColor: '#e5e7eb',
  },
  sectionTitle: {
    fontSize: 13,
    fontWeight: '700',
    color: '#374151',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
  },
  separator: {
    height: 1,
    backgroundColor: '#f3f4f6',
    marginLeft: 16,
  },
  footer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderColor: '#e5e7eb',
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 28,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  footerCounts: {
    flex: 1,
  },
  footerCount: {
    fontSize: 13,
  },
  countPass: {
    color: '#16a34a',
    fontWeight: '700',
  },
  countFail: {
    color: '#dc2626',
    fontWeight: '700',
  },
  countUnknown: {
    color: '#9ca3af',
    fontWeight: '600',
  },
  summaryBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: '#1e40af',
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 10,
  },
  summaryBtnText: {
    color: '#fff',
    fontWeight: '700',
    fontSize: 14,
  },
});

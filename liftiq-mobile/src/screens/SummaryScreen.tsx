import React from 'react';
import {
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { RootStackParamList } from '../navigation/AppNavigator';
import type { RuleStatus } from '../types/compliance';
import { applyOverrides, summarise } from '../utils/inspectionOutcome';

type Props = NativeStackScreenProps<RootStackParamList, 'Summary'>;

const BANNER_COLORS: Record<RuleStatus, string> = {
  pass: '#16a34a',
  fail: '#dc2626',
  unknown: '#6b7280',
};

const BANNER_LABELS: Record<RuleStatus, string> = {
  pass: 'ALL CHECKS PASSED',
  fail: 'INSPECTION FAILED',
  unknown: 'INCOMPLETE DATA',
};

export function SummaryScreen({ route, navigation }: Props) {
  const { unitTag, asOf, results, overrides, photos, technician } = route.params;
  const summary = summarise(applyOverrides(results, overrides));
  const bannerColor = BANNER_COLORS[summary.overall];
  const bannerLabel = BANNER_LABELS[summary.overall];
  const formattedDate = new Date(asOf).toLocaleString();

  return (
    <SafeAreaView style={styles.container}>
      <View style={[styles.banner, { backgroundColor: bannerColor }]}>
        <Text style={styles.bannerUnit}>{unitTag}</Text>
        <Text style={styles.bannerLabel}>{bannerLabel}</Text>
        <Text style={styles.bannerDate}>{formattedDate}</Text>
      </View>

      <ScrollView contentContainerStyle={styles.body}>
        <View style={styles.countsRow}>
          <View style={[styles.countCard, styles.countPass]}>
            <Text style={styles.countNumber}>{summary.pass}</Text>
            <Text style={styles.countLabel}>Pass</Text>
          </View>
          <View style={[styles.countCard, styles.countFail]}>
            <Text style={styles.countNumber}>{summary.fail}</Text>
            <Text style={styles.countLabel}>Fail</Text>
          </View>
          <View style={[styles.countCard, styles.countUnknown]}>
            <Text style={styles.countNumber}>{summary.unknown}</Text>
            <Text style={styles.countLabel}>Unknown</Text>
          </View>
        </View>

        <Text style={styles.meta}>
          {summary.pass + summary.fail + summary.unknown} ASME A17.1 checks evaluated
        </Text>

        {photos.length > 0 && (
          <Text style={styles.photoCount}>
            {photos.length} photo{photos.length !== 1 ? 's' : ''} attached
          </Text>
        )}
      </ScrollView>

      <TouchableOpacity
        style={styles.proceedButton}
        onPress={() =>
          navigation.navigate('Signature', {
            unitTag,
            asOf,
            results,
            overrides,
            photos,
            technician,
          })
        }
      >
        <Text style={styles.proceedButtonText}>Proceed to Sign</Text>
      </TouchableOpacity>

      <TouchableOpacity
        style={styles.backButton}
        onPress={() => navigation.goBack()}
      >
        <Text style={styles.backButtonText}>Back to Checklist</Text>
      </TouchableOpacity>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  banner: {
    paddingVertical: 40,
    paddingHorizontal: 24,
    alignItems: 'center',
  },
  bannerUnit: {
    fontSize: 16,
    fontWeight: '600',
    color: 'rgba(255,255,255,0.8)',
    marginBottom: 8,
    letterSpacing: 1,
  },
  bannerLabel: {
    fontSize: 28,
    fontWeight: '800',
    color: '#fff',
    letterSpacing: 1,
    marginBottom: 8,
  },
  bannerDate: {
    fontSize: 13,
    color: 'rgba(255,255,255,0.7)',
  },
  body: {
    padding: 24,
    alignItems: 'center',
  },
  countsRow: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
    width: '100%',
  },
  countCard: {
    flex: 1,
    borderRadius: 12,
    paddingVertical: 20,
    alignItems: 'center',
  },
  countPass: {
    backgroundColor: '#dcfce7',
  },
  countFail: {
    backgroundColor: '#fee2e2',
  },
  countUnknown: {
    backgroundColor: '#f3f4f6',
  },
  countNumber: {
    fontSize: 32,
    fontWeight: '800',
    color: '#111827',
  },
  countLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6b7280',
    marginTop: 4,
  },
  meta: {
    fontSize: 13,
    color: '#9ca3af',
    textAlign: 'center',
    marginBottom: 8,
  },
  photoCount: {
    fontSize: 13,
    color: '#6b7280',
    textAlign: 'center',
    marginTop: 4,
  },
  proceedButton: {
    marginHorizontal: 16,
    marginBottom: 8,
    backgroundColor: '#2563eb',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
  },
  proceedButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#fff',
  },
  backButton: {
    marginHorizontal: 16,
    marginBottom: 16,
    backgroundColor: '#1e40af',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
  },
  backButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#fff',
  },
});

import React from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { StatusBadge } from './StatusBadge';
import { effectiveStatus } from '../store/overrides';
import type { RuleResult } from '../types/compliance';

interface Props {
  result: RuleResult;
  override: 'pass' | 'fail' | undefined;
  onOverride: (ruleId: string, status: 'pass' | 'fail') => void;
  onClearOverride: (ruleId: string) => void;
  isActive?: boolean;
}

export function RuleRow({ result, override, onOverride, onClearOverride, isActive = false }: Props) {
  const effective = effectiveStatus(result.status, override);

  return (
    <View style={[styles.container, isActive && styles.containerActive]}>
      <View style={styles.topRow}>
        <Text style={styles.ruleId}>{result.rule_id}</Text>
        <StatusBadge status={effective} overridden={override !== undefined} />
      </View>
      <Text style={styles.description}>{result.description}</Text>
      <Text style={styles.asmeRef}>{result.asme_ref}</Text>
      {result.value !== null && (
        <Text style={styles.value}>
          {result.value} {result.unit} (threshold: {result.threshold} {result.unit})
        </Text>
      )}
      <Text style={styles.message}>{result.message}</Text>

      <View style={styles.buttons}>
        <TouchableOpacity
          style={[styles.btn, styles.btnPass, override === 'pass' && styles.btnActive]}
          onPress={() => onOverride(result.rule_id, 'pass')}
        >
          <Text style={styles.btnText}>Override Pass</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.btn, styles.btnFail, override === 'fail' && styles.btnActive]}
          onPress={() => onOverride(result.rule_id, 'fail')}
        >
          <Text style={styles.btnText}>Override Fail</Text>
        </TouchableOpacity>
        {override !== undefined && (
          <TouchableOpacity
            style={[styles.btn, styles.btnClear]}
            onPress={() => onClearOverride(result.rule_id)}
          >
            <Text style={styles.btnText}>Clear</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 14,
    backgroundColor: '#fff',
  },
  containerActive: {
    borderLeftWidth: 3,
    borderLeftColor: '#2563eb',
    backgroundColor: '#eff6ff',
    paddingLeft: 13, // compensate for the 3px border
  },
  topRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  ruleId: {
    fontSize: 11,
    color: '#9ca3af',
    fontWeight: '600',
    letterSpacing: 0.5,
  },
  description: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 2,
  },
  asmeRef: {
    fontSize: 12,
    color: '#6b7280',
    fontStyle: 'italic',
    marginBottom: 4,
  },
  value: {
    fontSize: 13,
    color: '#374151',
    marginBottom: 2,
  },
  message: {
    fontSize: 13,
    color: '#6b7280',
    marginBottom: 8,
  },
  buttons: {
    flexDirection: 'row',
    gap: 8,
  },
  btn: {
    paddingVertical: 5,
    paddingHorizontal: 12,
    borderRadius: 6,
    borderWidth: 1,
  },
  btnPass: {
    borderColor: '#86efac',
    backgroundColor: '#f0fdf4',
  },
  btnFail: {
    borderColor: '#fca5a5',
    backgroundColor: '#fef2f2',
  },
  btnClear: {
    borderColor: '#d1d5db',
    backgroundColor: '#f9fafb',
  },
  btnActive: {
    opacity: 0.5,
  },
  btnText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#374151',
  },
});

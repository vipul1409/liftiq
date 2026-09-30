import React from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { StatusBadge } from './StatusBadge';
import { PhotoStrip } from './PhotoStrip';
import { effectiveStatus, needsReview, type Override } from '../store/overrides';
import type { RuleResult } from '../types/compliance';
import type { CapturedPhoto } from '../store/photos';

interface Props {
  result: RuleResult;
  override: Override | undefined;
  onOverride: (ruleId: string, status: 'pass' | 'fail') => void;
  onClearOverride: (ruleId: string) => void;
  /** Re-confirm an Override whose Rule result has changed. */
  onConfirmOverride: (ruleId: string) => void;
  isActive?: boolean;
  photos?: CapturedPhoto[];
  onCapturePhoto?: (ruleId: string) => void;
  onRemovePhoto?: (ruleId: string, uri: string) => void;
}

export function RuleRow({
  result,
  override,
  onOverride,
  onClearOverride,
  onConfirmOverride,
  isActive = false,
  photos = [],
  onCapturePhoto,
  onRemovePhoto,
}: Props) {
  const effective = effectiveStatus(result.status, override);
  const review = override !== undefined && needsReview(override, result.status);

  return (
    <View style={[styles.container, isActive && styles.containerActive]}>
      <View style={styles.topRow}>
        <Text style={styles.ruleId}>{result.rule_id}</Text>
        <StatusBadge status={effective} overridden={override !== undefined && !review} />
      </View>
      {review && (
        <View style={styles.review}>
          <Text style={styles.reviewText}>
            Result changed: was {override.against}, now {result.status}. Your override ({override.status}) needs confirming.
          </Text>
          <View style={styles.buttons}>
            <TouchableOpacity
              style={[styles.btn, styles.btnConfirm]}
              onPress={() => onConfirmOverride(result.rule_id)}
            >
              <Text style={styles.btnText}>Confirm {override.status}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.btn, styles.btnClear]}
              onPress={() => onClearOverride(result.rule_id)}
            >
              <Text style={styles.btnText}>Clear</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
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
          style={[styles.btn, styles.btnPass, override?.status === 'pass' && styles.btnActive]}
          onPress={() => onOverride(result.rule_id, 'pass')}
        >
          <Text style={styles.btnText}>Override Pass</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.btn, styles.btnFail, override?.status === 'fail' && styles.btnActive]}
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
        {onCapturePhoto && (
          <TouchableOpacity
            style={[styles.btn, styles.btnPhoto]}
            onPress={() => onCapturePhoto(result.rule_id)}
          >
            <Text style={styles.btnText}>📷 Photo</Text>
          </TouchableOpacity>
        )}
      </View>

      {onRemovePhoto && (
        <PhotoStrip
          photos={photos}
          onRemove={(uri) => onRemovePhoto(result.rule_id, uri)}
        />
      )}
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
  btnPhoto: {
    borderColor: '#93c5fd',
    backgroundColor: '#eff6ff',
  },
  btnConfirm: {
    borderColor: '#fcd34d',
    backgroundColor: '#fffbeb',
  },
  review: {
    backgroundColor: '#fffbeb',
    borderColor: '#fcd34d',
    borderWidth: 1,
    borderRadius: 6,
    padding: 8,
    marginBottom: 8,
    gap: 6,
  },
  reviewText: {
    fontSize: 13,
    color: '#92400e',
    fontWeight: '600',
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

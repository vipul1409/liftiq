import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import type { RuleStatus } from '../types/compliance';

interface Props {
  status: RuleStatus;
  overridden?: boolean;
}

const COLORS: Record<RuleStatus, { bg: string; text: string }> = {
  pass: { bg: '#22c55e', text: '#fff' },
  fail: { bg: '#ef4444', text: '#fff' },
  unknown: { bg: '#9ca3af', text: '#fff' },
};

export function StatusBadge({ status, overridden = false }: Props) {
  const { bg, text } = COLORS[status];
  return (
    <View style={[styles.badge, { backgroundColor: bg }]}>
      <Text style={[styles.label, { color: text }]}>
        {status.toUpperCase()}
        {overridden ? ' (manual)' : ''}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 4,
    alignSelf: 'flex-start',
  },
  label: {
    fontSize: 11,
    fontWeight: '700',
    letterSpacing: 0.5,
  },
});

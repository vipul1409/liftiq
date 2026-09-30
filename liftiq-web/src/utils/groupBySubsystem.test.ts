import { describe, expect, it } from 'vitest';
import { groupBySubsystem } from './groupBySubsystem';
import type { RuleResult } from '../types/compliance';

function rule(id: string, subsystem: string): RuleResult {
  return {
    rule_id: id,
    description: id,
    asme_ref: 'ASME A17.1',
    subsystem,
    metric: 'm',
    value: 1,
    threshold: 1,
    comparison: 'at_most',
    unit: 'A',
    status: 'pass',
    message: '',
  };
}

describe('groupBySubsystem', () => {
  it('groups in the order subsystems first appear', () => {
    const sections = groupBySubsystem([
      rule('ASME-001', 'Motor'),
      rule('ASME-002', 'Motor'),
      rule('ASME-006', 'Door Operator'),
    ]);
    expect(sections.map((s) => s.title)).toEqual(['Motor', 'Door Operator']);
    expect(sections[0].data.map((r) => r.rule_id)).toEqual(['ASME-001', 'ASME-002']);
  });

  it('shows a rule for a subsystem the app has never seen', () => {
    const sections = groupBySubsystem([rule('ASME-021', 'Hoistway')]);
    expect(sections).toEqual([{ title: 'Hoistway', data: [expect.objectContaining({ rule_id: 'ASME-021' })] }]);
  });

  it('puts rules with no subsystem under Other, last', () => {
    const sections = groupBySubsystem([rule('ASME-099', ''), rule('ASME-001', 'Motor')]);
    expect(sections.map((s) => s.title)).toEqual(['Motor', 'Other']);
    expect(sections[1].data[0].rule_id).toBe('ASME-099');
  });

  it('returns no sections for no results', () => {
    expect(groupBySubsystem([])).toEqual([]);
  });
});

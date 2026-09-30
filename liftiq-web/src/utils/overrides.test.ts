import { describe, expect, it } from 'vitest';
import {
  effectiveStatus,
  needsReview,
  overridesNeedingReview,
  reconcileOverrides,
  type Overrides,
} from './overrides';
import type { RuleResult, RuleStatus } from '../types/compliance';

function result(id: string, status: RuleStatus): RuleResult {
  return {
    rule_id: id,
    description: id,
    asme_ref: 'ASME A17.1',
    subsystem: 'Brake System',
    metric: 'm',
    value: 1,
    threshold: 80,
    comparison: 'at_most',
    unit: 'ms',
    status,
    message: '',
  };
}

// ---------------------------------------------------------------------------
// effectiveStatus — the Override counts only while it is not awaiting review
// ---------------------------------------------------------------------------
describe('effectiveStatus', () => {
  it('uses the Rule result when there is no Override', () => {
    expect(effectiveStatus('pass', undefined)).toBe('pass');
    expect(effectiveStatus('fail', undefined)).toBe('fail');
    expect(effectiveStatus('unknown', undefined)).toBe('unknown');
  });

  it('uses the Override while the Rule result is the one it was made against', () => {
    expect(effectiveStatus('pass', { status: 'fail', against: 'pass' })).toBe('fail');
    expect(effectiveStatus('fail', { status: 'pass', against: 'fail' })).toBe('pass');
    expect(effectiveStatus('unknown', { status: 'pass', against: 'unknown' })).toBe('pass');
  });

  it('ignores an Override that needs re-confirmation', () => {
    // Passed while unknown; telemetry now reports a real fail.
    expect(effectiveStatus('fail', { status: 'pass', against: 'unknown' })).toBe('fail');
  });
});

// ---------------------------------------------------------------------------
// needsReview — decision 1 (status change only) and 2 (agreement is fine)
// ---------------------------------------------------------------------------
describe('needsReview', () => {
  it('is false while the Rule result is unchanged', () => {
    expect(needsReview({ status: 'fail', against: 'pass' }, 'pass')).toBe(false);
  });

  it('is true when the Rule result changes to disagree with the Override', () => {
    expect(needsReview({ status: 'pass', against: 'unknown' }, 'fail')).toBe(true);
    expect(needsReview({ status: 'fail', against: 'pass' }, 'unknown')).toBe(true);
  });

  it('is false when the new Rule result agrees with the Override', () => {
    expect(needsReview({ status: 'fail', against: 'pass' }, 'fail')).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// reconcileOverrides — re-anchors Overrides the new snapshot agrees with
// ---------------------------------------------------------------------------
describe('reconcileOverrides', () => {
  it('re-anchors an Override when the new Rule result agrees with it', () => {
    const overrides: Overrides = { 'ASME-011': { status: 'fail', against: 'pass' } };
    const next = reconcileOverrides(overrides, [result('ASME-011', 'fail')]);
    expect(next['ASME-011']).toEqual({ status: 'fail', against: 'fail' });
  });

  it('then flags it if the Rule result moves away again', () => {
    // fail made against pass → telemetry agrees (fail) → telemetry back to pass.
    const agreed = reconcileOverrides(
      { 'ASME-011': { status: 'fail', against: 'pass' } },
      [result('ASME-011', 'fail')],
    );
    expect(needsReview(agreed['ASME-011'], 'pass')).toBe(true);
  });

  it('leaves a disagreeing Override flagged, not silently re-anchored', () => {
    const overrides: Overrides = { 'ASME-011': { status: 'pass', against: 'unknown' } };
    const next = reconcileOverrides(overrides, [result('ASME-011', 'fail')]);
    expect(next['ASME-011']).toEqual({ status: 'pass', against: 'unknown' });
  });

  it('returns the same object when nothing changes, so it is safe in an effect', () => {
    const overrides: Overrides = { 'ASME-011': { status: 'fail', against: 'pass' } };
    expect(reconcileOverrides(overrides, [result('ASME-011', 'pass')])).toBe(overrides);
  });
});

// ---------------------------------------------------------------------------
// overridesNeedingReview — what blocks signing (decision 3)
// ---------------------------------------------------------------------------
describe('overridesNeedingReview', () => {
  it('lists flagged rules in result order', () => {
    const overrides: Overrides = {
      'ASME-014': { status: 'pass', against: 'unknown' },
      'ASME-006': { status: 'fail', against: 'pass' },
      'ASME-011': { status: 'pass', against: 'fail' },
    };
    const results = [result('ASME-006', 'pass'), result('ASME-011', 'unknown'), result('ASME-014', 'fail')];
    expect(overridesNeedingReview(overrides, results)).toEqual(['ASME-011', 'ASME-014']);
  });

  it('is empty when every Override still matches its evidence', () => {
    expect(
      overridesNeedingReview({ 'ASME-006': { status: 'fail', against: 'pass' } }, [result('ASME-006', 'pass')]),
    ).toEqual([]);
  });
});

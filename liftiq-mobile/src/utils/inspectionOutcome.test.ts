import { applyOverrides, buildReportRequest, summarise } from './inspectionOutcome';
import type { RuleResult, RuleStatus } from '../types/compliance';

function rule(id: string, status: RuleStatus, value: number | null = 1): RuleResult {
  return {
    rule_id: id,
    description: `Rule ${id}`,
    asme_ref: 'ASME A17.1',
    subsystem: 'Motor',
    metric: 'm',
    value,
    threshold: 10,
    comparison: 'at_most',
    unit: 'A',
    status,
    message: `${status} message`,
  };
}

const telemetry = [rule('ASME-001', 'pass'), rule('ASME-006', 'pass'), rule('ASME-014', 'unknown', null)];

describe('applyOverrides + summarise', () => {
  it('uses telemetry status when there are no overrides', () => {
    expect(summarise(applyOverrides(telemetry, {}))).toEqual({
      pass: 2, fail: 0, unknown: 1, overall: 'unknown',
    });
  });

  it('counts a technician fail and makes the inspection fail', () => {
    expect(summarise(applyOverrides(telemetry, { 'ASME-006': { status: 'fail', against: 'pass' } }))).toEqual({
      pass: 1, fail: 1, unknown: 1, overall: 'fail',
    });
  });

  it('lets a technician resolve an unknown', () => {
    expect(summarise(applyOverrides(telemetry, { 'ASME-014': { status: 'pass', against: 'unknown' } }))).toEqual({
      pass: 3, fail: 0, unknown: 0, overall: 'pass',
    });
  });

  it('ignores an Override whose Rule result changed to disagree with it', () => {
    // ASME-001 was passed while unknown; telemetry now says pass → agrees; but
    // an Override of fail made while unknown, now pass, needs review.
    expect(summarise(applyOverrides(telemetry, { 'ASME-001': { status: 'fail', against: 'unknown' } }))).toEqual({
      pass: 2, fail: 0, unknown: 1, overall: 'unknown',
    });
  });

  it('does not mutate the telemetry results', () => {
    applyOverrides(telemetry, { 'ASME-001': { status: 'fail', against: 'pass' } });
    expect(telemetry[0].status).toBe('pass');
  });
});

describe('buildReportRequest', () => {
  const req = buildReportRequest({
    unitTag: 'ELV-003',
    inspectedAt: '2026-09-29T12:00:00Z',
    technician: 'Inspector',
    results: telemetry,
    overrides: { 'ASME-006': { status: 'fail', against: 'pass' } },
    photos: [],
    signatureDataURI: 'data:image/svg+xml;base64,xx',
  });

  it('sends each Override with the result it was made against', () => {
    expect(req.overrides).toEqual({ 'ASME-006': { status: 'fail', against: 'pass' } });
  });

  it('sends telemetry status, not the overridden status', () => {
    expect(req.results.find((r) => r.rule_id === 'ASME-006')?.status).toBe('pass');
  });

  it('does not send a summary — the server derives it', () => {
    expect(req).not.toHaveProperty('summary');
  });

  it('forwards subsystem and comparison so the report can group and show limits', () => {
    expect(req.results[0]).toMatchObject({ subsystem: 'Motor', comparison: 'at_most' });
  });

  it('carries unit, technician and signature', () => {
    expect(req.unit_tag).toBe('ELV-003');
    expect(req.technician).toBe('Inspector');
    expect(req.signature_data_uri).toBe('data:image/svg+xml;base64,xx');
  });
});

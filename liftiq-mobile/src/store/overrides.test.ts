import { effectiveStatus } from './overrides';

// ---------------------------------------------------------------------------
// effectiveStatus — pure function: override takes precedence over telemetry
// ---------------------------------------------------------------------------
describe('effectiveStatus', () => {
  // No override — telemetry status is authoritative
  describe('when no override is set', () => {
    it('returns pass when telemetry is pass and override is undefined', () => {
      expect(effectiveStatus('pass', undefined)).toBe('pass');
    });
    it('returns fail when telemetry is fail and override is undefined', () => {
      expect(effectiveStatus('fail', undefined)).toBe('fail');
    });
    it('returns unknown when telemetry is unknown and override is undefined', () => {
      expect(effectiveStatus('unknown', undefined)).toBe('unknown');
    });
  });

  // Override pass — always wins regardless of telemetry
  describe('when override is pass', () => {
    it('returns pass when telemetry is also pass', () => {
      expect(effectiveStatus('pass', 'pass')).toBe('pass');
    });
    it('returns pass when telemetry is fail', () => {
      expect(effectiveStatus('fail', 'pass')).toBe('pass');
    });
    it('returns pass when telemetry is unknown', () => {
      expect(effectiveStatus('unknown', 'pass')).toBe('pass');
    });
  });

  // Override fail — always wins regardless of telemetry
  describe('when override is fail', () => {
    it('returns fail when telemetry is pass', () => {
      expect(effectiveStatus('pass', 'fail')).toBe('fail');
    });
    it('returns fail when telemetry is also fail', () => {
      expect(effectiveStatus('fail', 'fail')).toBe('fail');
    });
    it('returns fail when telemetry is unknown', () => {
      expect(effectiveStatus('unknown', 'fail')).toBe('fail');
    });
  });
});

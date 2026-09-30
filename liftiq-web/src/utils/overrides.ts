import type { RuleResult, RuleStatus } from '../types/compliance';

export type OverrideStatus = 'pass' | 'fail';

/**
 * A technician's call on a rule, anchored to the Rule result status it was
 * made (or last re-confirmed) against.
 */
export interface Override {
  status: OverrideStatus;
  against: RuleStatus;
}

/** rule_id → Override. Plain object so it survives navigation params. */
export type Overrides = Record<string, Override>;

/**
 * An Override needs re-confirmation when its rule's result status has changed
 * to something that disagrees with the Override. A value change within the
 * same status never counts, and agreement is never flagged.
 */
export function needsReview(override: Override, currentStatus: RuleStatus): boolean {
  return currentStatus !== override.against && currentStatus !== override.status;
}

/** The status the Inspection certifies: an Override counts only while it doesn't need review. */
export function effectiveStatus(
  currentStatus: RuleStatus,
  override: Override | undefined,
): RuleStatus {
  if (!override || needsReview(override, currentStatus)) return currentStatus;
  return override.status;
}

/**
 * Re-anchors every Override whose rule now agrees with it, so a later change
 * away from agreement is flagged. Returns the same object when nothing changed.
 */
export function reconcileOverrides(overrides: Overrides, results: RuleResult[]): Overrides {
  let next: Overrides | null = null;
  for (const r of results) {
    const o = overrides[r.rule_id];
    if (o && r.status === o.status && o.against !== r.status) {
      next = next ?? { ...overrides };
      next[r.rule_id] = { status: o.status, against: r.status };
    }
  }
  return next ?? overrides;
}

/** Rule IDs, in result order, whose Override needs re-confirmation before signing. */
export function overridesNeedingReview(overrides: Overrides, results: RuleResult[]): string[] {
  return results
    .filter((r) => overrides[r.rule_id] && needsReview(overrides[r.rule_id], r.status))
    .map((r) => r.rule_id);
}

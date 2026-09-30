import type { RuleResult } from '../types/compliance';

export interface Section {
  title: string;
  data: RuleResult[];
}

const OTHER = 'Other';

/**
 * Groups results by the subsystem the compliance engine assigned, in the order
 * each subsystem first appears. Results without a subsystem go under "Other",
 * last, so no rule is ever dropped from the checklist.
 */
export function groupBySubsystem(results: RuleResult[]): Section[] {
  const sections: Section[] = [];
  const byTitle = new Map<string, Section>();
  const other: RuleResult[] = [];
  for (const r of results) {
    if (!r.subsystem) {
      other.push(r);
      continue;
    }
    let section = byTitle.get(r.subsystem);
    if (!section) {
      section = { title: r.subsystem, data: [] };
      byTitle.set(r.subsystem, section);
      sections.push(section);
    }
    section.data.push(r);
  }
  if (other.length > 0) sections.push({ title: OTHER, data: other });
  return sections;
}

import type { RuleStatus } from '../types/compliance';

const COLORS: Record<RuleStatus, string> = {
  pass: 'bg-green-500',
  fail: 'bg-red-500',
  unknown: 'bg-gray-400',
};

interface Props {
  status: RuleStatus;
  overridden?: boolean;
}

export function StatusBadge({ status, overridden = false }: Props) {
  return (
    <span
      className={`inline-block px-2 py-0.5 rounded text-white text-[11px] font-bold tracking-wide ${COLORS[status]}`}
    >
      {status.toUpperCase()}
      {overridden ? ' (manual)' : ''}
    </span>
  );
}

import type { VoiceIntent } from '../types/voice';

// More specific multi-word phrases must come before single-word tokens so that
// e.g. "next item" matches 'skip' before the word "next" matches 'next'.
const INTENT_PATTERNS: { intent: VoiceIntent; phrases: string[] }[] = [
  {
    intent: 'skip',
    phrases: ['next item', 'skip', 'skipped', 'ignore'],
  },
  {
    intent: 'photo',
    phrases: ['take photo', 'take a photo', 'take picture', 'photo', 'picture', 'image', 'camera'],
  },
  {
    intent: 'pass',
    phrases: ['looks good', 'passed', 'pass', 'good', 'ok', 'okay', 'approve', 'approved'],
  },
  {
    intent: 'fail',
    phrases: ['no good', 'failed', 'fail', 'bad', 'reject', 'rejected', 'defect'],
  },
  {
    intent: 'next',
    phrases: ['move on', 'next', 'advance', 'continue'],
  },
  {
    intent: 'stop',
    phrases: ['stop listening', 'stop', 'cancel', 'quit', 'exit', 'done'],
  },
];

export function parseIntent(transcript: string): VoiceIntent {
  const normalized = transcript.toLowerCase().trim();
  for (const { intent, phrases } of INTENT_PATTERNS) {
    for (const phrase of phrases) {
      if (normalized.includes(phrase)) {
        return intent;
      }
    }
  }
  return 'unknown';
}

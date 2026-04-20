import type { VoiceIntent } from '../types/voice';

// Rules:
// 1. 'fail' is ordered before 'pass' so "no good" (contains "good") resolves to
//    fail rather than pass.
// 2. More specific multi-word phrases (e.g. "next item") come before single-word
//    tokens within each group so they are checked first.
// 3. Matching uses word-boundary regex so short tokens like "ok" do not match
//    inside longer words like "looks" or "book".
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
    intent: 'fail',
    phrases: ['no good', 'failed', 'fail', 'bad', 'reject', 'rejected', 'defect'],
  },
  {
    intent: 'pass',
    phrases: ['looks good', 'passed', 'pass', 'good', 'ok', 'okay', 'approve', 'approved'],
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

/** Build a word-boundary regex for a phrase (handles multi-word phrases too). */
function phraseRegex(phrase: string): RegExp {
  // Escape any regex-special chars, then wrap in word boundaries.
  const escaped = phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`\\b${escaped}\\b`);
}

export function parseIntent(transcript: string): VoiceIntent {
  const normalized = transcript.toLowerCase().trim();
  for (const { intent, phrases } of INTENT_PATTERNS) {
    for (const phrase of phrases) {
      if (phraseRegex(phrase).test(normalized)) {
        return intent;
      }
    }
  }
  return 'unknown';
}

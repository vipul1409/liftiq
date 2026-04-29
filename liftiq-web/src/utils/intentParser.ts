export type VoiceIntent = 'pass' | 'fail' | 'skip' | 'next' | 'photo' | 'stop' | 'unknown';

const INTENT_PATTERNS: { intent: VoiceIntent; phrases: string[] }[] = [
  { intent: 'skip', phrases: ['next item', 'skip', 'skipped', 'ignore'] },
  { intent: 'photo', phrases: ['take photo', 'take a photo', 'take picture', 'photo', 'picture', 'image', 'camera'] },
  { intent: 'fail', phrases: ['no good', 'failed', 'fail', 'bad', 'reject', 'rejected', 'defect'] },
  { intent: 'pass', phrases: ['looks good', 'passed', 'pass', 'good', 'ok', 'okay', 'approve', 'approved'] },
  { intent: 'next', phrases: ['move on', 'next', 'advance', 'continue'] },
  { intent: 'stop', phrases: ['stop listening', 'stop', 'cancel', 'quit', 'exit', 'done'] },
];

function phraseRegex(phrase: string): RegExp {
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

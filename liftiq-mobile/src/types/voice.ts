export type VoiceIntent =
  | 'pass'
  | 'fail'
  | 'skip'
  | 'next'
  | 'photo'
  | 'stop'
  | 'unknown';

export interface VoiceState {
  listening: boolean;
  isSpeaking: boolean;
  transcript: string;
  lastIntent: VoiceIntent | null;
}

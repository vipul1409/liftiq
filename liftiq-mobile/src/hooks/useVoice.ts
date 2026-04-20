import { useCallback, useEffect, useRef, useState } from 'react';
import {
  ExpoSpeechRecognitionModule,
  useSpeechRecognitionEvent,
} from 'expo-speech-recognition';
import * as Speech from 'expo-speech';
import { parseIntent } from '../utils/intentParser';
import type { VoiceIntent } from '../types/voice';

export interface UseVoiceOptions {
  /** Called every time a final transcript is parsed into an intent. */
  onIntent: (intent: VoiceIntent) => void;
  /**
   * When this string changes (and is non-empty) the hook speaks it aloud.
   * Listening is paused while the TTS is active to avoid the mic picking up
   * the app's own voice.
   */
  readText?: string;
  /** Set false to fully disable the voice feature (e.g. while loading). */
  enabled?: boolean;
}

export interface UseVoiceResult {
  listening: boolean;
  isSpeaking: boolean;
  transcript: string;
  lastIntent: VoiceIntent | null;
  hasPermission: boolean;
  startListening: () => void;
  stopListening: () => void;
}

export function useVoice({
  onIntent,
  readText,
  enabled = true,
}: UseVoiceOptions): UseVoiceResult {
  const [listening, setListening] = useState(false);
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [lastIntent, setLastIntent] = useState<VoiceIntent | null>(null);
  const [hasPermission, setHasPermission] = useState(false);

  // Keep a stable ref to onIntent so the event handler below never goes stale.
  const onIntentRef = useRef(onIntent);
  useEffect(() => {
    onIntentRef.current = onIntent;
  }, [onIntent]);

  // Request permissions on mount.
  useEffect(() => {
    ExpoSpeechRecognitionModule.requestPermissionsAsync().then(({ granted }) => {
      setHasPermission(granted);
    });
    return () => {
      Speech.stop();
      ExpoSpeechRecognitionModule.stop();
    };
  }, []);

  // Handle final recognition result.
  useSpeechRecognitionEvent('result', (event) => {
    const text = event.results[0]?.transcript ?? '';
    if (!event.isFinal || !text) return;
    const intent = parseIntent(text);
    setTranscript(text);
    setLastIntent(intent);
    onIntentRef.current(intent);
  });

  useSpeechRecognitionEvent('end', () => {
    setListening(false);
  });

  useSpeechRecognitionEvent('error', () => {
    setListening(false);
  });

  const stopListening = useCallback(() => {
    ExpoSpeechRecognitionModule.stop();
    setListening(false);
  }, []);

  const startListening = useCallback(() => {
    if (!hasPermission || !enabled || isSpeaking) return;
    Speech.stop(); // stop any active TTS first
    ExpoSpeechRecognitionModule.start({
      lang: 'en-US',
      interimResults: false,
      continuous: false,
    });
    setListening(true);
  }, [hasPermission, enabled, isSpeaking]);

  // Speak readText whenever it changes.
  const prevReadText = useRef('');
  useEffect(() => {
    if (!readText || readText === prevReadText.current) return;
    prevReadText.current = readText;

    // Stop listening before speaking to avoid feedback loop.
    if (listening) {
      ExpoSpeechRecognitionModule.stop();
      setListening(false);
    }

    Speech.speak(readText, {
      language: 'en-US',
      onStart: () => setIsSpeaking(true),
      onDone: () => setIsSpeaking(false),
      onStopped: () => setIsSpeaking(false),
      onError: () => setIsSpeaking(false),
    });
  }, [readText]); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    listening,
    isSpeaking,
    transcript,
    lastIntent,
    hasPermission,
    startListening,
    stopListening,
  };
}

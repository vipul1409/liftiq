import { useState, useEffect, useRef, useCallback } from 'react';
import { parseIntent, type VoiceIntent } from '../utils/intentParser';

interface UseVoiceOptions {
  onIntent: (intent: VoiceIntent) => void;
  readText: string;
  enabled: boolean;
}

interface UseVoiceResult {
  listening: boolean;
  isSpeaking: boolean;
  transcript: string;
  lastIntent: VoiceIntent | null;
  hasPermission: boolean;
  startListening: () => void;
  stopListening: () => void;
}

const SpeechRecognitionCtor =
  (typeof window !== 'undefined' &&
    ((window as unknown as Record<string, unknown>).SpeechRecognition ||
     (window as unknown as Record<string, unknown>).webkitSpeechRecognition)) as
  (new () => SpeechRecognition) | undefined;

export function useVoice({ onIntent, readText, enabled }: UseVoiceOptions): UseVoiceResult {
  const [listening, setListening] = useState(false);
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [lastIntent, setLastIntent] = useState<VoiceIntent | null>(null);
  const recognitionRef = useRef<SpeechRecognition | null>(null);
  const onIntentRef = useRef(onIntent);
  onIntentRef.current = onIntent;

  const hasPermission = !!SpeechRecognitionCtor;

  const stopListening = useCallback(() => {
    recognitionRef.current?.stop();
    setListening(false);
  }, []);

  const startListening = useCallback(() => {
    if (!SpeechRecognitionCtor || !enabled) return;

    const recognition = new SpeechRecognitionCtor();
    recognition.continuous = false;
    recognition.interimResults = false;
    recognition.lang = 'en-US';

    recognition.onresult = (event: SpeechRecognitionEvent) => {
      const text = event.results[0]?.[0]?.transcript ?? '';
      setTranscript(text);
      const intent = parseIntent(text);
      setLastIntent(intent);
      onIntentRef.current(intent);
    };

    recognition.onerror = () => {
      setListening(false);
    };

    recognition.onend = () => {
      setListening(false);
    };

    recognitionRef.current = recognition;
    recognition.start();
    setListening(true);
    setTranscript('');
    setLastIntent(null);
  }, [enabled]);

  // TTS readback — pause mic while speaking to prevent feedback.
  useEffect(() => {
    if (!readText || !window.speechSynthesis) return;
    const utterance = new SpeechSynthesisUtterance(readText);
    utterance.rate = 1.1;

    const wasListening = listening;
    if (wasListening) stopListening();

    setIsSpeaking(true);
    utterance.onend = () => {
      setIsSpeaking(false);
    };
    utterance.onerror = () => {
      setIsSpeaking(false);
    };
    window.speechSynthesis.speak(utterance);
  }, [readText]); // eslint-disable-line react-hooks/exhaustive-deps

  return { listening, isSpeaking, transcript, lastIntent, hasPermission, startListening, stopListening };
}

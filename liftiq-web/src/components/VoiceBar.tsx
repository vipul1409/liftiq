import type { VoiceIntent } from '../utils/intentParser';

interface Props {
  listening: boolean;
  isSpeaking: boolean;
  transcript: string;
  lastIntent: VoiceIntent | null;
  hasPermission: boolean;
  onStartListening: () => void;
  onStopListening: () => void;
}

export function VoiceBar({
  listening,
  isSpeaking,
  transcript,
  lastIntent,
  hasPermission,
  onStartListening,
  onStopListening,
}: Props) {
  const micDisabled = isSpeaking || !hasPermission;

  function getMicColor() {
    if (!hasPermission) return 'bg-gray-400';
    if (isSpeaking) return 'bg-gray-400';
    if (listening) return 'bg-red-600';
    return 'bg-blue-600';
  }

  function getCenterText() {
    if (!hasPermission) return 'Voice not supported in this browser';
    if (isSpeaking) return 'Speaking...';
    if (listening) return 'Listening...';
    if (lastIntent && lastIntent !== 'unknown') return `Heard: ${lastIntent}`;
    if (transcript && lastIntent === 'unknown') return "Didn't catch that";
    return 'Tap mic for voice commands';
  }

  return (
    <div className="fixed bottom-[76px] left-4 right-4 max-w-3xl mx-auto flex items-center gap-2.5 bg-white/[.97] border border-gray-200 rounded-xl px-3 py-2.5 shadow-lg z-10">
      {/* Mic button */}
      <button
        onClick={listening ? onStopListening : onStartListening}
        disabled={micDisabled}
        className={`relative w-[38px] h-[38px] rounded-full flex items-center justify-center shrink-0 ${getMicColor()} text-white disabled:cursor-not-allowed`}
      >
        {listening && (
          <span className="absolute inset-0 rounded-full border-2 border-red-600/30 animate-pulse-ring" />
        )}
        <span className="text-base">{listening ? '\u25A0' : '\uD83C\uDFA4'}</span>
      </button>

      {/* Status text */}
      <span className="flex-1 text-[13px] text-gray-700 font-medium truncate">
        {getCenterText()}
      </span>

      {/* Speaking dots */}
      {isSpeaking && (
        <div className="flex items-center gap-[3px] shrink-0">
          <span className="w-[5px] h-[5px] rounded-full bg-blue-600 animate-bounce-dot-1" />
          <span className="w-[5px] h-[5px] rounded-full bg-blue-600 animate-bounce-dot-2" />
          <span className="w-[5px] h-[5px] rounded-full bg-blue-600 animate-bounce-dot-3" />
        </div>
      )}
    </div>
  );
}

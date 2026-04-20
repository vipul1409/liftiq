/**
 * useVoice hook tests.
 *
 * Both expo-speech and expo-speech-recognition ship native modules that cannot
 * run in a Node/Jest environment. We mock them at the module boundary so the
 * hook logic (permission handling, intent dispatch, TTS readback gating) can
 * be exercised without a device.
 */

import { renderHook, act, waitFor } from '@testing-library/react-native';
import type { VoiceIntent } from '../types/voice';

// ---------------------------------------------------------------------------
// Stable mock handles — captured before jest.mock() hoisting runs.
// These are updated inside each test via mockImplementation / mockResolvedValue.
// ---------------------------------------------------------------------------
const mockStart = jest.fn();
const mockStop = jest.fn();
const mockRequestPermissions = jest.fn();
const mockSpeak = jest.fn();
const mockSpeechStop = jest.fn();

// Event listener registry — keyed by event name.
const listeners: Record<string, ((...args: unknown[]) => void)[]> = {};

const mockAddListener = jest.fn((event: string, cb: (...args: unknown[]) => void) => {
  if (!listeners[event]) listeners[event] = [];
  listeners[event].push(cb);
  return { remove: () => {} };
});

/** Fire a registered listener by event name. */
function emit(event: string, payload?: unknown) {
  (listeners[event] ?? []).forEach((cb) => cb(payload));
}

// ---------------------------------------------------------------------------
// Module mocks
// ---------------------------------------------------------------------------
jest.mock('expo-speech-recognition', () => ({
  ExpoSpeechRecognitionModule: {
    start: (...args: unknown[]) => mockStart(...args),
    stop: () => mockStop(),
    requestPermissionsAsync: () => mockRequestPermissions(),
    addListener: (event: string, cb: (...args: unknown[]) => void) =>
      mockAddListener(event, cb),
  },
  useSpeechRecognitionEvent: (
    event: string,
    cb: (...args: unknown[]) => void,
  ) => {
    // Register via the shared listener map so tests can fire events.
    if (!listeners[event]) listeners[event] = [];
    // Avoid duplicate registrations on re-renders.
    if (!listeners[event].includes(cb)) {
      listeners[event].push(cb);
    }
  },
}));

jest.mock('expo-speech', () => ({
  speak: (...args: unknown[]) => mockSpeak(...args),
  stop: () => mockSpeechStop(),
}));

// ---------------------------------------------------------------------------
// Import AFTER mocks are set up.
// ---------------------------------------------------------------------------
import { useVoice } from './useVoice';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function grantedPermissions() {
  mockRequestPermissions.mockResolvedValue({ granted: true });
}
function deniedPermissions() {
  mockRequestPermissions.mockResolvedValue({ granted: false });
}

beforeEach(() => {
  jest.clearAllMocks();
  // Reset listener registry between tests.
  Object.keys(listeners).forEach((k) => delete listeners[k]);
  grantedPermissions();
});

// ---------------------------------------------------------------------------
// Permission handling
// ---------------------------------------------------------------------------
describe('useVoice – permissions', () => {
  it('requests permissions on mount', async () => {
    renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(mockRequestPermissions).toHaveBeenCalledTimes(1));
  });

  it('sets hasPermission true when permission is granted', async () => {
    grantedPermissions();
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
  });

  it('sets hasPermission false when permission is denied', async () => {
    deniedPermissions();
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(false));
  });
});

// ---------------------------------------------------------------------------
// Initial state
// ---------------------------------------------------------------------------
describe('useVoice – initial state', () => {
  it('starts with listening=false', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
    expect(result.current.listening).toBe(false);
  });

  it('starts with isSpeaking=false', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
    expect(result.current.isSpeaking).toBe(false);
  });

  it('starts with empty transcript', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
    expect(result.current.transcript).toBe('');
  });

  it('starts with lastIntent=null', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
    expect(result.current.lastIntent).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// startListening / stopListening
// ---------------------------------------------------------------------------
describe('useVoice – startListening', () => {
  it('calls ExpoSpeechRecognitionModule.start when permission is granted', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());

    expect(mockStart).toHaveBeenCalledWith(
      expect.objectContaining({ lang: 'en-US', interimResults: false, continuous: false }),
    );
  });

  it('sets listening=true after startListening', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    expect(result.current.listening).toBe(true);
  });

  it('does not start if permission is denied', async () => {
    deniedPermissions();
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(false));

    act(() => result.current.startListening());
    expect(mockStart).not.toHaveBeenCalled();
  });

  it('does not start if enabled=false', async () => {
    const { result } = renderHook(() =>
      useVoice({ onIntent: jest.fn(), enabled: false }),
    );
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    expect(mockStart).not.toHaveBeenCalled();
  });

  it('stops any active TTS before starting recognition', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    expect(mockSpeechStop).toHaveBeenCalled();
  });
});

describe('useVoice – stopListening', () => {
  it('calls ExpoSpeechRecognitionModule.stop', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    act(() => result.current.stopListening());
    expect(mockStop).toHaveBeenCalled();
  });

  it('sets listening=false after stopListening', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    act(() => result.current.stopListening());
    expect(result.current.listening).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// Recognition result events → intent dispatch
// ---------------------------------------------------------------------------
describe('useVoice – result events', () => {
  it('calls onIntent with parsed intent when a final result arrives', async () => {
    const onIntent = jest.fn();
    const { result } = renderHook(() => useVoice({ onIntent }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: true, results: [{ transcript: 'pass' }] });
    });

    expect(onIntent).toHaveBeenCalledWith('pass');
  });

  it('updates transcript state with the recognised text', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: true, results: [{ transcript: 'looks good' }] });
    });

    expect(result.current.transcript).toBe('looks good');
  });

  it('updates lastIntent state', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: true, results: [{ transcript: 'fail' }] });
    });

    expect(result.current.lastIntent).toBe('fail');
  });

  it('does not call onIntent for non-final results', async () => {
    const onIntent = jest.fn();
    const { result } = renderHook(() => useVoice({ onIntent }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: false, results: [{ transcript: 'pa' }] });
    });

    expect(onIntent).not.toHaveBeenCalled();
  });

  it('does not call onIntent for empty transcript', async () => {
    const onIntent = jest.fn();
    const { result } = renderHook(() => useVoice({ onIntent }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: true, results: [{ transcript: '' }] });
    });

    expect(onIntent).not.toHaveBeenCalled();
  });

  it('dispatches unknown intent for unrecognised speech', async () => {
    const onIntent = jest.fn();
    const { result } = renderHook(() => useVoice({ onIntent }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => {
      emit('result', { isFinal: true, results: [{ transcript: 'xyzzy nonsense' }] });
    });

    expect(onIntent).toHaveBeenCalledWith('unknown');
  });

  it('correctly dispatches each supported intent', async () => {
    const cases: [string, VoiceIntent][] = [
      ['pass', 'pass'],
      ['fail', 'fail'],
      ['next', 'next'],
      ['skip', 'skip'],
      ['camera', 'photo'],
      ['stop', 'stop'],
    ];

    for (const [transcript, expected] of cases) {
      const onIntent = jest.fn();
      const { result } = renderHook(() => useVoice({ onIntent }));
      await waitFor(() => expect(result.current.hasPermission).toBe(true));

      act(() => {
        emit('result', { isFinal: true, results: [{ transcript }] });
      });

      expect(onIntent).toHaveBeenCalledWith(expected);
    }
  });
});

// ---------------------------------------------------------------------------
// Recognition lifecycle events (end / error)
// ---------------------------------------------------------------------------
describe('useVoice – lifecycle events', () => {
  it('sets listening=false when end event fires', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    expect(result.current.listening).toBe(true);

    act(() => emit('end'));
    expect(result.current.listening).toBe(false);
  });

  it('sets listening=false when error event fires', async () => {
    const { result } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => result.current.startListening());
    act(() => emit('error'));
    expect(result.current.listening).toBe(false);
  });

  it('does not call onIntent when error event fires', async () => {
    const onIntent = jest.fn();
    const { result } = renderHook(() => useVoice({ onIntent }));
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    act(() => emit('error'));
    expect(onIntent).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// TTS readback via readText prop
// ---------------------------------------------------------------------------
describe('useVoice – TTS readback', () => {
  it('calls Speech.speak when readText changes to a non-empty value', async () => {
    const { result, rerender } = renderHook(
      ({ readText }) => useVoice({ onIntent: jest.fn(), readText }),
      { initialProps: { readText: '' } },
    );
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    rerender({ readText: 'ASME-001: motor rule. Status: pass.' });
    await waitFor(() => expect(mockSpeak).toHaveBeenCalledTimes(1));

    expect(mockSpeak).toHaveBeenCalledWith(
      'ASME-001: motor rule. Status: pass.',
      expect.objectContaining({ language: 'en-US' }),
    );
  });

  it('does not call Speech.speak again with the same readText', async () => {
    const text = 'same text';
    const { rerender } = renderHook(
      ({ readText }) => useVoice({ onIntent: jest.fn(), readText }),
      { initialProps: { readText: text } },
    );
    await waitFor(() => expect(mockRequestPermissions).toHaveBeenCalled());

    // Re-render with the same text — should not call speak again.
    rerender({ readText: text });
    expect(mockSpeak).toHaveBeenCalledTimes(1);
  });

  it('does not call Speech.speak when readText is empty', async () => {
    renderHook(() => useVoice({ onIntent: jest.fn(), readText: '' }));
    await waitFor(() => expect(mockRequestPermissions).toHaveBeenCalled());
    expect(mockSpeak).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// Cleanup on unmount
// ---------------------------------------------------------------------------
describe('useVoice – cleanup on unmount', () => {
  it('calls Speech.stop and ExpoSpeechRecognitionModule.stop on unmount', async () => {
    const { unmount } = renderHook(() => useVoice({ onIntent: jest.fn() }));
    await waitFor(() => expect(mockRequestPermissions).toHaveBeenCalled());

    unmount();
    expect(mockSpeechStop).toHaveBeenCalled();
    expect(mockStop).toHaveBeenCalled();
  });
});

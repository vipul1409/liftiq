import React, { useEffect, useRef } from 'react';
import {
  Animated,
  Easing,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { VoiceIntent } from '../types/voice';

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
  // Pulsing ring animation when listening.
  const pulseAnim = useRef(new Animated.Value(1)).current;
  const pulseLoop = useRef<Animated.CompositeAnimation | null>(null);

  useEffect(() => {
    if (listening) {
      pulseLoop.current = Animated.loop(
        Animated.sequence([
          Animated.timing(pulseAnim, {
            toValue: 1.4,
            duration: 600,
            easing: Easing.inOut(Easing.ease),
            useNativeDriver: true,
          }),
          Animated.timing(pulseAnim, {
            toValue: 1,
            duration: 600,
            easing: Easing.inOut(Easing.ease),
            useNativeDriver: true,
          }),
        ]),
      );
      pulseLoop.current.start();
    } else {
      pulseLoop.current?.stop();
      pulseAnim.setValue(1);
    }
  }, [listening, pulseAnim]);

  // Bouncing dots animation when speaking.
  const dot1 = useRef(new Animated.Value(0)).current;
  const dot2 = useRef(new Animated.Value(0)).current;
  const dot3 = useRef(new Animated.Value(0)).current;
  const dotsLoop = useRef<Animated.CompositeAnimation | null>(null);

  useEffect(() => {
    const bounce = (anim: Animated.Value, delay: number) =>
      Animated.loop(
        Animated.sequence([
          Animated.delay(delay),
          Animated.timing(anim, { toValue: -4, duration: 250, useNativeDriver: true }),
          Animated.timing(anim, { toValue: 0, duration: 250, useNativeDriver: true }),
          Animated.delay(500 - delay),
        ]),
      );

    if (isSpeaking) {
      dotsLoop.current = Animated.parallel([
        bounce(dot1, 0),
        bounce(dot2, 150),
        bounce(dot3, 300),
      ]);
      dotsLoop.current.start();
    } else {
      dotsLoop.current?.stop();
      [dot1, dot2, dot3].forEach((d) => d.setValue(0));
    }
  }, [isSpeaking, dot1, dot2, dot3]);

  const micDisabled = isSpeaking || !hasPermission;

  function getMicColor() {
    if (!hasPermission) return '#9ca3af';
    if (isSpeaking) return '#9ca3af';
    if (listening) return '#dc2626';
    return '#2563eb';
  }

  function getCenterText() {
    if (!hasPermission) return 'Microphone access denied';
    if (isSpeaking) return 'Speaking…';
    if (listening) return 'Listening…';
    if (lastIntent && lastIntent !== 'unknown') return `Heard: ${lastIntent}`;
    if (transcript && lastIntent === 'unknown') return `Didn't catch that`;
    return 'Tap mic for voice commands';
  }

  return (
    <View style={styles.container}>
      {/* Mic button */}
      <TouchableOpacity
        style={[styles.micBtn, { backgroundColor: getMicColor() }]}
        onPress={listening ? onStopListening : onStartListening}
        disabled={micDisabled}
        activeOpacity={0.7}
      >
        <Animated.View
          style={[
            styles.micRing,
            listening && { transform: [{ scale: pulseAnim }] },
          ]}
        />
        <Text style={styles.micIcon}>{listening ? '■' : '🎤'}</Text>
      </TouchableOpacity>

      {/* Status text */}
      <Text style={styles.statusText} numberOfLines={1}>
        {getCenterText()}
      </Text>

      {/* Speaking dots */}
      {isSpeaking && (
        <View style={styles.dots}>
          {[dot1, dot2, dot3].map((anim, i) => (
            <Animated.View
              key={i}
              style={[styles.dot, { transform: [{ translateY: anim }] }]}
            />
          ))}
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 76,        // sits above the ~68px footer
    left: 16,
    right: 16,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.97)',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    paddingVertical: 10,
    paddingHorizontal: 12,
    gap: 10,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 6,
    elevation: 4,
    zIndex: 10,
  },
  micBtn: {
    width: 38,
    height: 38,
    borderRadius: 19,
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
  },
  micRing: {
    position: 'absolute',
    width: 38,
    height: 38,
    borderRadius: 19,
    borderWidth: 2,
    borderColor: '#dc262650',
  },
  micIcon: {
    fontSize: 16,
  },
  statusText: {
    flex: 1,
    fontSize: 13,
    color: '#374151',
    fontWeight: '500',
  },
  dots: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
    flexShrink: 0,
  },
  dot: {
    width: 5,
    height: 5,
    borderRadius: 2.5,
    backgroundColor: '#2563eb',
  },
});

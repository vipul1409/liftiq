import React, { useRef, useState, useCallback } from 'react';
import { PanResponder, StyleSheet, View } from 'react-native';
import type { LayoutRectangle, GestureResponderEvent, PanResponderGestureState } from 'react-native';

interface Point { x: number; y: number }
type Stroke = Point[];

interface SignatureCaptureProps {
  onStrokesChange: (hasStrokes: boolean) => void;
  onDataURIChange: (dataURI: string) => void;
  clearTrigger: number;   // increment to clear
  width?: number;
  height?: number;
}

/** Converts recorded strokes to a data:image/svg+xml;base64,... URI. */
export function buildSVGDataURIForTesting(strokes: Stroke[], width: number, height: number): string {
  return buildSVGDataURI(strokes, width, height);
}

function buildSVGDataURI(strokes: Stroke[], width: number, height: number): string {
  const polylines = strokes
    .filter((s) => s.length > 0)
    .map((pts) => {
      const points = pts.map((p) => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' ');
      return `<polyline points="${points}" stroke="#1e3a8a" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"/>`;
    })
    .join('');
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">${polylines}</svg>`;
  const b64 = btoa(svg);
  return `data:image/svg+xml;base64,${b64}`;
}

export function SignatureCapture({
  onStrokesChange,
  onDataURIChange,
  clearTrigger,
  width = 320,
  height = 180,
}: SignatureCaptureProps) {
  const [strokes, setStrokes] = useState<Stroke[]>([]);
  const layout = useRef<LayoutRectangle>({ x: 0, y: 0, width, height });
  const prevClearTrigger = useRef(clearTrigger);

  // Handle external clear trigger
  if (clearTrigger !== prevClearTrigger.current) {
    prevClearTrigger.current = clearTrigger;
    setStrokes([]);
    onStrokesChange(false);
  }

  const toLocal = useCallback((evt: GestureResponderEvent): Point => {
    const { pageX, pageY } = evt.nativeEvent;
    return {
      x: Math.max(0, Math.min(width, pageX - layout.current.x)),
      y: Math.max(0, Math.min(height, pageY - layout.current.y)),
    };
  }, [width, height]);

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: () => true,
      onPanResponderGrant: (evt) => {
        const pt = toLocal(evt);
        setStrokes((prev) => [...prev, [pt]]);
      },
      onPanResponderMove: (evt, _: PanResponderGestureState) => {
        const pt = toLocal(evt);
        setStrokes((prev) => {
          if (prev.length === 0) return prev;
          const next = [...prev];
          next[next.length - 1] = [...next[next.length - 1], pt];
          return next;
        });
      },
      onPanResponderRelease: (evt) => {
        setStrokes((prev) => {
          const hasStrokes = prev.some((s) => s.length > 0);
          onStrokesChange(hasStrokes);
          if (hasStrokes) {
            onDataURIChange(buildSVGDataURI(prev, width, height));
          }
          return prev;
        });
      },
    }),
  ).current;

  // Render each stroke as dots (small absolute Views)
  const dots: React.ReactElement[] = [];
  strokes.forEach((stroke, si) => {
    stroke.forEach((pt, pi) => {
      dots.push(
        <View
          key={`${si}-${pi}`}
          style={[styles.dot, { left: pt.x - 1.5, top: pt.y - 1.5 }]}
        />,
      );
    });
  });

  return (
    <View
      style={[styles.canvas, { width, height }]}
      onLayout={(e) => {
        layout.current = e.nativeEvent.layout;
      }}
      {...panResponder.panHandlers}
    >
      {dots}
    </View>
  );
}

const styles = StyleSheet.create({
  canvas: {
    backgroundColor: '#fff',
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: '#d1d5db',
    overflow: 'hidden',
    position: 'relative',
  },
  dot: {
    position: 'absolute',
    width: 3,
    height: 3,
    borderRadius: 1.5,
    backgroundColor: '#1e3a8a',
  },
});

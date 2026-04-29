import { useRef, useEffect, useCallback, useState } from 'react';

interface Point { x: number; y: number }
type Stroke = Point[];

interface Props {
  width?: number;
  height?: number;
  onStrokesChange: (hasStrokes: boolean) => void;
  onDataURIChange: (dataURI: string) => void;
  clearTrigger: number;
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
  return `data:image/svg+xml;base64,${btoa(svg)}`;
}

export function SignatureCanvas({
  width = 400,
  height = 200,
  onStrokesChange,
  onDataURIChange,
  clearTrigger,
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const strokesRef = useRef<Stroke[]>([]);
  const drawingRef = useRef(false);
  const [, forceRender] = useState(0);

  const redraw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, width, height);
    ctx.strokeStyle = '#1e3a8a';
    ctx.lineWidth = 2.5;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    for (const stroke of strokesRef.current) {
      if (stroke.length === 0) continue;
      ctx.beginPath();
      ctx.moveTo(stroke[0].x, stroke[0].y);
      for (let i = 1; i < stroke.length; i++) {
        ctx.lineTo(stroke[i].x, stroke[i].y);
      }
      ctx.stroke();
    }
  }, [width, height]);

  // Clear when trigger changes
  useEffect(() => {
    strokesRef.current = [];
    onStrokesChange(false);
    redraw();
  }, [clearTrigger]); // eslint-disable-line react-hooks/exhaustive-deps

  function getPoint(e: React.MouseEvent | React.TouchEvent): Point {
    const canvas = canvasRef.current!;
    const rect = canvas.getBoundingClientRect();
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
    const clientY = 'touches' in e ? e.touches[0].clientY : e.clientY;
    return {
      x: Math.max(0, Math.min(width, clientX - rect.left)),
      y: Math.max(0, Math.min(height, clientY - rect.top)),
    };
  }

  function handleStart(e: React.MouseEvent | React.TouchEvent) {
    e.preventDefault();
    drawingRef.current = true;
    const pt = getPoint(e);
    strokesRef.current.push([pt]);
    redraw();
  }

  function handleMove(e: React.MouseEvent | React.TouchEvent) {
    if (!drawingRef.current) return;
    e.preventDefault();
    const pt = getPoint(e);
    const current = strokesRef.current[strokesRef.current.length - 1];
    current.push(pt);
    redraw();
  }

  function handleEnd() {
    drawingRef.current = false;
    const hasStrokes = strokesRef.current.some((s) => s.length > 0);
    onStrokesChange(hasStrokes);
    if (hasStrokes) {
      onDataURIChange(buildSVGDataURI(strokesRef.current, width, height));
    }
    forceRender((n) => n + 1);
  }

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      className="bg-white rounded-lg border-[1.5px] border-gray-300 cursor-crosshair touch-none"
      onMouseDown={handleStart}
      onMouseMove={handleMove}
      onMouseUp={handleEnd}
      onMouseLeave={handleEnd}
      onTouchStart={handleStart}
      onTouchMove={handleMove}
      onTouchEnd={handleEnd}
    />
  );
}

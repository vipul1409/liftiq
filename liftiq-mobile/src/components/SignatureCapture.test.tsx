/**
 * SignatureCapture component tests.
 *
 * PanResponder cannot be triggered via RNTL fire events easily, so we test:
 *  - rendering (canvas area appears)
 *  - SVG data URI builder (pure function, exported for testing)
 *  - clear trigger resets strokes
 */

import React, { useState } from 'react';
import { render, act } from '@testing-library/react-native';
import { SignatureCapture, buildSVGDataURIForTesting } from './SignatureCapture';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function Wrapper({ clearTrigger = 0 }: { clearTrigger?: number }) {
  const [hasStrokes, setHasStrokes] = useState(false);
  const [dataURI, setDataURI] = useState('');
  return (
    <>
      <SignatureCapture
        onStrokesChange={setHasStrokes}
        onDataURIChange={setDataURI}
        clearTrigger={clearTrigger}
        width={300}
        height={150}
      />
    </>
  );
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

describe('SignatureCapture – rendering', () => {
  it('renders without crashing', () => {
    const { toJSON } = render(
      <SignatureCapture
        onStrokesChange={jest.fn()}
        onDataURIChange={jest.fn()}
        clearTrigger={0}
        width={300}
        height={150}
      />,
    );
    expect(toJSON()).not.toBeNull();
  });

  it('renders with default width/height when not provided', () => {
    const { toJSON } = render(
      <SignatureCapture
        onStrokesChange={jest.fn()}
        onDataURIChange={jest.fn()}
        clearTrigger={0}
      />,
    );
    expect(toJSON()).not.toBeNull();
  });
});

// ---------------------------------------------------------------------------
// SVG builder (pure function)
// ---------------------------------------------------------------------------

describe('buildSVGDataURIForTesting – SVG generation', () => {
  it('returns a data:image/svg+xml;base64 URI', () => {
    const uri = buildSVGDataURIForTesting(
      [[{ x: 0, y: 0 }, { x: 10, y: 10 }]],
      100,
      100,
    );
    expect(uri).toMatch(/^data:image\/svg\+xml;base64,/);
  });

  it('decodes to valid SVG with a polyline', () => {
    const uri = buildSVGDataURIForTesting(
      [[{ x: 5, y: 5 }, { x: 50, y: 50 }]],
      100,
      100,
    );
    const b64 = uri.replace('data:image/svg+xml;base64,', '');
    const svg = atob(b64);
    expect(svg).toContain('<svg');
    expect(svg).toContain('<polyline');
    expect(svg).toContain('5.0,5.0');
  });

  it('encodes width and height in the SVG viewBox', () => {
    const uri = buildSVGDataURIForTesting([[{ x: 1, y: 1 }]], 320, 200);
    const svg = atob(uri.replace('data:image/svg+xml;base64,', ''));
    expect(svg).toContain('width="320"');
    expect(svg).toContain('height="200"');
  });

  it('encodes multiple strokes as separate polylines', () => {
    const uri = buildSVGDataURIForTesting(
      [
        [{ x: 0, y: 0 }, { x: 10, y: 10 }],
        [{ x: 20, y: 20 }, { x: 30, y: 30 }],
      ],
      100,
      100,
    );
    const svg = atob(uri.replace('data:image/svg+xml;base64,', ''));
    const count = (svg.match(/<polyline/g) || []).length;
    expect(count).toBe(2);
  });

  it('skips empty strokes', () => {
    const uri = buildSVGDataURIForTesting(
      [[], [{ x: 1, y: 1 }, { x: 2, y: 2 }]],
      100,
      100,
    );
    const svg = atob(uri.replace('data:image/svg+xml;base64,', ''));
    const count = (svg.match(/<polyline/g) || []).length;
    expect(count).toBe(1);
  });

  it('uses LiftIQ brand colour (#1e3a8a) for stroke', () => {
    const uri = buildSVGDataURIForTesting([[{ x: 0, y: 0 }]], 100, 100);
    const svg = atob(uri.replace('data:image/svg+xml;base64,', ''));
    expect(svg).toContain('#1e3a8a');
  });
});

// ---------------------------------------------------------------------------
// Clear trigger
// ---------------------------------------------------------------------------

describe('SignatureCapture – clear trigger', () => {
  it('calls onStrokesChange(false) when clearTrigger increments', () => {
    const onStrokesChange = jest.fn();
    const { rerender } = render(
      <SignatureCapture
        onStrokesChange={onStrokesChange}
        onDataURIChange={jest.fn()}
        clearTrigger={0}
        width={300}
        height={150}
      />,
    );

    act(() => {
      rerender(
        <SignatureCapture
          onStrokesChange={onStrokesChange}
          onDataURIChange={jest.fn()}
          clearTrigger={1}
          width={300}
          height={150}
        />,
      );
    });

    expect(onStrokesChange).toHaveBeenCalledWith(false);
  });
});

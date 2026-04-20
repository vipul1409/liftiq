import { renderHook, act } from '@testing-library/react-native';
import { usePhotos } from './photos';
import type { CapturedPhoto } from './photos';

function makePhoto(ruleId: string, uri: string, overrides: Partial<CapturedPhoto> = {}): CapturedPhoto {
  return {
    uri,
    ruleId,
    timestamp: '2026-04-20T10:00:00.000Z',
    latitude: 37.7749,
    longitude: -122.4194,
    ...overrides,
  };
}

describe('usePhotos', () => {
  // ---------------------------------------------------------------------------
  // Initial state
  // ---------------------------------------------------------------------------
  describe('initial state', () => {
    it('starts with an empty photo map', () => {
      const { result } = renderHook(() => usePhotos());
      expect(result.current.photos.size).toBe(0);
    });

    it('getPhotos returns empty array for unknown ruleId', () => {
      const { result } = renderHook(() => usePhotos());
      expect(result.current.getPhotos('ASME-001')).toEqual([]);
    });
  });

  // ---------------------------------------------------------------------------
  // addPhoto
  // ---------------------------------------------------------------------------
  describe('addPhoto', () => {
    it('adds a photo and makes it retrievable by ruleId', () => {
      const { result } = renderHook(() => usePhotos());
      const photo = makePhoto('ASME-006', 'file:///img1.jpg');

      act(() => result.current.addPhoto(photo));

      expect(result.current.getPhotos('ASME-006')).toHaveLength(1);
      expect(result.current.getPhotos('ASME-006')[0]).toEqual(photo);
    });

    it('appends multiple photos under the same ruleId', () => {
      const { result } = renderHook(() => usePhotos());
      const p1 = makePhoto('ASME-006', 'file:///img1.jpg');
      const p2 = makePhoto('ASME-006', 'file:///img2.jpg');

      act(() => {
        result.current.addPhoto(p1);
        result.current.addPhoto(p2);
      });

      expect(result.current.getPhotos('ASME-006')).toHaveLength(2);
    });

    it('keeps photos for different ruleIds isolated', () => {
      const { result } = renderHook(() => usePhotos());
      act(() => {
        result.current.addPhoto(makePhoto('ASME-006', 'file:///a.jpg'));
        result.current.addPhoto(makePhoto('ASME-011', 'file:///b.jpg'));
      });

      expect(result.current.getPhotos('ASME-006')).toHaveLength(1);
      expect(result.current.getPhotos('ASME-011')).toHaveLength(1);
    });

    it('stores the full photo metadata (uri, timestamp, coords)', () => {
      const { result } = renderHook(() => usePhotos());
      const photo = makePhoto('ASME-001', 'file:///x.jpg', {
        timestamp: '2026-04-20T12:30:00.000Z',
        latitude: 51.5074,
        longitude: -0.1278,
      });

      act(() => result.current.addPhoto(photo));

      const stored = result.current.getPhotos('ASME-001')[0];
      expect(stored.timestamp).toBe('2026-04-20T12:30:00.000Z');
      expect(stored.latitude).toBe(51.5074);
      expect(stored.longitude).toBe(-0.1278);
    });

    it('stores a photo with null coordinates (location denied)', () => {
      const { result } = renderHook(() => usePhotos());
      const photo = makePhoto('ASME-001', 'file:///x.jpg', {
        latitude: null,
        longitude: null,
      });

      act(() => result.current.addPhoto(photo));

      const stored = result.current.getPhotos('ASME-001')[0];
      expect(stored.latitude).toBeNull();
      expect(stored.longitude).toBeNull();
    });
  });

  // ---------------------------------------------------------------------------
  // removePhoto
  // ---------------------------------------------------------------------------
  describe('removePhoto', () => {
    it('removes the photo with the matching uri', () => {
      const { result } = renderHook(() => usePhotos());
      const photo = makePhoto('ASME-006', 'file:///img1.jpg');

      act(() => result.current.addPhoto(photo));
      act(() => result.current.removePhoto('ASME-006', 'file:///img1.jpg'));

      expect(result.current.getPhotos('ASME-006')).toHaveLength(0);
    });

    it('removes only the targeted photo, leaving others intact', () => {
      const { result } = renderHook(() => usePhotos());
      act(() => {
        result.current.addPhoto(makePhoto('ASME-006', 'file:///img1.jpg'));
        result.current.addPhoto(makePhoto('ASME-006', 'file:///img2.jpg'));
      });

      act(() => result.current.removePhoto('ASME-006', 'file:///img1.jpg'));

      const remaining = result.current.getPhotos('ASME-006');
      expect(remaining).toHaveLength(1);
      expect(remaining[0].uri).toBe('file:///img2.jpg');
    });

    it('does nothing when ruleId has no photos', () => {
      const { result } = renderHook(() => usePhotos());
      // Should not throw.
      act(() => result.current.removePhoto('ASME-999', 'file:///img.jpg'));
      expect(result.current.getPhotos('ASME-999')).toHaveLength(0);
    });

    it('does nothing when uri does not match any photo', () => {
      const { result } = renderHook(() => usePhotos());
      act(() => result.current.addPhoto(makePhoto('ASME-006', 'file:///img1.jpg')));
      act(() => result.current.removePhoto('ASME-006', 'file:///nonexistent.jpg'));

      expect(result.current.getPhotos('ASME-006')).toHaveLength(1);
    });

    it('does not affect photos for other ruleIds', () => {
      const { result } = renderHook(() => usePhotos());
      act(() => {
        result.current.addPhoto(makePhoto('ASME-006', 'file:///a.jpg'));
        result.current.addPhoto(makePhoto('ASME-011', 'file:///b.jpg'));
      });

      act(() => result.current.removePhoto('ASME-006', 'file:///a.jpg'));

      expect(result.current.getPhotos('ASME-011')).toHaveLength(1);
    });
  });

  // ---------------------------------------------------------------------------
  // getPhotos — reactive updates
  // ---------------------------------------------------------------------------
  describe('getPhotos reactivity', () => {
    it('reflects additions immediately', () => {
      const { result } = renderHook(() => usePhotos());

      expect(result.current.getPhotos('ASME-001')).toHaveLength(0);
      act(() => result.current.addPhoto(makePhoto('ASME-001', 'file:///a.jpg')));
      expect(result.current.getPhotos('ASME-001')).toHaveLength(1);
    });

    it('reflects removals immediately', () => {
      const { result } = renderHook(() => usePhotos());

      act(() => result.current.addPhoto(makePhoto('ASME-001', 'file:///a.jpg')));
      expect(result.current.getPhotos('ASME-001')).toHaveLength(1);

      act(() => result.current.removePhoto('ASME-001', 'file:///a.jpg'));
      expect(result.current.getPhotos('ASME-001')).toHaveLength(0);
    });
  });
});

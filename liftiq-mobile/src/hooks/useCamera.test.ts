/**
 * useCamera hook tests.
 *
 * expo-image-picker and expo-location both require native modules.
 * We mock them at the module boundary to test the hook's logic:
 * permission handling, successful capture with GPS, capture without GPS,
 * and cancelled camera sessions.
 */

import { renderHook, act, waitFor } from '@testing-library/react-native';

// ---------------------------------------------------------------------------
// Mock handles
// ---------------------------------------------------------------------------
const mockRequestCameraPermissions = jest.fn();
const mockRequestForegroundPermissions = jest.fn();
const mockLaunchCameraAsync = jest.fn();
const mockGetCurrentPosition = jest.fn();

jest.mock('expo-image-picker', () => ({
  requestCameraPermissionsAsync: () => mockRequestCameraPermissions(),
  launchCameraAsync: (...args: unknown[]) => mockLaunchCameraAsync(...args),
  MediaTypeOptions: { Images: 'Images' },
}));

jest.mock('expo-location', () => ({
  requestForegroundPermissionsAsync: () => mockRequestForegroundPermissions(),
  getCurrentPositionAsync: (...args: unknown[]) => mockGetCurrentPosition(...args),
  Accuracy: { Balanced: 3 },
}));

import { useCamera } from './useCamera';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function grantAll() {
  mockRequestCameraPermissions.mockResolvedValue({ granted: true });
  mockRequestForegroundPermissions.mockResolvedValue({ granted: true });
}
function denyCamera() {
  mockRequestCameraPermissions.mockResolvedValue({ granted: false });
  mockRequestForegroundPermissions.mockResolvedValue({ granted: true });
}
function denyLocation() {
  mockRequestCameraPermissions.mockResolvedValue({ granted: true });
  mockRequestForegroundPermissions.mockResolvedValue({ granted: false });
}
function mockLocation(lat: number, lon: number) {
  mockGetCurrentPosition.mockResolvedValue({
    coords: { latitude: lat, longitude: lon },
  });
}
function mockLocationError() {
  mockGetCurrentPosition.mockRejectedValue(new Error('Location unavailable'));
}
function mockCameraResult(uri: string) {
  mockLaunchCameraAsync.mockResolvedValue({
    canceled: false,
    assets: [{ uri }],
  });
}
function mockCameraCancelled() {
  mockLaunchCameraAsync.mockResolvedValue({ canceled: true, assets: [] });
}

beforeEach(() => {
  jest.clearAllMocks();
  grantAll();
  mockLocation(37.7749, -122.4194);
  mockCameraResult('file:///captured.jpg');
});

// ---------------------------------------------------------------------------
// Permission handling
// ---------------------------------------------------------------------------
describe('useCamera – permissions', () => {
  it('requests camera and location permissions on mount', async () => {
    renderHook(() => useCamera());
    await waitFor(() => {
      expect(mockRequestCameraPermissions).toHaveBeenCalledTimes(1);
      expect(mockRequestForegroundPermissions).toHaveBeenCalledTimes(1);
    });
  });

  it('sets hasPermission true when camera permission is granted', async () => {
    grantAll();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
  });

  it('sets hasPermission false when camera permission is denied', async () => {
    denyCamera();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(false));
  });

  it('still sets hasPermission true when only location is denied', async () => {
    denyLocation();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));
  });
});

// ---------------------------------------------------------------------------
// capture – successful with GPS
// ---------------------------------------------------------------------------
describe('useCamera – capture with GPS', () => {
  it('returns a CapturedPhoto with the correct uri', async () => {
    mockCameraResult('file:///photo1.jpg');
    mockLocation(51.5074, -0.1278);

    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-006');
    });

    expect(photo?.uri).toBe('file:///photo1.jpg');
  });

  it('returns a CapturedPhoto with the provided ruleId', async () => {
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-011');
    });

    expect(photo?.ruleId).toBe('ASME-011');
  });

  it('returns GPS coordinates from expo-location', async () => {
    mockLocation(40.7128, -74.006);
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-001');
    });

    expect(photo?.latitude).toBeCloseTo(40.7128);
    expect(photo?.longitude).toBeCloseTo(-74.006);
  });

  it('returns a valid ISO timestamp', async () => {
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-001');
    });

    expect(photo?.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
  });

  it('passes quality 0.7 and mediaTypes to launchCameraAsync', async () => {
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    await act(async () => {
      await result.current.capture('ASME-001');
    });

    expect(mockLaunchCameraAsync).toHaveBeenCalledWith(
      expect.objectContaining({ quality: 0.7 }),
    );
  });
});

// ---------------------------------------------------------------------------
// capture – location unavailable
// ---------------------------------------------------------------------------
describe('useCamera – capture without GPS', () => {
  it('returns null coordinates when location throws', async () => {
    mockLocationError();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-001');
    });

    expect(photo?.latitude).toBeNull();
    expect(photo?.longitude).toBeNull();
  });

  it('still returns the photo uri when location is unavailable', async () => {
    mockCameraResult('file:///noloc.jpg');
    mockLocationError();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-001');
    });

    expect(photo?.uri).toBe('file:///noloc.jpg');
  });
});

// ---------------------------------------------------------------------------
// capture – camera cancelled
// ---------------------------------------------------------------------------
describe('useCamera – cancelled capture', () => {
  it('returns null when the user cancels the camera', async () => {
    mockCameraCancelled();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    let photo: Awaited<ReturnType<typeof result.current.capture>> = null;
    await act(async () => {
      photo = await result.current.capture('ASME-006');
    });

    expect(photo).toBeNull();
  });

  it('does not call getCurrentPositionAsync when capture is cancelled', async () => {
    mockCameraCancelled();
    const { result } = renderHook(() => useCamera());
    await waitFor(() => expect(result.current.hasPermission).toBe(true));

    await act(async () => {
      await result.current.capture('ASME-006');
    });

    expect(mockGetCurrentPosition).not.toHaveBeenCalled();
  });
});

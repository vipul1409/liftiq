import { useCallback, useEffect, useState } from 'react';
import * as ImagePicker from 'expo-image-picker';
import * as Location from 'expo-location';
import type { CapturedPhoto } from '../store/photos';

interface UseCameraResult {
  hasPermission: boolean;
  capture: (ruleId: string) => Promise<CapturedPhoto | null>;
}

export function useCamera(): UseCameraResult {
  const [hasPermission, setHasPermission] = useState(false);

  useEffect(() => {
    Promise.all([
      ImagePicker.requestCameraPermissionsAsync(),
      Location.requestForegroundPermissionsAsync(),
    ]).then(([camera, location]) => {
      // Camera permission is required; location is best-effort.
      setHasPermission(camera.granted);
      if (!location.granted) {
        console.warn('[useCamera] Location permission denied — photos will have null coords');
      }
    });
  }, []);

  const capture = useCallback(async (ruleId: string): Promise<CapturedPhoto | null> => {
    const result = await ImagePicker.launchCameraAsync({
      mediaTypes: ['images'],
      quality: 0.7,
      allowsEditing: false,
    });

    if (result.canceled || !result.assets[0]) return null;

    const uri = result.assets[0].uri;
    const timestamp = new Date().toISOString();

    // Best-effort location — null if denied or unavailable.
    let latitude: number | null = null;
    let longitude: number | null = null;
    try {
      const loc = await Location.getCurrentPositionAsync({
        accuracy: Location.Accuracy.Balanced,
      });
      latitude = loc.coords.latitude;
      longitude = loc.coords.longitude;
    } catch {
      // Location unavailable — continue without coords.
    }

    return { uri, ruleId, timestamp, latitude, longitude };
  }, []);

  return { hasPermission, capture };
}

import { useState, useCallback } from 'react';

export interface CapturedPhoto {
  uri: string;
  ruleId: string;
  timestamp: string;        // ISO-8601
  latitude: number | null;
  longitude: number | null;
}

interface UsePhotosResult {
  photos: Map<string, CapturedPhoto[]>;
  addPhoto: (photo: CapturedPhoto) => void;
  removePhoto: (ruleId: string, uri: string) => void;
  getPhotos: (ruleId: string) => CapturedPhoto[];
}

export function usePhotos(): UsePhotosResult {
  const [photos, setPhotos] = useState<Map<string, CapturedPhoto[]>>(new Map());

  const addPhoto = useCallback((photo: CapturedPhoto) => {
    setPhotos((prev) => {
      const next = new Map(prev);
      const existing = next.get(photo.ruleId) ?? [];
      next.set(photo.ruleId, [...existing, photo]);
      return next;
    });
  }, []);

  const removePhoto = useCallback((ruleId: string, uri: string) => {
    setPhotos((prev) => {
      const next = new Map(prev);
      const existing = next.get(ruleId) ?? [];
      next.set(ruleId, existing.filter((p) => p.uri !== uri));
      return next;
    });
  }, []);

  const getPhotos = useCallback(
    (ruleId: string): CapturedPhoto[] => photos.get(ruleId) ?? [],
    [photos],
  );

  return { photos, addPhoto, removePhoto, getPhotos };
}

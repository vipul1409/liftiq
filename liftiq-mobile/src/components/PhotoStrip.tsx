import React from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { CapturedPhoto } from '../store/photos';

interface Props {
  photos: CapturedPhoto[];
  onRemove: (uri: string) => void;
}

export function PhotoStrip({ photos, onRemove }: Props) {
  if (photos.length === 0) return null;

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Evidence ({photos.length})</Text>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.scroll}>
        {photos.map((photo) => (
          <View key={photo.uri} style={styles.thumb}>
            <Image source={{ uri: photo.uri }} style={styles.image} />
            <TouchableOpacity
              style={styles.removeBtn}
              onPress={() => onRemove(photo.uri)}
              hitSlop={{ top: 6, bottom: 6, left: 6, right: 6 }}
            >
              <Text style={styles.removeText}>✕</Text>
            </TouchableOpacity>
            {(photo.latitude !== null && photo.longitude !== null) && (
              <Text style={styles.gps} numberOfLines={1}>
                {photo.latitude.toFixed(4)}, {photo.longitude.toFixed(4)}
              </Text>
            )}
          </View>
        ))}
      </ScrollView>
    </View>
  );
}

const THUMB = 72;

const styles = StyleSheet.create({
  container: {
    marginTop: 8,
  },
  label: {
    fontSize: 11,
    fontWeight: '600',
    color: '#6b7280',
    marginBottom: 6,
    textTransform: 'uppercase',
    letterSpacing: 0.4,
  },
  scroll: {
    flexDirection: 'row',
  },
  thumb: {
    marginRight: 8,
    position: 'relative',
  },
  image: {
    width: THUMB,
    height: THUMB,
    borderRadius: 6,
    backgroundColor: '#e5e7eb',
  },
  removeBtn: {
    position: 'absolute',
    top: -4,
    right: -4,
    width: 18,
    height: 18,
    borderRadius: 9,
    backgroundColor: '#ef4444',
    alignItems: 'center',
    justifyContent: 'center',
  },
  removeText: {
    color: '#fff',
    fontSize: 9,
    fontWeight: '700',
    lineHeight: 12,
  },
  gps: {
    width: THUMB,
    marginTop: 3,
    fontSize: 8,
    color: '#9ca3af',
    textAlign: 'center',
  },
});

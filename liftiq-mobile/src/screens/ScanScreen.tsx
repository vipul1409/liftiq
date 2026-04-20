import React, { useState } from 'react';
import {
  ActivityIndicator,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { UnitPicker } from '../components/UnitPicker';
import { useUnits } from '../hooks/useUnits';
import type { RootStackParamList } from '../navigation/AppNavigator';

type Props = NativeStackScreenProps<RootStackParamList, 'Scan'>;

export function ScanScreen({ navigation }: Props) {
  const { units, loading, error } = useUnits();
  const [pickerVisible, setPickerVisible] = useState(false);

  function handleSelect(tag: string) {
    setPickerVisible(false);
    navigation.navigate('Compliance', { unitTag: tag });
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.logo}>LiftIQ</Text>
        <Text style={styles.subtitle}>Elevator Inspection Platform</Text>

        <View style={styles.scanArea}>
          {loading ? (
            <ActivityIndicator size="large" color="#2563eb" />
          ) : error ? (
            <View style={styles.errorBox}>
              <Text style={styles.errorText}>Could not reach compliance engine</Text>
              <Text style={styles.errorDetail}>{error}</Text>
              <Text style={styles.errorHint}>Is the compliance engine running on port 8080?</Text>
            </View>
          ) : (
            <TouchableOpacity
              style={styles.scanButton}
              onPress={() => setPickerVisible(true)}
              activeOpacity={0.8}
            >
              <Text style={styles.scanIcon}>⬡</Text>
              <Text style={styles.scanLabel}>Scan Elevator Tag</Text>
              <Text style={styles.scanHint}>NFC stub — tap to select unit</Text>
            </TouchableOpacity>
          )}
        </View>

        <Text style={styles.footer}>Phase 1 · Week 4 prototype</Text>
      </View>

      <UnitPicker
        visible={pickerVisible}
        units={units}
        onSelect={handleSelect}
        onClose={() => setPickerVisible(false)}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  content: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 32,
  },
  logo: {
    fontSize: 42,
    fontWeight: '800',
    color: '#1e40af',
    letterSpacing: -1,
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 15,
    color: '#6b7280',
    marginBottom: 64,
  },
  scanArea: {
    width: '100%',
    alignItems: 'center',
    marginBottom: 64,
  },
  scanButton: {
    backgroundColor: '#2563eb',
    borderRadius: 20,
    paddingVertical: 40,
    paddingHorizontal: 48,
    alignItems: 'center',
    shadowColor: '#2563eb',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.3,
    shadowRadius: 16,
    elevation: 8,
  },
  scanIcon: {
    fontSize: 48,
    color: '#93c5fd',
    marginBottom: 12,
  },
  scanLabel: {
    fontSize: 20,
    fontWeight: '700',
    color: '#fff',
    marginBottom: 6,
  },
  scanHint: {
    fontSize: 13,
    color: '#bfdbfe',
  },
  errorBox: {
    backgroundColor: '#fef2f2',
    borderRadius: 12,
    padding: 20,
    borderWidth: 1,
    borderColor: '#fecaca',
    alignItems: 'center',
    width: '100%',
  },
  errorText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#991b1b',
    marginBottom: 4,
  },
  errorDetail: {
    fontSize: 13,
    color: '#b91c1c',
    marginBottom: 8,
    textAlign: 'center',
  },
  errorHint: {
    fontSize: 12,
    color: '#9ca3af',
    textAlign: 'center',
  },
  footer: {
    position: 'absolute',
    bottom: 24,
    fontSize: 12,
    color: '#d1d5db',
  },
});

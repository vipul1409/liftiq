import React, { useState } from 'react';
import {
  Alert,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import * as FileSystem from 'expo-file-system/legacy';
import * as Sharing from 'expo-sharing';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { SignatureCapture } from '../components/SignatureCapture';
import type { RootStackParamList } from '../navigation/AppNavigator';
import type { ReportRequest, ReportRuleResult } from '../types/report';
import { generateReport } from '../api/report';

type Props = NativeStackScreenProps<RootStackParamList, 'Signature'>;

export function SignatureScreen({ route, navigation }: Props) {
  const { unitTag, summary, asOf, results, photos, technician } = route.params;

  const [hasStrokes, setHasStrokes] = useState(false);
  const [signatureDataURI, setSignatureDataURI] = useState('');
  const [clearTrigger, setClearTrigger] = useState(0);
  const [generating, setGenerating] = useState(false);

  async function handleSignAndGenerate() {
    if (!hasStrokes) return;
    setGenerating(true);
    try {
      const reportPhotos = await Promise.all(
        photos.map(async (p) => {
          let dataUri = p.uri;
          if (!p.uri.startsWith('data:')) {
            const base64 = await FileSystem.readAsStringAsync(p.uri, {
              encoding: FileSystem.EncodingType.Base64,
            });
            dataUri = `data:image/jpeg;base64,${base64}`;
          }
          return {
            rule_id: p.ruleId,
            data_uri: dataUri,
            timestamp: p.timestamp,
            latitude: p.latitude,
            longitude: p.longitude,
          };
        }),
      );

      const reportResults: ReportRuleResult[] = results.map((r) => ({
        rule_id: r.rule_id,
        description: r.description,
        asme_ref: r.asme_ref,
        metric: r.metric,
        value: r.value ?? 0,
        threshold: r.threshold,
        unit: r.unit,
        status: r.status,
        message: r.message,
      }));

      const req: ReportRequest = {
        unit_tag: unitTag,
        inspected_at: asOf,
        technician,
        results: reportResults,
        summary: {
          pass: summary.pass,
          fail: summary.fail,
          unknown: summary.unknown,
          overall: summary.overall,
        },
        photos: reportPhotos,
        signature_data_uri: signatureDataURI,
      };

      const pdfBase64 = await generateReport(req);

      const filename = `liftiq-report-${unitTag}-${Date.now()}.pdf`;
      const fileUri = `${FileSystem.cacheDirectory}${filename}`;
      await FileSystem.writeAsStringAsync(fileUri, pdfBase64, {
        encoding: FileSystem.EncodingType.Base64,
      });

      const canShare = await Sharing.isAvailableAsync();
      if (canShare) {
        await Sharing.shareAsync(fileUri, {
          mimeType: 'application/pdf',
          dialogTitle: `LiftIQ Report — ${unitTag}`,
        });
      } else {
        Alert.alert('Report saved', `PDF saved to: ${fileUri}`);
      }

      // Return to scan screen after sharing
      navigation.navigate('Scan');
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      Alert.alert('Report generation failed', message);
    } finally {
      setGenerating(false);
    }
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Sign to Certify</Text>
        <Text style={styles.subtitle}>
          {technician} · {new Date(asOf).toLocaleDateString()}
        </Text>
        <Text style={styles.instruction}>
          Draw your signature below to certify the inspection results
        </Text>
      </View>

      <View style={styles.canvasArea}>
        <SignatureCapture
          onStrokesChange={setHasStrokes}
          onDataURIChange={setSignatureDataURI}
          clearTrigger={clearTrigger}
          width={320}
          height={200}
        />
        <Text style={styles.sigLabel}>
          {hasStrokes ? 'Signature captured' : 'Sign above'}
        </Text>
      </View>

      <View style={styles.buttons}>
        <TouchableOpacity
          style={styles.clearButton}
          onPress={() => setClearTrigger((t) => t + 1)}
          disabled={generating}
        >
          <Text style={styles.clearButtonText}>Clear</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[
            styles.signButton,
            (!hasStrokes || generating) && styles.buttonDisabled,
          ]}
          onPress={handleSignAndGenerate}
          disabled={!hasStrokes || generating}
        >
          <Text style={styles.signButtonText}>
            {generating ? 'Generating…' : 'Sign & Generate PDF'}
          </Text>
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  header: {
    paddingHorizontal: 24,
    paddingTop: 24,
    paddingBottom: 16,
    alignItems: 'center',
  },
  title: {
    fontSize: 22,
    fontWeight: '800',
    color: '#111827',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#6b7280',
    marginBottom: 8,
  },
  instruction: {
    fontSize: 13,
    color: '#9ca3af',
    textAlign: 'center',
  },
  canvasArea: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  sigLabel: {
    fontSize: 12,
    color: '#9ca3af',
    marginTop: 8,
  },
  buttons: {
    flexDirection: 'row',
    gap: 12,
    paddingHorizontal: 24,
    paddingBottom: 32,
  },
  clearButton: {
    flex: 1,
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    borderWidth: 1.5,
    borderColor: '#d1d5db',
    backgroundColor: '#fff',
  },
  clearButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#374151',
  },
  signButton: {
    flex: 2,
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    backgroundColor: '#2563eb',
  },
  buttonDisabled: {
    backgroundColor: '#93c5fd',
  },
  signButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#fff',
  },
});

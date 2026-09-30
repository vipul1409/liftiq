import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { ScanScreen } from '../screens/ScanScreen';
import { ComplianceScreen } from '../screens/ComplianceScreen';
import { SummaryScreen } from '../screens/SummaryScreen';
import { SignatureScreen } from '../screens/SignatureScreen';
import type { RuleResult } from '../types/compliance';
import type { Overrides } from '../utils/inspectionOutcome';
import type { CapturedPhoto } from '../store/photos';

export type RootStackParamList = {
  Scan: undefined;
  Compliance: { unitTag: string };
  Summary: {
    unitTag: string;
    asOf: string;
    /** Telemetry results as returned by the compliance engine. */
    results: RuleResult[];
    overrides: Overrides;
    photos: CapturedPhoto[];
    technician: string;
  };
  Signature: {
    unitTag: string;
    asOf: string;
    /** Telemetry results as returned by the compliance engine. */
    results: RuleResult[];
    overrides: Overrides;
    photos: CapturedPhoto[];
    technician: string;
  };
};

const Stack = createNativeStackNavigator<RootStackParamList>();

export function AppNavigator() {
  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#1e3a8a' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: '700' },
      }}
    >
      <Stack.Screen
        name="Scan"
        component={ScanScreen}
        options={{ title: 'LiftIQ', headerShown: false }}
      />
      <Stack.Screen
        name="Compliance"
        component={ComplianceScreen}
        options={({ route }) => ({
          title: route.params.unitTag,
        })}
      />
      <Stack.Screen
        name="Summary"
        component={SummaryScreen}
        options={{ title: 'Inspection Summary' }}
      />
      <Stack.Screen
        name="Signature"
        component={SignatureScreen}
        options={{ title: 'Certify Inspection' }}
      />
    </Stack.Navigator>
  );
}

import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { ScanScreen } from '../screens/ScanScreen';
import { ComplianceScreen } from '../screens/ComplianceScreen';
import { SummaryScreen } from '../screens/SummaryScreen';
import type { ComplianceSummary } from '../types/compliance';

export type RootStackParamList = {
  Scan: undefined;
  Compliance: { unitTag: string };
  Summary: { unitTag: string; summary: ComplianceSummary; asOf: string };
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
    </Stack.Navigator>
  );
}

import { Routes, Route, Navigate } from 'react-router';
import { ScanScreen } from './screens/ScanScreen';
import { ComplianceScreen } from './screens/ComplianceScreen';
import { SummaryScreen } from './screens/SummaryScreen';
import { SignatureScreen } from './screens/SignatureScreen';

export function App() {
  return (
    <Routes>
      <Route path="/" element={<ScanScreen />} />
      <Route path="/inspect/:tag" element={<ComplianceScreen />} />
      <Route path="/summary" element={<SummaryScreen />} />
      <Route path="/sign" element={<SignatureScreen />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

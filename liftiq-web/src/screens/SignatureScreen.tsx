import { useState } from 'react';
import { useNavigate } from 'react-router';
import { SignatureCanvas } from '../components/SignatureCanvas';
import { useInspection } from '../context/InspectionContext';
import { generateReport, downloadPDF } from '../api/report';
import type { ReportRequest, ReportRuleResult } from '../types/report';

export function SignatureScreen() {
  const navigate = useNavigate();
  const inspection = useInspection();

  const [hasStrokes, setHasStrokes] = useState(false);
  const [signatureDataURI, setSignatureDataURI] = useState('');
  const [clearTrigger, setClearTrigger] = useState(0);
  const [generating, setGenerating] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  if (!inspection.complianceData || !inspection.unitTag) {
    navigate('/');
    return null;
  }

  const { unitTag, complianceData, technician } = inspection;
  const asOf = complianceData.as_of;
  const allPhotos = inspection.getAllPhotos();

  async function handleSignAndGenerate() {
    if (!hasStrokes) return;
    setGenerating(true);
    setErrorMsg('');

    try {
      // Send telemetry results untouched plus the overrides; the report-generator
      // derives effective statuses and the summary.
      const reportResults: ReportRuleResult[] = complianceData.results.map((r) => ({
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

      const reportPhotos = allPhotos.map((p) => ({
        rule_id: p.ruleId,
        data_uri: p.uri,
        timestamp: p.timestamp,
        latitude: p.latitude,
        longitude: p.longitude,
      }));

      const req: ReportRequest = {
        unit_tag: unitTag!,
        inspected_at: asOf,
        technician,
        results: reportResults,
        overrides: Object.fromEntries(inspection.overrides),
        photos: reportPhotos,
        signature_data_uri: signatureDataURI,
      };

      const blob = await generateReport(req);
      downloadPDF(blob, unitTag!);

      // Return to scan screen
      navigate('/');
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setGenerating(false);
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <div className="px-6 pt-8 pb-4 text-center">
        <h1 className="text-[22px] font-extrabold text-gray-900 mb-1">Sign to Certify</h1>
        <p className="text-sm font-semibold text-gray-500">
          {technician} &middot; {new Date(asOf).toLocaleDateString()}
        </p>
        <p className="text-[13px] text-gray-400 mt-1">
          Draw your signature below to certify the inspection results
        </p>
      </div>

      {/* Canvas */}
      <div className="flex-1 flex flex-col items-center justify-center px-6">
        <SignatureCanvas
          width={400}
          height={200}
          onStrokesChange={setHasStrokes}
          onDataURIChange={setSignatureDataURI}
          clearTrigger={clearTrigger}
        />
        <p className="text-xs text-gray-400 mt-2">
          {hasStrokes ? 'Signature captured' : 'Sign above'}
        </p>
      </div>

      {/* Error */}
      {errorMsg && (
        <div className="mx-6 mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-700 text-sm font-semibold">Report generation failed</p>
          <p className="text-red-600 text-xs mt-1">{errorMsg}</p>
        </div>
      )}

      {/* Buttons */}
      <div className="flex gap-3 px-6 pb-8">
        <button
          onClick={() => setClearTrigger((t) => t + 1)}
          disabled={generating}
          className="flex-1 py-4 border-[1.5px] border-gray-300 bg-white text-gray-700 font-semibold rounded-xl hover:bg-gray-50 disabled:opacity-50"
        >
          Clear
        </button>
        <button
          onClick={handleSignAndGenerate}
          disabled={!hasStrokes || generating}
          className={`flex-[2] py-4 text-white font-bold rounded-xl transition-colors ${
            !hasStrokes || generating
              ? 'bg-blue-300 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700'
          }`}
        >
          {generating ? 'Generating...' : 'Sign & Generate PDF'}
        </button>
      </div>
    </div>
  );
}

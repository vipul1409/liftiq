import { useNavigate } from 'react-router';
import { useInspection } from '../context/InspectionContext';
import type { RuleStatus } from '../types/compliance';

const BANNER_COLORS: Record<RuleStatus, string> = {
  pass: 'bg-green-600',
  fail: 'bg-red-600',
  unknown: 'bg-gray-500',
};

const BANNER_LABELS: Record<RuleStatus, string> = {
  pass: 'ALL CHECKS PASSED',
  fail: 'INSPECTION FAILED',
  unknown: 'INCOMPLETE DATA',
};

export function SummaryScreen() {
  const navigate = useNavigate();
  const inspection = useInspection();
  const summary = inspection.getEffectiveSummary();
  const reviewCount = inspection.getOverridesNeedingReview().length;
  const allPhotos = inspection.getAllPhotos();

  if (!inspection.complianceData || !inspection.unitTag) {
    navigate('/');
    return null;
  }

  const formattedDate = new Date(inspection.complianceData.as_of).toLocaleString();

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Banner */}
      <div className={`${BANNER_COLORS[summary.overall]} py-10 px-6 text-center`}>
        <p className="text-white/80 text-base font-semibold tracking-wider mb-2">
          {inspection.unitTag}
        </p>
        <h1 className="text-3xl font-extrabold text-white tracking-wide mb-2">
          {BANNER_LABELS[summary.overall]}
        </h1>
        <p className="text-white/70 text-sm">{formattedDate}</p>
      </div>

      {/* Counts */}
      <div className="flex-1 px-6 py-8 max-w-lg mx-auto w-full">
        <div className="flex gap-3 mb-6">
          <div className="flex-1 bg-green-100 rounded-xl py-5 text-center">
            <p className="text-3xl font-extrabold text-gray-900">{summary.pass}</p>
            <p className="text-sm font-semibold text-gray-500 mt-1">Pass</p>
          </div>
          <div className="flex-1 bg-red-100 rounded-xl py-5 text-center">
            <p className="text-3xl font-extrabold text-gray-900">{summary.fail}</p>
            <p className="text-sm font-semibold text-gray-500 mt-1">Fail</p>
          </div>
          <div className="flex-1 bg-gray-100 rounded-xl py-5 text-center">
            <p className="text-3xl font-extrabold text-gray-900">{summary.unknown}</p>
            <p className="text-sm font-semibold text-gray-500 mt-1">Unknown</p>
          </div>
        </div>

        <p className="text-sm text-gray-400 text-center mb-2">
          {summary.pass + summary.fail + summary.unknown} ASME A17.1 checks evaluated
        </p>

        {allPhotos.length > 0 && (
          <p className="text-sm text-gray-500 text-center">
            {allPhotos.length} photo{allPhotos.length !== 1 ? 's' : ''} attached
          </p>
        )}
      </div>

      {/* Actions */}
      <div className="px-6 pb-8 max-w-lg mx-auto w-full space-y-3">
        {reviewCount > 0 && (
          <p className="text-center text-sm font-semibold text-amber-700">
            {reviewCount} override{reviewCount === 1 ? ' needs' : 's need'} review before signing
          </p>
        )}
        <button
          onClick={() => navigate('/sign')}
          disabled={reviewCount > 0}
          className="w-full py-4 bg-blue-600 text-white font-bold text-base rounded-xl hover:bg-blue-700 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Proceed to Sign
        </button>
        <button
          onClick={() => navigate(-1)}
          className="w-full py-4 bg-blue-800 text-white font-bold text-base rounded-xl hover:bg-blue-900 transition-colors"
        >
          Back to Checklist
        </button>
      </div>
    </div>
  );
}

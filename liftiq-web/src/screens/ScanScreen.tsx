import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import { useUnits } from '../hooks/useUnits';
import { useInspection } from '../context/InspectionContext';

export function ScanScreen() {
  const navigate = useNavigate();
  const { units, loading, error } = useUnits();
  const { setUnitTag, reset } = useInspection();
  const [scanning, setScanning] = useState(false);
  const [pickerOpen, setPickerOpen] = useState(false);

  // Reset inspection state when landing on scan screen
  useEffect(() => {
    reset();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  function handleScan() {
    setScanning(true);
    setTimeout(() => {
      setScanning(false);
      setPickerOpen(true);
    }, 1000);
  }

  function handleSelectUnit(tag: string) {
    setPickerOpen(false);
    setUnitTag(tag);
    navigate(`/inspect/${encodeURIComponent(tag)}`);
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-8">
      {/* Logo / Brand */}
      <div className="mb-12 text-center">
        <h1 className="text-5xl font-extrabold text-blue-800 tracking-tight">LiftIQ</h1>
        <p className="text-gray-500 mt-2 text-lg">Elevator Inspection Platform</p>
      </div>

      {/* Scan Button */}
      <div className="relative mb-8">
        {scanning && (
          <span className="absolute inset-0 rounded-full border-4 border-blue-400/50 animate-pulse-ring" />
        )}
        <button
          onClick={handleScan}
          disabled={loading || scanning}
          className={`w-48 h-48 rounded-full flex flex-col items-center justify-center text-white font-bold text-lg shadow-xl transition-transform hover:scale-105 ${
            scanning
              ? 'bg-blue-400 cursor-wait'
              : loading
              ? 'bg-gray-300 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700'
          }`}
        >
          {scanning ? (
            <>
              <span className="text-3xl mb-1">...</span>
              <span className="text-sm">Scanning</span>
            </>
          ) : loading ? (
            <>
              <span className="text-3xl mb-1">...</span>
              <span className="text-sm">Loading</span>
            </>
          ) : (
            <>
              <span className="text-4xl mb-1">&#x1F4F1;</span>
              <span>Scan Elevator</span>
              <span className="text-xs font-normal opacity-80 mt-1">NFC Tag</span>
            </>
          )}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg px-6 py-4 max-w-md text-center">
          <p className="text-red-800 font-semibold mb-1">Connection Error</p>
          <p className="text-red-600 text-sm">{error}</p>
          <p className="text-gray-500 text-xs mt-2">
            Make sure the backend is running (make demo-up)
          </p>
        </div>
      )}

      {/* Unit Picker Modal */}
      {pickerOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl shadow-2xl w-full max-w-sm mx-4 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-100">
              <h2 className="text-lg font-bold text-gray-900">Select Elevator Unit</h2>
              <p className="text-xs text-gray-400 mt-0.5">NFC stub — select unit manually</p>
            </div>
            <div className="max-h-[320px] overflow-y-auto">
              {units.map((unit) => (
                <button
                  key={unit}
                  onClick={() => handleSelectUnit(unit)}
                  className="w-full px-6 py-4 text-left hover:bg-blue-50 flex justify-between items-center border-b border-gray-50 transition-colors"
                >
                  <span className="font-semibold text-gray-800">{unit}</span>
                  <span className="text-gray-400 text-lg">&rsaquo;</span>
                </button>
              ))}
            </div>
            <div className="px-6 py-3 border-t border-gray-100">
              <button
                onClick={() => setPickerOpen(false)}
                className="w-full py-2.5 text-sm font-semibold text-gray-500 hover:text-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Hint */}
      <p className="text-gray-400 text-sm mt-4 max-w-xs text-center">
        {units.length > 0
          ? `${units.length} elevator${units.length !== 1 ? 's' : ''} available`
          : loading
          ? 'Connecting to compliance engine...'
          : ''}
      </p>
    </div>
  );
}

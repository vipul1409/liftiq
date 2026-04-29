import { useRef } from 'react';
import { StatusBadge } from './StatusBadge';
import { PhotoStrip } from './PhotoStrip';
import { effectiveStatus } from '../hooks/useOverrides';
import { fileToDataURI, type CapturedPhoto } from '../hooks/usePhotos';
import type { RuleResult } from '../types/compliance';

interface Props {
  result: RuleResult;
  override: 'pass' | 'fail' | undefined;
  onOverride: (ruleId: string, status: 'pass' | 'fail') => void;
  onClearOverride: (ruleId: string) => void;
  isActive?: boolean;
  photos?: CapturedPhoto[];
  onAddPhoto?: (photo: CapturedPhoto) => void;
  onRemovePhoto?: (ruleId: string, uri: string) => void;
}

export function RuleRow({
  result,
  override,
  onOverride,
  onClearOverride,
  isActive = false,
  photos = [],
  onAddPhoto,
  onRemovePhoto,
}: Props) {
  const effective = effectiveStatus(result.status, override);
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function handleFileSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !onAddPhoto) return;

    const dataUri = await fileToDataURI(file);

    let latitude: number | null = null;
    let longitude: number | null = null;
    try {
      const pos = await new Promise<GeolocationPosition>((resolve, reject) =>
        navigator.geolocation.getCurrentPosition(resolve, reject, { timeout: 5000 }),
      );
      latitude = pos.coords.latitude;
      longitude = pos.coords.longitude;
    } catch {
      // GPS best-effort
    }

    onAddPhoto({
      uri: dataUri,
      ruleId: result.rule_id,
      timestamp: new Date().toISOString(),
      latitude,
      longitude,
    });

    // Reset input so the same file can be re-selected
    e.target.value = '';
  }

  return (
    <div
      className={`px-4 py-3.5 bg-white ${
        isActive ? 'border-l-[3px] border-l-blue-600 bg-blue-50 pl-[13px]' : ''
      }`}
    >
      <div className="flex justify-between items-center mb-1">
        <span className="text-[11px] text-gray-400 font-semibold tracking-wide">
          {result.rule_id}
        </span>
        <StatusBadge status={effective} overridden={override !== undefined} />
      </div>
      <p className="text-[15px] font-semibold text-gray-900 mb-0.5">
        {result.description}
      </p>
      <p className="text-xs text-gray-500 italic mb-1">{result.asme_ref}</p>
      {result.value !== null && (
        <p className="text-[13px] text-gray-700 mb-0.5">
          {result.value} {result.unit} (threshold: {result.threshold} {result.unit})
        </p>
      )}
      <p className="text-[13px] text-gray-500 mb-2">{result.message}</p>

      <div className="flex gap-2 flex-wrap">
        <button
          onClick={() => onOverride(result.rule_id, 'pass')}
          className={`px-3 py-1 rounded-md border text-xs font-semibold text-gray-700 border-green-300 bg-green-50 hover:bg-green-100 ${
            override === 'pass' ? 'opacity-50' : ''
          }`}
        >
          Override Pass
        </button>
        <button
          onClick={() => onOverride(result.rule_id, 'fail')}
          className={`px-3 py-1 rounded-md border text-xs font-semibold text-gray-700 border-red-300 bg-red-50 hover:bg-red-100 ${
            override === 'fail' ? 'opacity-50' : ''
          }`}
        >
          Override Fail
        </button>
        {override !== undefined && (
          <button
            onClick={() => onClearOverride(result.rule_id)}
            className="px-3 py-1 rounded-md border text-xs font-semibold text-gray-700 border-gray-300 bg-gray-50 hover:bg-gray-100"
          >
            Clear
          </button>
        )}
        {onAddPhoto && (
          <>
            <button
              onClick={() => fileInputRef.current?.click()}
              className="px-3 py-1 rounded-md border text-xs font-semibold text-gray-700 border-blue-300 bg-blue-50 hover:bg-blue-100"
            >
              Photo
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleFileSelected}
            />
          </>
        )}
      </div>

      {onRemovePhoto && (
        <PhotoStrip
          photos={photos}
          onRemove={(uri) => onRemovePhoto(result.rule_id, uri)}
        />
      )}
    </div>
  );
}

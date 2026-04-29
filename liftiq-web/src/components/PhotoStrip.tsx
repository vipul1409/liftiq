import type { CapturedPhoto } from '../hooks/usePhotos';

interface Props {
  photos: CapturedPhoto[];
  onRemove: (uri: string) => void;
}

export function PhotoStrip({ photos, onRemove }: Props) {
  if (photos.length === 0) return null;

  return (
    <div className="mt-2">
      <p className="text-[11px] font-semibold text-gray-500 uppercase tracking-wide mb-1.5">
        Evidence ({photos.length})
      </p>
      <div className="flex gap-2 overflow-x-auto">
        {photos.map((photo) => (
          <div key={photo.uri} className="relative shrink-0">
            <img
              src={photo.uri}
              alt="evidence"
              className="w-[72px] h-[72px] rounded-md object-cover bg-gray-200"
            />
            <button
              onClick={() => onRemove(photo.uri)}
              className="absolute -top-1 -right-1 w-[18px] h-[18px] rounded-full bg-red-500 text-white text-[9px] font-bold flex items-center justify-center hover:bg-red-600"
            >
              x
            </button>
            {photo.latitude !== null && photo.longitude !== null && (
              <p className="w-[72px] mt-0.5 text-[8px] text-gray-400 text-center truncate">
                {photo.latitude.toFixed(4)}, {photo.longitude.toFixed(4)}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

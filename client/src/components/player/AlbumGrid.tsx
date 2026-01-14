import { useAlbums, useAlbum } from "@/hooks/player/use-albums";
import { formatMediaUrl } from "@/lib/utils";
import { Disc } from "lucide-react";

const i18nText = {
  noAlbums: "No albums available",
  loading: "Loading albums...",
  unknownYear: "Unknown",
};

interface AlbumCardProps {
  albumId: number;
  onSelect: (albumId: number) => void;
}

function AlbumCard({ albumId, onSelect }: AlbumCardProps) {
  const { data: album } = useAlbum(albumId);

  if (!album) return null;

  return (
    <div
      onClick={() => onSelect(albumId)}
      className="group cursor-pointer animate-in fade-in slide-in-from-bottom-4"
    >
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-gray-100 to-gray-200 shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:from-gray-800 dark:to-gray-900">
        {/* Album Cover */}
        <div className="aspect-square overflow-hidden">
          {album.cover ? (
            <img
              src={formatMediaUrl(album.cover)}
              alt={album.name}
              className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
              <Disc className="h-20 w-20 text-white opacity-50" />
            </div>
          )}
        </div>

        {/* Hover Overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100">
          <div className="absolute bottom-4 left-4 right-4">
            <div className="flex items-center justify-center">
              <button className="rounded-full bg-white/90 p-3 shadow-lg backdrop-blur-sm transition-transform duration-200 hover:scale-110 hover:bg-white">
                <Disc className="h-6 w-6 text-gray-900" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Album Info */}
      <div className="mt-3 space-y-1">
        <h3 className="truncate font-bold text-gray-900 dark:text-white">
          {album.name}
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {album.year || i18nText.unknownYear}
        </p>
      </div>
    </div>
  );
}

interface AlbumGridProps {
  onSelectAlbum?: (albumId: number) => void;
}

export default function AlbumGrid({ onSelectAlbum }: AlbumGridProps) {
  const { data: albumIds = [], isLoading } = useAlbums();

  const handleSelect = (albumId: number) => {
    console.log("Selected album:", albumId);
    onSelectAlbum?.(albumId);
  };

  if (isLoading) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-400 dark:text-gray-500">{i18nText.loading}</p>
        </div>
      </div>
    );
  }

  if (albumIds.length === 0) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-400 dark:text-gray-500">
            {i18nText.noAlbums}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {albumIds.map((id) => (
        <AlbumCard key={id} albumId={id} onSelect={handleSelect} />
      ))}
    </div>
  );
}

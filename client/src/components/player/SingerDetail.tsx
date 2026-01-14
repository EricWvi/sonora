import { useSinger } from "@/hooks/player/use-singers";
import { useSingles, useTrack } from "@/hooks/player/use-tracks";
import { useAlbums, useAlbum } from "@/hooks/player/use-albums";
import { formatMediaUrl } from "@/lib/utils";
import { Play, Shuffle, Clock, ArrowLeft, Users, Disc } from "lucide-react";

const i18nText = {
  playAll: "Play All",
  shuffle: "Shuffle",
  back: "Back",
  singles: "Singles",
  albums: "Albums",
  duration: "Duration",
  loading: "Loading...",
  noSingles: "No singles available",
  noAlbums: "No albums available",
};

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

interface TrackRowProps {
  trackId: number;
  index: number;
  onPlay: (trackId: number) => void;
}

function TrackRow({ trackId, index, onPlay }: TrackRowProps) {
  const { data: track } = useTrack(trackId);

  if (!track) return null;

  return (
    <div
      onClick={() => onPlay(trackId)}
      className="group grid grid-cols-[auto_1fr_auto] gap-4 rounded-lg border border-transparent px-4 py-3 transition-all duration-200 hover:border-gray-200/50 hover:bg-gradient-to-r hover:from-purple-50/50 hover:to-pink-50/50 hover:shadow-md dark:hover:border-gray-700/50 dark:hover:from-purple-900/10 dark:hover:to-pink-900/10"
    >
      {/* Track Number / Play Button */}
      <div className="flex w-8 items-center justify-center text-sm font-medium text-gray-500 dark:text-gray-400">
        <span className="group-hover:hidden">{index + 1}</span>
        <Play className="hidden h-4 w-4 fill-purple-600 text-purple-600 group-hover:block dark:fill-purple-400 dark:text-purple-400" />
      </div>

      {/* Track Name */}
      <div className="min-w-0">
        <p className="truncate font-semibold text-gray-900 dark:text-white">
          {track.name}
        </p>
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {track.albumText || "Single"}
        </p>
      </div>

      {/* Duration */}
      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <Clock className="h-3 w-3" />
        {formatDuration(track.duration)}
      </div>
    </div>
  );
}

interface AlbumCardProps {
  albumId: number;
  onSelect: (albumId: number) => void;
}

function AlbumCard({ albumId, onSelect }: AlbumCardProps) {
  const { data: album } = useAlbum(albumId);

  if (!album) return null;

  return (
    <div onClick={() => onSelect(albumId)} className="group cursor-pointer">
      <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-gray-100 to-gray-200 shadow-md transition-all duration-300 hover:scale-105 hover:shadow-xl dark:from-gray-800 dark:to-gray-900">
        <div className="aspect-square overflow-hidden">
          {album.cover ? (
            <img
              src={formatMediaUrl(album.cover)}
              alt={album.name}
              className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
              <Disc className="h-12 w-12 text-white opacity-50" />
            </div>
          )}
        </div>
      </div>
      <div className="mt-2">
        <h4 className="truncate text-sm font-bold text-gray-900 dark:text-white">
          {album.name}
        </h4>
        <p className="text-xs text-gray-600 dark:text-gray-400">{album.year}</p>
      </div>
    </div>
  );
}

interface SingerDetailProps {
  singerId: number;
  onBack?: () => void;
  onPlayTrack?: (trackId: number) => void;
  onPlayAll?: (singerId: number) => void;
  onShuffle?: (singerId: number) => void;
  onSelectAlbum?: (albumId: number) => void;
}

export default function SingerDetail({
  singerId,
  onBack,
  onPlayTrack,
  onPlayAll,
  onShuffle,
  onSelectAlbum,
}: SingerDetailProps) {
  const { data: singer } = useSinger(singerId);
  const { data: singleIds = [], isLoading: singlesLoading } = useSingles();
  const { data: allAlbumIds = [] } = useAlbums();

  if (!singer) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-400 dark:text-gray-500">{i18nText.loading}</p>
        </div>
      </div>
    );
  }

  // Filter singles by singer name
  const singerSingles = singleIds.filter((id) => {
    // This will be filtered by the useSingles hook implementation
    return true; // Placeholder
  });

  // Filter albums by checking tracks for this singer
  // Note: This is simplified - in production you'd want a proper singer-album relationship
  const singerAlbumIds = allAlbumIds; // Placeholder - should filter by singer

  const handlePlayTrack = (trackId: number) => {
    console.log("Playing track:", trackId);
    onPlayTrack?.(trackId);
  };

  const handlePlayAll = () => {
    console.log("Playing all tracks from singer:", singerId);
    onPlayAll?.(singerId);
  };

  const handleShuffle = () => {
    console.log("Shuffling singer:", singerId);
    onShuffle?.(singerId);
  };

  const handleSelectAlbum = (albumId: number) => {
    console.log("Selected album:", albumId);
    onSelectAlbum?.(albumId);
  };

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 space-y-8">
      {/* Back Button */}
      {onBack && (
        <button
          onClick={onBack}
          className="group flex items-center gap-2 text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
        >
          <ArrowLeft className="h-5 w-5 transition-transform group-hover:-translate-x-1" />
          <span className="font-medium">{i18nText.back}</span>
        </button>
      )}

      {/* Hero Section */}
      <div className="grid gap-8 md:grid-cols-[auto_1fr]">
        {/* Singer Avatar */}
        <div className="w-full md:w-64">
          <div className="aspect-square overflow-hidden rounded-2xl shadow-2xl">
            {singer.avatar ? (
              <img
                src={formatMediaUrl(singer.avatar)}
                alt={singer.name}
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-pink-500 to-purple-500">
                <Users className="h-32 w-32 text-white opacity-50" />
              </div>
            )}
          </div>
        </div>

        {/* Singer Info */}
        <div className="flex flex-col justify-end space-y-4">
          <div>
            <p className="text-sm font-semibold uppercase tracking-wider text-gray-600 dark:text-gray-400">
              Artist
            </p>
            <h1 className="mt-2 bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-5xl font-black tracking-tight text-transparent">
              {singer.name}
            </h1>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-3">
            <button
              onClick={handlePlayAll}
              className="group flex items-center gap-2 rounded-full bg-gradient-to-r from-purple-600 to-pink-600 px-6 py-3 font-bold text-white shadow-lg transition-all duration-200 hover:scale-105 hover:shadow-xl"
            >
              <Play className="h-5 w-5 fill-white" />
              {i18nText.playAll}
            </button>
            <button
              onClick={handleShuffle}
              className="flex items-center gap-2 rounded-full border-2 border-gray-300 bg-white/80 px-6 py-3 font-bold text-gray-900 backdrop-blur-sm transition-all duration-200 hover:scale-105 hover:border-gray-400 hover:bg-white dark:border-gray-700 dark:bg-gray-800/80 dark:text-white dark:hover:border-gray-600 dark:hover:bg-gray-800"
            >
              <Shuffle className="h-5 w-5" />
              {i18nText.shuffle}
            </button>
          </div>
        </div>
      </div>

      {/* Singles Section */}
      <div className="rounded-2xl border border-gray-200/50 bg-white/80 p-6 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
          {i18nText.singles}
        </h2>

        {singlesLoading ? (
          <div className="flex items-center justify-center py-12">
            <p className="text-gray-400 dark:text-gray-500">
              {i18nText.loading}
            </p>
          </div>
        ) : singerSingles.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <p className="text-gray-400 dark:text-gray-500">
              {i18nText.noSingles}
            </p>
          </div>
        ) : (
          <div className="space-y-1">
            {singerSingles.map((id, index) => (
              <TrackRow
                key={id}
                trackId={id}
                index={index}
                onPlay={handlePlayTrack}
              />
            ))}
          </div>
        )}
      </div>

      {/* Albums Section */}
      {singerAlbumIds.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            {i18nText.albums}
          </h2>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            {singerAlbumIds.slice(0, 10).map((id) => (
              <AlbumCard key={id} albumId={id} onSelect={handleSelectAlbum} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

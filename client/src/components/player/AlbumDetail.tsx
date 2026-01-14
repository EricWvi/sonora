import { useAlbum } from "@/hooks/player/use-albums";
import { useAlbumTracks, useTrack } from "@/hooks/player/use-tracks";
import { formatMediaUrl } from "@/lib/utils";
import { Play, Shuffle, Clock, ArrowLeft, Disc } from "lucide-react";

const i18nText = {
  playAll: "Play All",
  shuffle: "Shuffle",
  back: "Back",
  tracks: "Tracks",
  duration: "Duration",
  loading: "Loading...",
  noTracks: "No tracks in this album",
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
        <span className="group-hover:hidden">
          {track.trackNumber || index + 1}
        </span>
        <Play className="hidden h-4 w-4 fill-purple-600 text-purple-600 group-hover:block dark:fill-purple-400 dark:text-purple-400" />
      </div>

      {/* Track Name */}
      <div className="min-w-0">
        <p className="truncate font-semibold text-gray-900 dark:text-white">
          {track.name}
        </p>
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {track.singer}
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

interface AlbumDetailProps {
  albumId: number;
  onBack?: () => void;
  onPlayTrack?: (trackId: number) => void;
  onPlayAll?: (albumId: number) => void;
  onShuffle?: (albumId: number) => void;
}

export default function AlbumDetail({
  albumId,
  onBack,
  onPlayTrack,
  onPlayAll,
  onShuffle,
}: AlbumDetailProps) {
  const { data: album } = useAlbum(albumId);
  const { data: trackIds = [], isLoading } = useAlbumTracks(albumId);

  const handlePlayTrack = (trackId: number) => {
    console.log("Playing track:", trackId);
    onPlayTrack?.(trackId);
  };

  const handlePlayAll = () => {
    console.log("Playing all tracks from album:", albumId);
    onPlayAll?.(albumId);
  };

  const handleShuffle = () => {
    console.log("Shuffling album:", albumId);
    onShuffle?.(albumId);
  };

  if (!album) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-400 dark:text-gray-500">{i18nText.loading}</p>
        </div>
      </div>
    );
  }

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
        {/* Album Cover */}
        <div className="w-full md:w-64">
          <div className="aspect-square overflow-hidden rounded-2xl shadow-2xl">
            {album.cover ? (
              <img
                src={formatMediaUrl(album.cover)}
                alt={album.name}
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
                <Disc className="h-32 w-32 text-white opacity-50" />
              </div>
            )}
          </div>
        </div>

        {/* Album Info */}
        <div className="flex flex-col justify-end space-y-4">
          <div>
            <p className="text-sm font-semibold uppercase tracking-wider text-gray-600 dark:text-gray-400">
              Album
            </p>
            <h1 className="mt-2 bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-5xl font-black tracking-tight text-transparent">
              {album.name}
            </h1>
            <p className="mt-3 text-lg text-gray-600 dark:text-gray-400">
              {album.year}
            </p>
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

      {/* Tracklist */}
      <div className="rounded-2xl border border-gray-200/50 bg-white/80 p-6 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
          {i18nText.tracks}
        </h2>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <p className="text-gray-400 dark:text-gray-500">
              {i18nText.loading}
            </p>
          </div>
        ) : trackIds.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <p className="text-gray-400 dark:text-gray-500">
              {i18nText.noTracks}
            </p>
          </div>
        ) : (
          <div className="space-y-1">
            {trackIds.map((id, index) => (
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
    </div>
  );
}

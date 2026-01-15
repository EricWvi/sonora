import { useTracks, useTrack } from "@/hooks/player/use-tracks";
import { useAlbum } from "@/hooks/player/use-albums";
import { useAudioState, useAudioControls } from "@/lib/AudioContext";
import { dbClient } from "@/lib/localCache";
import { Play, Pause, Clock } from "lucide-react";
import { cn } from "@/lib/utils";

const i18nText = {
  title: "Title",
  artist: "Artist",
  album: "Album",
  duration: "Duration",
  noTracks: "No tracks available",
  loading: "Loading tracks...",
};

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

interface TrackRowProps {
  trackId: number;
  index: number;
  isCurrentTrack: boolean;
  isPlaying: boolean;
  onPlay: (trackId: number) => void;
}

function TrackRow({
  trackId,
  index,
  isCurrentTrack,
  isPlaying,
  onPlay,
}: TrackRowProps) {
  const { data: track } = useTrack(trackId);
  const { data: album } = useAlbum(track?.album || 0);

  if (!track) return null;

  const albumName = album?.name || track.albumText || "Unknown";

  return (
    <div
      onClick={() => onPlay(trackId)}
      className={cn(
        "group grid grid-cols-[auto_1fr_auto_auto_auto] gap-4 rounded-lg border border-transparent px-4 py-3 transition-all duration-200 md:grid-cols-[auto_2fr_1.5fr_1fr_auto]",
        isCurrentTrack
          ? "border-purple-200/50 bg-gradient-to-r from-purple-50 to-pink-50 shadow-md dark:border-purple-700/50 dark:from-purple-900/20 dark:to-pink-900/20"
          : "hover:border-gray-200/50 hover:bg-gradient-to-r hover:from-purple-50/50 hover:to-pink-50/50 hover:shadow-md dark:hover:border-gray-700/50 dark:hover:from-purple-900/10 dark:hover:to-pink-900/10"
      )}
    >
      {/* Index / Play Button / Playing Indicator */}
      <div className="flex w-8 items-center justify-center text-sm font-medium text-gray-500 dark:text-gray-400">
        {isCurrentTrack && isPlaying ? (
          <div className="flex gap-0.5">
            <span
              className="h-3 w-0.5 animate-pulse rounded-full bg-purple-600 dark:bg-purple-400"
              style={{ animationDelay: "0ms" }}
            />
            <span
              className="h-4 w-0.5 animate-pulse rounded-full bg-purple-600 dark:bg-purple-400"
              style={{ animationDelay: "150ms" }}
            />
            <span
              className="h-2 w-0.5 animate-pulse rounded-full bg-purple-600 dark:bg-purple-400"
              style={{ animationDelay: "300ms" }}
            />
          </div>
        ) : isCurrentTrack ? (
          <Pause className="h-4 w-4 fill-purple-600 text-purple-600 dark:fill-purple-400 dark:text-purple-400" />
        ) : (
          <>
            <span className="group-hover:hidden">{index + 1}</span>
            <Play className="hidden h-4 w-4 fill-purple-600 text-purple-600 group-hover:block dark:fill-purple-400 dark:text-purple-400" />
          </>
        )}
      </div>

      {/* Title */}
      <div className="min-w-0">
        <p
          className={cn(
            "truncate font-semibold",
            isCurrentTrack
              ? "text-purple-600 dark:text-purple-400"
              : "text-gray-900 dark:text-white"
          )}
        >
          {track.name}
        </p>
      </div>

      {/* Artist */}
      <div className="hidden min-w-0 md:block">
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {track.singer}
        </p>
      </div>

      {/* Album */}
      <div className="hidden min-w-0 md:block">
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {albumName}
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

export default function SongList() {
  const { data: trackIds = [], isLoading } = useTracks();
  const { currentTrack, isPlaying } = useAudioState();
  const { playTracks, togglePlay } = useAudioControls();

  const handlePlay = async (trackId: number) => {
    // If clicking the current track, toggle play/pause
    if (currentTrack?.id === trackId) {
      togglePlay();
      return;
    }

    // Get all tracks and play from the selected index
    const tracks = await dbClient.getAllTracks();
    const index = tracks.findIndex((t) => t.id === trackId);
    if (index !== -1) {
      playTracks(tracks, index);
    }
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

  if (trackIds.length === 0) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80">
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-400 dark:text-gray-500">
            {i18nText.noTracks}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 space-y-4">
      {/* Header */}
      <div className="grid grid-cols-[auto_1fr_auto_auto_auto] gap-4 border-b border-gray-200 px-4 pb-3 text-xs font-bold uppercase tracking-wider text-gray-500 dark:border-gray-700 dark:text-gray-400 md:grid-cols-[auto_2fr_1.5fr_1fr_auto]">
        <div className="w-8 text-center">#</div>
        <div>{i18nText.title}</div>
        <div className="hidden md:block">{i18nText.artist}</div>
        <div className="hidden md:block">{i18nText.album}</div>
        <div>{i18nText.duration}</div>
      </div>

      {/* Track List */}
      <div className="space-y-1">
        {trackIds.map((id, index) => (
          <TrackRow
            key={id}
            trackId={id}
            index={index}
            isCurrentTrack={currentTrack?.id === id}
            isPlaying={isPlaying && currentTrack?.id === id}
            onPlay={handlePlay}
          />
        ))}
      </div>
    </div>
  );
}

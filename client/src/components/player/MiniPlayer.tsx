import { useAudioState, useAudioControls } from "@/lib/AudioContext";
import { formatMediaUrl, cn } from "@/lib/utils";
import ProgressBar from "./ProgressBar";
import PlaybackControls from "./PlaybackControls";
import { Music, ChevronUp, ListMusic } from "lucide-react";

const i18nText = {
  nowPlaying: "Now Playing",
  queue: "Queue",
  expand: "Expand player",
};

export default function MiniPlayer() {
  const state = useAudioState();
  const controls = useAudioControls();

  const {
    currentTrack,
    isPlaying,
    isLoading,
    currentTime,
    duration,
    queue,
    queueIndex,
    repeatMode,
    isShuffled,
  } = state;

  // Don't render if no track
  if (!currentTrack) {
    return null;
  }

  const hasNext = queueIndex < queue.length - 1 || repeatMode === "all";
  const hasPrevious = queueIndex > 0 || repeatMode === "all" || currentTime > 3;

  return (
    <div
      className={cn(
        "fixed bottom-0 left-0 right-0 z-50",
        "h-20 border-t border-gray-200/50 dark:border-gray-800/50",
        "bg-white/80 backdrop-blur-lg dark:bg-black/80",
        "animate-in slide-in-from-bottom-full duration-300"
      )}
    >
      {/* Progress bar at top (thin version) */}
      <div className="absolute left-0 right-0 top-0 h-1">
        <div
          className="h-full bg-gradient-to-r from-purple-500 to-pink-500 transition-all duration-150"
          style={{
            width: `${duration > 0 ? (currentTime / duration) * 100 : 0}%`,
          }}
        />
      </div>

      <div className="mx-auto flex h-full max-w-7xl items-center gap-4 px-4 md:px-6">
        {/* Left Section: Cover + Track Info */}
        <div
          className="flex flex-1 cursor-pointer items-center gap-3 overflow-hidden"
          onClick={() => controls.setShowFullPlayer(true)}
        >
          {/* Cover Thumbnail */}
          <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded-lg shadow-md">
            {currentTrack.cover ? (
              <img
                src={formatMediaUrl(currentTrack.cover)}
                alt={currentTrack.name}
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
                <Music className="h-6 w-6 text-white opacity-70" />
              </div>
            )}
            {/* Playing indicator */}
            {isPlaying && (
              <div className="absolute inset-0 flex items-center justify-center bg-black/20">
                <div className="flex gap-0.5">
                  <span
                    className="h-3 w-0.5 animate-pulse rounded-full bg-white"
                    style={{ animationDelay: "0ms" }}
                  />
                  <span
                    className="h-4 w-0.5 animate-pulse rounded-full bg-white"
                    style={{ animationDelay: "150ms" }}
                  />
                  <span
                    className="h-2 w-0.5 animate-pulse rounded-full bg-white"
                    style={{ animationDelay: "300ms" }}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Track Info */}
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-semibold text-gray-900 dark:text-white">
              {currentTrack.name}
            </p>
            <p className="truncate text-xs text-gray-600 dark:text-gray-400">
              {currentTrack.singer}
            </p>
          </div>
        </div>

        {/* Center Section: Controls + Progress (hidden on mobile) */}
        <div className="hidden flex-1 flex-col items-center gap-1 md:flex">
          <PlaybackControls
            isPlaying={isPlaying}
            isLoading={isLoading}
            repeatMode={repeatMode}
            isShuffled={isShuffled}
            hasNext={hasNext}
            hasPrevious={hasPrevious}
            onPlay={controls.play}
            onPause={controls.pause}
            onNext={controls.nextTrack}
            onPrevious={controls.previousTrack}
            onToggleShuffle={controls.toggleShuffle}
            onCycleRepeat={controls.cycleRepeatMode}
            size="sm"
          />
          <ProgressBar
            currentTime={currentTime}
            duration={duration}
            onSeek={controls.seek}
            size="sm"
            showTime={true}
            className="w-full max-w-md"
          />
        </div>

        {/* Mobile Controls (visible only on mobile) */}
        <div className="flex items-center gap-2 md:hidden">
          <button
            onClick={isPlaying ? controls.pause : controls.play}
            className="rounded-full bg-gradient-to-r from-purple-600 to-pink-600 p-2 text-white shadow-md"
          >
            {isLoading ? (
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent" />
            ) : isPlaying ? (
              <svg className="h-5 w-5 fill-current" viewBox="0 0 24 24">
                <rect x="6" y="4" width="4" height="16" />
                <rect x="14" y="4" width="4" height="16" />
              </svg>
            ) : (
              <svg className="h-5 w-5 fill-current" viewBox="0 0 24 24">
                <polygon points="5,3 19,12 5,21" />
              </svg>
            )}
          </button>
        </div>

        {/* Right Section: Queue + Expand */}
        <div className="flex items-center gap-1">
          {/* Queue Button */}
          <button
            onClick={() => controls.setShowQueue(!controls.showQueue)}
            className={cn(
              "rounded-full p-2 transition-colors",
              controls.showQueue
                ? "bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400"
                : "text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
            )}
            title={i18nText.queue}
          >
            <ListMusic className="h-5 w-5" />
          </button>

          {/* Expand Button */}
          <button
            onClick={() => controls.setShowFullPlayer(true)}
            className="rounded-full p-2 text-gray-500 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
            title={i18nText.expand}
          >
            <ChevronUp className="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>
  );
}

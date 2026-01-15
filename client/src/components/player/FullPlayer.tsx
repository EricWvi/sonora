import { useAudioState, useAudioControls } from "@/lib/AudioContext";
import { formatMediaUrl, cn } from "@/lib/utils";
import ProgressBar from "./ProgressBar";
import PlaybackControls from "./PlaybackControls";
import Queue from "./Queue";
import { Music, ChevronDown, ListMusic, X } from "lucide-react";
import { useEffect, useRef } from "react";

const i18nText = {
  nowPlaying: "Now Playing",
  queue: "Queue",
  close: "Close",
  upNext: "Up Next",
};

export default function FullPlayer() {
  const state = useAudioState();
  const controls = useAudioControls();
  const containerRef = useRef<HTMLDivElement>(null);

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

  const { showFullPlayer, showQueue, setShowFullPlayer, setShowQueue } =
    controls;

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && showFullPlayer) {
        setShowFullPlayer(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [showFullPlayer, setShowFullPlayer]);

  // Handle swipe down to close
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let startY = 0;
    let currentY = 0;

    const handleTouchStart = (e: TouchEvent) => {
      startY = e.touches[0].clientY;
      currentY = startY;
    };

    const handleTouchMove = (e: TouchEvent) => {
      currentY = e.touches[0].clientY;
      const diff = currentY - startY;
      if (diff > 0) {
        container.style.transform = `translateY(${diff}px)`;
      }
    };

    const handleTouchEnd = () => {
      const diff = currentY - startY;
      if (diff > 100) {
        setShowFullPlayer(false);
      }
      container.style.transform = "";
    };

    container.addEventListener("touchstart", handleTouchStart);
    container.addEventListener("touchmove", handleTouchMove);
    container.addEventListener("touchend", handleTouchEnd);

    return () => {
      container.removeEventListener("touchstart", handleTouchStart);
      container.removeEventListener("touchmove", handleTouchMove);
      container.removeEventListener("touchend", handleTouchEnd);
    };
  }, [setShowFullPlayer]);

  if (!showFullPlayer || !currentTrack) {
    return null;
  }

  const hasNext = queueIndex < queue.length - 1 || repeatMode === "all";
  const hasPrevious = queueIndex > 0 || repeatMode === "all" || currentTime > 3;

  return (
    <div
      ref={containerRef}
      className={cn(
        "fixed inset-0 z-[100] overflow-hidden",
        "bg-gradient-to-b from-gray-900 via-gray-900 to-black",
        "animate-in slide-in-from-bottom-full duration-300"
      )}
    >
      {/* Background Album Art */}
      {currentTrack.cover && (
        <div
          className="absolute inset-0 bg-cover bg-center opacity-20 blur-3xl"
          style={{
            backgroundImage: `url(${formatMediaUrl(currentTrack.cover)})`,
          }}
        />
      )}

      {/* Content Container */}
      <div className="relative flex h-full flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4">
          <button
            onClick={() => setShowFullPlayer(false)}
            className="rounded-full p-2 text-white/70 transition-colors hover:bg-white/10 hover:text-white"
            aria-label={i18nText.close}
          >
            <ChevronDown className="h-6 w-6" />
          </button>

          <p className="text-sm font-medium uppercase tracking-wider text-white/60">
            {i18nText.nowPlaying}
          </p>

          <button
            onClick={() => setShowQueue(!showQueue)}
            className={cn(
              "rounded-full p-2 transition-colors",
              showQueue
                ? "bg-white/20 text-white"
                : "text-white/70 hover:bg-white/10 hover:text-white"
            )}
            aria-label={i18nText.queue}
          >
            <ListMusic className="h-6 w-6" />
          </button>
        </div>

        {/* Main Content */}
        <div className="flex flex-1 flex-col items-center justify-center px-8 pb-8">
          {showQueue ? (
            <div className="w-full max-w-md flex-1 overflow-hidden rounded-2xl bg-white/10 backdrop-blur-sm">
              <div className="flex items-center justify-between border-b border-white/10 p-4">
                <h2 className="text-lg font-bold text-white">
                  {i18nText.upNext}
                </h2>
                <button
                  onClick={() => setShowQueue(false)}
                  className="rounded-full p-1 text-white/70 hover:bg-white/10 hover:text-white"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>
              <Queue />
            </div>
          ) : (
            <>
              {/* Album Art */}
              <div className="w-full max-w-sm">
                <div className="aspect-square overflow-hidden rounded-2xl shadow-2xl">
                  {currentTrack.cover ? (
                    <img
                      src={formatMediaUrl(currentTrack.cover)}
                      alt={currentTrack.name}
                      className={cn(
                        "h-full w-full object-cover transition-transform duration-300",
                        isPlaying && "scale-[1.02]"
                      )}
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
                      <Music className="h-32 w-32 text-white opacity-50" />
                    </div>
                  )}
                </div>
              </div>

              {/* Track Info */}
              <div className="mt-8 w-full max-w-sm text-center">
                <h1 className="truncate text-2xl font-bold text-white">
                  {currentTrack.name}
                </h1>
                <p className="mt-2 truncate text-lg text-white/70">
                  {currentTrack.singer}
                </p>
                {currentTrack.albumText && (
                  <p className="mt-1 truncate text-sm text-white/50">
                    {currentTrack.albumText}
                  </p>
                )}
              </div>
            </>
          )}
        </div>

        {/* Bottom Controls */}
        <div className="space-y-4 px-8 pb-8 pt-4">
          {/* Progress Bar */}
          <ProgressBar
            currentTime={currentTime}
            duration={duration}
            onSeek={controls.seek}
            size="lg"
            showTime={true}
            className="[&_span]:text-white/70"
          />

          {/* Playback Controls */}
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
            size="lg"
            className="[&_button:not(:disabled)]:text-white/70 [&_button:not(:disabled)]:hover:text-white [&_button.text-purple-600]:!text-purple-400"
          />
        </div>
      </div>
    </div>
  );
}

import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Shuffle,
  Repeat,
  Repeat1,
  Loader2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { RepeatMode } from "@/lib/AudioContext";

interface PlaybackControlsProps {
  isPlaying: boolean;
  isLoading?: boolean;
  repeatMode: RepeatMode;
  isShuffled: boolean;
  hasNext: boolean;
  hasPrevious: boolean;
  onPlay: () => void;
  onPause: () => void;
  onNext: () => void;
  onPrevious: () => void;
  onToggleShuffle: () => void;
  onCycleRepeat: () => void;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const i18nText = {
  play: "Play",
  pause: "Pause",
  previous: "Previous",
  next: "Next",
  shuffle: "Shuffle",
  shuffleOn: "Shuffle on",
  repeatOff: "Repeat off",
  repeatAll: "Repeat all",
  repeatOne: "Repeat one",
};

export default function PlaybackControls({
  isPlaying,
  isLoading = false,
  repeatMode,
  isShuffled,
  hasNext,
  hasPrevious,
  onPlay,
  onPause,
  onNext,
  onPrevious,
  onToggleShuffle,
  onCycleRepeat,
  size = "md",
  className,
}: PlaybackControlsProps) {
  const sizeClasses = {
    sm: {
      button: "p-1.5",
      icon: "h-4 w-4",
      playButton: "p-2.5",
      playIcon: "h-5 w-5",
    },
    md: {
      button: "p-2",
      icon: "h-5 w-5",
      playButton: "p-3",
      playIcon: "h-6 w-6",
    },
    lg: {
      button: "p-3",
      icon: "h-6 w-6",
      playButton: "p-4",
      playIcon: "h-8 w-8",
    },
  }[size];

  const RepeatIcon = repeatMode === "one" ? Repeat1 : Repeat;

  return (
    <div className={cn("flex items-center justify-center gap-2", className)}>
      {/* Shuffle Button */}
      <button
        onClick={onToggleShuffle}
        className={cn(
          "rounded-full transition-all duration-200 hover:bg-gray-100 dark:hover:bg-gray-800",
          sizeClasses.button,
          isShuffled
            ? "text-purple-600 dark:text-purple-400"
            : "text-gray-500 dark:text-gray-400"
        )}
        title={isShuffled ? i18nText.shuffleOn : i18nText.shuffle}
        aria-label={isShuffled ? i18nText.shuffleOn : i18nText.shuffle}
      >
        <Shuffle className={sizeClasses.icon} />
        {isShuffled && (
          <span className="absolute -bottom-0.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-purple-600 dark:bg-purple-400" />
        )}
      </button>

      {/* Previous Button */}
      <button
        onClick={onPrevious}
        disabled={!hasPrevious}
        className={cn(
          "rounded-full transition-all duration-200",
          sizeClasses.button,
          hasPrevious
            ? "text-gray-700 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white"
            : "cursor-not-allowed text-gray-300 dark:text-gray-600"
        )}
        title={i18nText.previous}
        aria-label={i18nText.previous}
      >
        <SkipBack className={cn(sizeClasses.icon, "fill-current")} />
      </button>

      {/* Play/Pause Button */}
      <button
        onClick={isPlaying ? onPause : onPlay}
        disabled={isLoading}
        className={cn(
          "relative rounded-full bg-gradient-to-r from-purple-600 to-pink-600 text-white shadow-lg transition-all duration-200 hover:scale-105 hover:shadow-xl active:scale-95",
          sizeClasses.playButton,
          isLoading && "cursor-wait"
        )}
        title={isPlaying ? i18nText.pause : i18nText.play}
        aria-label={isPlaying ? i18nText.pause : i18nText.play}
      >
        {isLoading ? (
          <Loader2 className={cn(sizeClasses.playIcon, "animate-spin")} />
        ) : isPlaying ? (
          <Pause className={cn(sizeClasses.playIcon, "fill-current")} />
        ) : (
          <Play className={cn(sizeClasses.playIcon, "fill-current ml-0.5")} />
        )}
      </button>

      {/* Next Button */}
      <button
        onClick={onNext}
        disabled={!hasNext}
        className={cn(
          "rounded-full transition-all duration-200",
          sizeClasses.button,
          hasNext
            ? "text-gray-700 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white"
            : "cursor-not-allowed text-gray-300 dark:text-gray-600"
        )}
        title={i18nText.next}
        aria-label={i18nText.next}
      >
        <SkipForward className={cn(sizeClasses.icon, "fill-current")} />
      </button>

      {/* Repeat Button */}
      <button
        onClick={onCycleRepeat}
        className={cn(
          "relative rounded-full transition-all duration-200 hover:bg-gray-100 dark:hover:bg-gray-800",
          sizeClasses.button,
          repeatMode !== "off"
            ? "text-purple-600 dark:text-purple-400"
            : "text-gray-500 dark:text-gray-400"
        )}
        title={
          repeatMode === "off"
            ? i18nText.repeatOff
            : repeatMode === "all"
              ? i18nText.repeatAll
              : i18nText.repeatOne
        }
        aria-label={
          repeatMode === "off"
            ? i18nText.repeatOff
            : repeatMode === "all"
              ? i18nText.repeatAll
              : i18nText.repeatOne
        }
      >
        <RepeatIcon className={sizeClasses.icon} />
        {repeatMode !== "off" && (
          <span className="absolute -bottom-0.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-purple-600 dark:bg-purple-400" />
        )}
      </button>
    </div>
  );
}

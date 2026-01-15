import { useRef, useState, useCallback } from "react";
import { cn } from "@/lib/utils";

interface ProgressBarProps {
  currentTime: number;
  duration: number;
  buffered?: number;
  onSeek: (time: number) => void;
  className?: string;
  showTime?: boolean;
  size?: "sm" | "md" | "lg";
}

function formatTime(seconds: number): string {
  if (!isFinite(seconds) || seconds < 0) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

export default function ProgressBar({
  currentTime,
  duration,
  buffered = 0,
  onSeek,
  className,
  showTime = true,
  size = "md",
}: ProgressBarProps) {
  const progressRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [hoverPosition, setHoverPosition] = useState<number | null>(null);

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;
  const bufferedProgress = duration > 0 ? (buffered / duration) * 100 : 0;

  const calculateTime = useCallback(
    (clientX: number): number => {
      if (!progressRef.current || duration <= 0) return 0;
      const rect = progressRef.current.getBoundingClientRect();
      const position = (clientX - rect.left) / rect.width;
      return Math.max(0, Math.min(duration, position * duration));
    },
    [duration]
  );

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      setIsDragging(true);
      const time = calculateTime(e.clientX);
      onSeek(time);
    },
    [calculateTime, onSeek]
  );

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!progressRef.current) return;
      const rect = progressRef.current.getBoundingClientRect();
      const position = (e.clientX - rect.left) / rect.width;
      const hoverTime = Math.max(0, Math.min(duration, position * duration));
      setHoverPosition(hoverTime);

      if (isDragging) {
        onSeek(hoverTime);
      }
    },
    [isDragging, duration, onSeek]
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  const handleMouseLeave = useCallback(() => {
    setHoverPosition(null);
    if (isDragging) {
      setIsDragging(false);
    }
  }, [isDragging]);

  const handleTouchStart = useCallback(
    (e: React.TouchEvent) => {
      setIsDragging(true);
      const touch = e.touches[0];
      const time = calculateTime(touch.clientX);
      onSeek(time);
    },
    [calculateTime, onSeek]
  );

  const handleTouchMove = useCallback(
    (e: React.TouchEvent) => {
      if (!isDragging) return;
      const touch = e.touches[0];
      const time = calculateTime(touch.clientX);
      onSeek(time);
    },
    [isDragging, calculateTime, onSeek]
  );

  const handleTouchEnd = useCallback(() => {
    setIsDragging(false);
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const step = e.shiftKey ? 10 : 5;
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        onSeek(Math.max(0, currentTime - step));
      } else if (e.key === "ArrowRight") {
        e.preventDefault();
        onSeek(Math.min(duration, currentTime + step));
      }
    },
    [currentTime, duration, onSeek]
  );

  const heightClass = {
    sm: "h-1",
    md: "h-1.5",
    lg: "h-2",
  }[size];

  const thumbSize = {
    sm: "h-3 w-3",
    md: "h-4 w-4",
    lg: "h-5 w-5",
  }[size];

  return (
    <div className={cn("flex items-center gap-3", className)}>
      {/* Current Time */}
      {showTime && (
        <span className="min-w-[40px] text-xs font-medium tabular-nums text-gray-500 dark:text-gray-400">
          {formatTime(currentTime)}
        </span>
      )}

      {/* Progress Bar */}
      <div
        ref={progressRef}
        role="slider"
        tabIndex={0}
        aria-label="Seek slider"
        aria-valuemin={0}
        aria-valuemax={duration}
        aria-valuenow={currentTime}
        className={cn(
          "group relative flex-1 cursor-pointer rounded-full bg-gray-200 dark:bg-gray-700",
          heightClass
        )}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseLeave}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onKeyDown={handleKeyDown}
      >
        {/* Buffered Progress */}
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-gray-300 dark:bg-gray-600"
          style={{ width: `${bufferedProgress}%` }}
        />

        {/* Current Progress */}
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 transition-all duration-75"
          style={{ width: `${progress}%` }}
        />

        {/* Thumb */}
        <div
          className={cn(
            "absolute top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white shadow-lg ring-2 ring-purple-500 transition-transform duration-150",
            thumbSize,
            "opacity-0 scale-0 group-hover:opacity-100 group-hover:scale-100",
            isDragging && "opacity-100 scale-100"
          )}
          style={{ left: `${progress}%` }}
        />

        {/* Hover Time Tooltip */}
        {hoverPosition !== null && (
          <div
            className="pointer-events-none absolute -top-8 -translate-x-1/2 rounded bg-gray-900 px-2 py-1 text-xs font-medium text-white shadow-lg dark:bg-gray-700"
            style={{
              left: `${(hoverPosition / duration) * 100}%`,
            }}
          >
            {formatTime(hoverPosition)}
          </div>
        )}
      </div>

      {/* Duration */}
      {showTime && (
        <span className="min-w-[40px] text-xs font-medium tabular-nums text-gray-500 dark:text-gray-400">
          {formatTime(duration)}
        </span>
      )}
    </div>
  );
}

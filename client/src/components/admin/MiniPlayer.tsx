import { useEffect, useRef } from "react";
import { usePlayer } from "@/hooks/use-miniplayer";
import { formatMediaUrl } from "@/lib/utils";
import { X, Play, Pause } from "lucide-react";

export function MiniPlayer() {
  const {
    currentTrack,
    isPlaying,
    currentTime,
    pause,
    resume,
    setCurrentTime,
    close,
  } = usePlayer();

  const audioRef = useRef<HTMLAudioElement>(null);

  useEffect(() => {
    if (!audioRef.current || !currentTrack) return;

    const audio = audioRef.current;

    // Load new track
    audio.src = formatMediaUrl(currentTrack.url);
    audio.load();

    if (isPlaying) {
      audio.play().catch((e) => {
        console.error("Failed to play audio:", e);
        pause();
      });
    }
  }, [currentTrack]);

  useEffect(() => {
    if (!audioRef.current) return;

    const audio = audioRef.current;

    if (isPlaying) {
      audio.play().catch((e) => {
        console.error("Failed to play audio:", e);
        pause();
      });
    } else {
      audio.pause();
    }
  }, [isPlaying]);

  useEffect(() => {
    if (!audioRef.current) return;

    const audio = audioRef.current;

    const handleTimeUpdate = () => {
      setCurrentTime(audio.currentTime);
    };

    const handleEnded = () => {
      pause();
      setCurrentTime(0);
    };

    audio.addEventListener("timeupdate", handleTimeUpdate);
    audio.addEventListener("ended", handleEnded);

    return () => {
      audio.removeEventListener("timeupdate", handleTimeUpdate);
      audio.removeEventListener("ended", handleEnded);
    };
  }, [currentTrack, pause, setCurrentTime]);

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTime = parseFloat(e.target.value);
    setCurrentTime(newTime);
    if (audioRef.current) {
      audioRef.current.currentTime = newTime;
    }
  };

  const togglePlayPause = () => {
    if (isPlaying) {
      pause();
    } else {
      resume();
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  if (!currentTrack) return null;

  return (
    <>
      <audio ref={audioRef} />
      <div className="fixed bottom-4 right-4 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-lg p-3 w-80 z-50">
        <div className="flex items-start justify-between mb-2">
          <div className="flex-1 min-w-0 mr-2">
            <div className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
              {currentTrack.name}
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400 truncate">
              {currentTrack.singer}
            </div>
          </div>
          <button
            onClick={close}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 flex-shrink-0"
            aria-label="Close player"
          >
            <X size={16} />
          </button>
        </div>

        <div className="flex items-center space-x-2 mb-2">
          <button
            onClick={togglePlayPause}
            className="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white rounded-full transition-colors"
            aria-label={isPlaying ? "Pause" : "Play"}
          >
            {isPlaying ? <Pause size={16} /> : <Play size={16} />}
          </button>

          <div className="flex-1 flex items-center space-x-2">
            <span className="text-xs text-gray-600 dark:text-gray-400 w-10 text-right">
              {formatTime(currentTime)}
            </span>
            <input
              type="range"
              min="0"
              max={currentTrack.duration || 0}
              value={currentTime}
              onChange={handleSeek}
              className="flex-1 h-1 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
              style={{
                background: `linear-gradient(to right, rgb(37, 99, 235) 0%, rgb(37, 99, 235) ${(currentTime / (currentTrack.duration || 1)) * 100}%, rgb(229, 231, 235) ${(currentTime / (currentTrack.duration || 1)) * 100}%, rgb(229, 231, 235) 100%)`,
              }}
            />
            <span className="text-xs text-gray-600 dark:text-gray-400 w-10">
              {formatTime(currentTrack.duration)}
            </span>
          </div>
        </div>
      </div>
    </>
  );
}

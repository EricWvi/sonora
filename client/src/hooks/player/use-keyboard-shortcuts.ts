import { useEffect } from "react";
import { useAudioControls, useAudioState } from "@/lib/AudioContext";

export function useKeyboardShortcuts() {
  const { currentTrack, currentTime, duration } = useAudioState();
  const {
    togglePlay,
    seek,
    nextTrack,
    previousTrack,
    showFullPlayer,
    setShowFullPlayer,
  } = useAudioControls();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't handle if typing in input or textarea
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable
      ) {
        return;
      }

      // Only handle shortcuts if we have a track
      if (!currentTrack) return;

      switch (e.code) {
        case "Space":
          e.preventDefault();
          togglePlay();
          break;

        case "ArrowLeft":
          if (e.shiftKey) {
            e.preventDefault();
            previousTrack();
          } else {
            e.preventDefault();
            // Seek backward 5 seconds
            seek(Math.max(0, currentTime - 5));
          }
          break;

        case "ArrowRight":
          if (e.shiftKey) {
            e.preventDefault();
            nextTrack();
          } else {
            e.preventDefault();
            // Seek forward 5 seconds
            seek(Math.min(duration, currentTime + 5));
          }
          break;

        case "Escape":
          if (showFullPlayer) {
            e.preventDefault();
            setShowFullPlayer(false);
          }
          break;

        default:
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    currentTrack,
    currentTime,
    duration,
    togglePlay,
    seek,
    nextTrack,
    previousTrack,
    showFullPlayer,
    setShowFullPlayer,
  ]);
}

import { useAudioState, useAudioControls } from "@/lib/AudioContext";
import { formatMediaUrl, cn } from "@/lib/utils";
import { Music, Play, X, GripVertical, Trash2 } from "lucide-react";
import { useRef, useState } from "react";

const i18nText = {
  nowPlaying: "Now Playing",
  upNext: "Up Next",
  clearQueue: "Clear queue",
  removeFromQueue: "Remove from queue",
  emptyQueue: "Queue is empty",
};

function formatDuration(seconds: number): string {
  if (!isFinite(seconds) || seconds < 0) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

export default function Queue() {
  const state = useAudioState();
  const controls = useAudioControls();
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const { queue, queueIndex, currentTrack, isPlaying } = state;

  const handleDragStart = (e: React.DragEvent, index: number) => {
    setDraggedIndex(index);
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", index.toString());
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    if (draggedIndex !== null && draggedIndex !== index) {
      setDragOverIndex(index);
    }
  };

  const handleDragLeave = () => {
    setDragOverIndex(null);
  };

  const handleDrop = (e: React.DragEvent, toIndex: number) => {
    e.preventDefault();
    const fromIndex = parseInt(e.dataTransfer.getData("text/plain"), 10);
    if (!isNaN(fromIndex) && fromIndex !== toIndex) {
      controls.reorderQueue(fromIndex, toIndex);
    }
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  const handleDragEnd = () => {
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  if (queue.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <p className="text-sm text-white/50">{i18nText.emptyQueue}</p>
      </div>
    );
  }

  // Separate current track and upcoming tracks
  const upcomingTracks = queue.slice(queueIndex + 1);

  return (
    <div ref={listRef} className="flex h-full flex-col overflow-hidden">
      {/* Current Track */}
      {currentTrack && (
        <div className="border-b border-white/10 p-4">
          <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-white/50">
            {i18nText.nowPlaying}
          </p>
          <div className="flex items-center gap-3">
            <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded-lg">
              {currentTrack.cover ? (
                <img
                  src={formatMediaUrl(currentTrack.cover)}
                  alt={currentTrack.name}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
                  <Music className="h-5 w-5 text-white opacity-70" />
                </div>
              )}
              {isPlaying && (
                <div className="absolute inset-0 flex items-center justify-center bg-black/30">
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
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-semibold text-white">
                {currentTrack.name}
              </p>
              <p className="truncate text-xs text-white/60">
                {currentTrack.singer}
              </p>
            </div>
            <span className="text-xs text-white/50">
              {formatDuration(currentTrack.duration)}
            </span>
          </div>
        </div>
      )}

      {/* Upcoming Tracks */}
      {upcomingTracks.length > 0 && (
        <div className="flex-1 overflow-y-auto">
          <div className="flex items-center justify-between px-4 py-2">
            <p className="text-xs font-semibold uppercase tracking-wider text-white/50">
              {i18nText.upNext}
            </p>
            <button
              onClick={controls.clearQueue}
              className="flex items-center gap-1 rounded px-2 py-1 text-xs text-white/50 transition-colors hover:bg-white/10 hover:text-white"
            >
              <Trash2 className="h-3 w-3" />
              {i18nText.clearQueue}
            </button>
          </div>

          <div className="space-y-0.5 px-2 pb-4">
            {upcomingTracks.map((track, relativeIndex) => {
              const absoluteIndex = queueIndex + 1 + relativeIndex;
              const isDragging = draggedIndex === absoluteIndex;
              const isDragOver = dragOverIndex === absoluteIndex;

              return (
                <div
                  key={`${track.id}-${absoluteIndex}`}
                  draggable
                  onDragStart={(e) => handleDragStart(e, absoluteIndex)}
                  onDragOver={(e) => handleDragOver(e, absoluteIndex)}
                  onDragLeave={handleDragLeave}
                  onDrop={(e) => handleDrop(e, absoluteIndex)}
                  onDragEnd={handleDragEnd}
                  className={cn(
                    "group flex cursor-grab items-center gap-2 rounded-lg px-2 py-2 transition-all",
                    isDragging && "opacity-50",
                    isDragOver && "border-t-2 border-purple-500",
                    "hover:bg-white/10"
                  )}
                >
                  {/* Drag Handle */}
                  <div className="flex-shrink-0 text-white/30 group-hover:text-white/50">
                    <GripVertical className="h-4 w-4" />
                  </div>

                  {/* Cover */}
                  <div
                    onClick={() => controls.jumpToQueueIndex(absoluteIndex)}
                    className="relative h-10 w-10 flex-shrink-0 cursor-pointer overflow-hidden rounded"
                  >
                    {track.cover ? (
                      <img
                        src={formatMediaUrl(track.cover)}
                        alt={track.name}
                        className="h-full w-full object-cover"
                      />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500">
                        <Music className="h-4 w-4 text-white opacity-70" />
                      </div>
                    )}
                    <div className="absolute inset-0 flex items-center justify-center bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
                      <Play className="h-4 w-4 fill-white text-white" />
                    </div>
                  </div>

                  {/* Track Info */}
                  <div
                    onClick={() => controls.jumpToQueueIndex(absoluteIndex)}
                    className="min-w-0 flex-1 cursor-pointer"
                  >
                    <p className="truncate text-sm text-white">{track.name}</p>
                    <p className="truncate text-xs text-white/50">
                      {track.singer}
                    </p>
                  </div>

                  {/* Duration */}
                  <span className="text-xs text-white/40">
                    {formatDuration(track.duration)}
                  </span>

                  {/* Remove Button */}
                  <button
                    onClick={() => controls.removeFromQueue(absoluteIndex)}
                    className="rounded p-1 text-white/30 opacity-0 transition-all hover:bg-white/10 hover:text-white group-hover:opacity-100"
                    title={i18nText.removeFromQueue}
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

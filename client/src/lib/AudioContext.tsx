import {
  createContext,
  useContext,
  useState,
  useCallback,
  useRef,
  useEffect,
  type ReactNode,
} from "react";
import { Howl } from "howler";
import { formatMediaUrl } from "./utils";
import type { Track } from "./localCache";

export type RepeatMode = "off" | "all" | "one";

export interface AudioState {
  // Current track
  currentTrack: Track | null;
  // Playback state
  isPlaying: boolean;
  isLoading: boolean;
  duration: number;
  currentTime: number;
  buffered: number;
  // Queue state
  queue: Track[];
  queueIndex: number;
  // Repeat and shuffle
  repeatMode: RepeatMode;
  isShuffled: boolean;
  // Original queue order (for unshuffle)
  originalQueue: Track[];
  // Error state
  error: string | null;
}

export interface AudioControls {
  // Playback controls
  play: () => void;
  pause: () => void;
  togglePlay: () => void;
  seek: (time: number) => void;
  // Track navigation
  playTrack: (track: Track) => void;
  playTracks: (tracks: Track[], startIndex?: number) => void;
  nextTrack: () => void;
  previousTrack: () => void;
  // Queue management
  addToQueue: (track: Track) => void;
  removeFromQueue: (index: number) => void;
  clearQueue: () => void;
  reorderQueue: (fromIndex: number, toIndex: number) => void;
  jumpToQueueIndex: (index: number) => void;
  // Shuffle and repeat
  toggleShuffle: () => void;
  cycleRepeatMode: () => void;
  setRepeatMode: (mode: RepeatMode) => void;
  // Player UI
  showFullPlayer: boolean;
  setShowFullPlayer: (show: boolean) => void;
  showQueue: boolean;
  setShowQueue: (show: boolean) => void;
}

interface AudioContextType {
  state: AudioState;
  controls: AudioControls;
}

const AudioContext = createContext<AudioContextType | null>(null);

const initialState: AudioState = {
  currentTrack: null,
  isPlaying: false,
  isLoading: false,
  duration: 0,
  currentTime: 0,
  buffered: 0,
  queue: [],
  queueIndex: -1,
  repeatMode: "off",
  isShuffled: false,
  originalQueue: [],
  error: null,
};

interface AudioProviderProps {
  children: ReactNode;
}

export function AudioProvider({ children }: AudioProviderProps) {
  const [state, setState] = useState<AudioState>(initialState);
  const [showFullPlayer, setShowFullPlayer] = useState(false);
  const [showQueue, setShowQueue] = useState(false);

  const howlRef = useRef<Howl | null>(null);
  const progressIntervalRef = useRef<number | null>(null);

  // Cleanup function
  const cleanup = useCallback(() => {
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }
    if (howlRef.current) {
      howlRef.current.unload();
      howlRef.current = null;
    }
  }, []);

  // Update progress
  const startProgressTracking = useCallback(() => {
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
    }
    progressIntervalRef.current = window.setInterval(() => {
      if (howlRef.current && howlRef.current.playing()) {
        const currentTime = howlRef.current.seek() as number;
        setState((prev) => ({ ...prev, currentTime }));
      }
    }, 250);
  }, []);

  const stopProgressTracking = useCallback(() => {
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }
  }, []);

  // Load and play a track
  const loadTrack = useCallback(
    (track: Track, autoplay: boolean = true) => {
      cleanup();

      setState((prev) => ({
        ...prev,
        currentTrack: track,
        isLoading: true,
        isPlaying: false,
        currentTime: 0,
        duration: track.duration || 0,
        error: null,
      }));

      const url = formatMediaUrl(track.url);

      howlRef.current = new Howl({
        src: [url],
        html5: true,
        preload: true,
        onload: () => {
          const duration = howlRef.current?.duration() || track.duration || 0;
          setState((prev) => ({
            ...prev,
            isLoading: false,
            duration,
          }));
        },
        onplay: () => {
          setState((prev) => ({ ...prev, isPlaying: true, isLoading: false }));
          startProgressTracking();
          updateMediaSession(track);
        },
        onpause: () => {
          setState((prev) => ({ ...prev, isPlaying: false }));
          stopProgressTracking();
        },
        onstop: () => {
          setState((prev) => ({ ...prev, isPlaying: false, currentTime: 0 }));
          stopProgressTracking();
        },
        onend: () => {
          setState((prev) => ({ ...prev, isPlaying: false }));
          stopProgressTracking();
          handleTrackEnd();
        },
        onloaderror: (_id, error) => {
          console.error("Audio load error:", error);
          setState((prev) => ({
            ...prev,
            isLoading: false,
            error: "Failed to load audio",
          }));
        },
        onplayerror: (_id, error) => {
          console.error("Audio play error:", error);
          setState((prev) => ({
            ...prev,
            isPlaying: false,
            error: "Failed to play audio",
          }));
        },
      });

      if (autoplay) {
        howlRef.current.play();
      }
    },
    [cleanup, startProgressTracking, stopProgressTracking]
  );

  // Handle track end
  const handleTrackEnd = useCallback(() => {
    setState((prev) => {
      const { queue, queueIndex, repeatMode } = prev;

      // Repeat one: replay current track
      if (repeatMode === "one") {
        setTimeout(() => {
          if (howlRef.current) {
            howlRef.current.seek(0);
            howlRef.current.play();
          }
        }, 0);
        return prev;
      }

      // Check if there's a next track
      const nextIndex = queueIndex + 1;
      if (nextIndex < queue.length) {
        // Play next track
        setTimeout(() => {
          loadTrack(queue[nextIndex], true);
        }, 0);
        return { ...prev, queueIndex: nextIndex };
      }

      // Repeat all: go back to start
      if (repeatMode === "all" && queue.length > 0) {
        setTimeout(() => {
          loadTrack(queue[0], true);
        }, 0);
        return { ...prev, queueIndex: 0 };
      }

      // End of queue
      return { ...prev, isPlaying: false };
    });
  }, [loadTrack]);

  // Update Media Session API
  const updateMediaSession = useCallback((track: Track) => {
    if ("mediaSession" in navigator) {
      navigator.mediaSession.metadata = new MediaMetadata({
        title: track.name,
        artist: track.singer,
        album: track.albumText || "",
        artwork: track.cover
          ? [
              {
                src: formatMediaUrl(track.cover),
                sizes: "512x512",
                type: "image/jpeg",
              },
            ]
          : [],
      });

      navigator.mediaSession.setActionHandler("play", () => {
        howlRef.current?.play();
      });
      navigator.mediaSession.setActionHandler("pause", () => {
        howlRef.current?.pause();
      });
      navigator.mediaSession.setActionHandler("previoustrack", () => {
        controls.previousTrack();
      });
      navigator.mediaSession.setActionHandler("nexttrack", () => {
        controls.nextTrack();
      });
      navigator.mediaSession.setActionHandler("seekto", (details) => {
        if (details.seekTime !== undefined && howlRef.current) {
          howlRef.current.seek(details.seekTime);
          setState((prev) => ({ ...prev, currentTime: details.seekTime! }));
        }
      });
    }
  }, []);

  // Shuffle array helper
  const shuffleArray = <T,>(array: T[]): T[] => {
    const shuffled = [...array];
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }
    return shuffled;
  };

  // Controls
  const controls: AudioControls = {
    play: useCallback(() => {
      if (howlRef.current) {
        howlRef.current.play();
      }
    }, []),

    pause: useCallback(() => {
      if (howlRef.current) {
        howlRef.current.pause();
      }
    }, []),

    togglePlay: useCallback(() => {
      if (howlRef.current) {
        if (howlRef.current.playing()) {
          howlRef.current.pause();
        } else {
          howlRef.current.play();
        }
      }
    }, []),

    seek: useCallback((time: number) => {
      if (howlRef.current) {
        howlRef.current.seek(time);
        setState((prev) => ({ ...prev, currentTime: time }));
      }
    }, []),

    playTrack: useCallback(
      (track: Track) => {
        setState((prev) => ({
          ...prev,
          queue: [track],
          originalQueue: [track],
          queueIndex: 0,
        }));
        loadTrack(track, true);
      },
      [loadTrack]
    ),

    playTracks: useCallback(
      (tracks: Track[], startIndex: number = 0) => {
        if (tracks.length === 0) return;

        const queue = state.isShuffled ? shuffleArray(tracks) : tracks;
        const actualIndex = state.isShuffled
          ? queue.findIndex((t) => t.id === tracks[startIndex]?.id)
          : startIndex;

        setState((prev) => ({
          ...prev,
          queue,
          originalQueue: tracks,
          queueIndex: actualIndex >= 0 ? actualIndex : 0,
        }));

        loadTrack(queue[actualIndex >= 0 ? actualIndex : 0], true);
      },
      [loadTrack, state.isShuffled]
    ),

    nextTrack: useCallback(() => {
      setState((prev) => {
        const { queue, queueIndex, repeatMode } = prev;
        let nextIndex = queueIndex + 1;

        if (nextIndex >= queue.length) {
          if (repeatMode === "all") {
            nextIndex = 0;
          } else {
            return prev; // No next track
          }
        }

        setTimeout(() => {
          loadTrack(queue[nextIndex], true);
        }, 0);

        return { ...prev, queueIndex: nextIndex };
      });
    }, [loadTrack]),

    previousTrack: useCallback(() => {
      // If more than 3 seconds in, restart current track
      if (state.currentTime > 3 && howlRef.current) {
        howlRef.current.seek(0);
        setState((prev) => ({ ...prev, currentTime: 0 }));
        return;
      }

      setState((prev) => {
        const { queue, queueIndex, repeatMode } = prev;
        let prevIndex = queueIndex - 1;

        if (prevIndex < 0) {
          if (repeatMode === "all") {
            prevIndex = queue.length - 1;
          } else {
            // Restart current track
            if (howlRef.current) {
              howlRef.current.seek(0);
            }
            return { ...prev, currentTime: 0 };
          }
        }

        setTimeout(() => {
          loadTrack(queue[prevIndex], true);
        }, 0);

        return { ...prev, queueIndex: prevIndex };
      });
    }, [loadTrack, state.currentTime]),

    addToQueue: useCallback((track: Track) => {
      setState((prev) => ({
        ...prev,
        queue: [...prev.queue, track],
        originalQueue: [...prev.originalQueue, track],
      }));
    }, []),

    removeFromQueue: useCallback(
      (index: number) => {
        setState((prev) => {
          const newQueue = [...prev.queue];
          newQueue.splice(index, 1);

          const newOriginalQueue = [...prev.originalQueue];
          const trackToRemove = prev.queue[index];
          const origIndex = newOriginalQueue.findIndex(
            (t) => t.id === trackToRemove?.id
          );
          if (origIndex !== -1) {
            newOriginalQueue.splice(origIndex, 1);
          }

          let newQueueIndex = prev.queueIndex;
          if (index < prev.queueIndex) {
            newQueueIndex--;
          } else if (index === prev.queueIndex) {
            // Current track removed, stop playback
            cleanup();
            return {
              ...prev,
              queue: newQueue,
              originalQueue: newOriginalQueue,
              queueIndex: -1,
              currentTrack: null,
              isPlaying: false,
            };
          }

          return {
            ...prev,
            queue: newQueue,
            originalQueue: newOriginalQueue,
            queueIndex: newQueueIndex,
          };
        });
      },
      [cleanup]
    ),

    clearQueue: useCallback(() => {
      cleanup();
      setState((prev) => ({
        ...prev,
        queue: [],
        originalQueue: [],
        queueIndex: -1,
        currentTrack: null,
        isPlaying: false,
        currentTime: 0,
      }));
    }, [cleanup]),

    reorderQueue: useCallback((fromIndex: number, toIndex: number) => {
      setState((prev) => {
        const newQueue = [...prev.queue];
        const [removed] = newQueue.splice(fromIndex, 1);
        newQueue.splice(toIndex, 0, removed);

        let newQueueIndex = prev.queueIndex;
        if (fromIndex === prev.queueIndex) {
          newQueueIndex = toIndex;
        } else if (fromIndex < prev.queueIndex && toIndex >= prev.queueIndex) {
          newQueueIndex--;
        } else if (fromIndex > prev.queueIndex && toIndex <= prev.queueIndex) {
          newQueueIndex++;
        }

        return { ...prev, queue: newQueue, queueIndex: newQueueIndex };
      });
    }, []),

    jumpToQueueIndex: useCallback(
      (index: number) => {
        setState((prev) => {
          if (index >= 0 && index < prev.queue.length) {
            setTimeout(() => {
              loadTrack(prev.queue[index], true);
            }, 0);
            return { ...prev, queueIndex: index };
          }
          return prev;
        });
      },
      [loadTrack]
    ),

    toggleShuffle: useCallback(() => {
      setState((prev) => {
        if (prev.isShuffled) {
          // Unshuffle: restore original order
          const currentTrack = prev.currentTrack;
          const newIndex = currentTrack
            ? prev.originalQueue.findIndex((t) => t.id === currentTrack.id)
            : 0;
          return {
            ...prev,
            isShuffled: false,
            queue: [...prev.originalQueue],
            queueIndex: newIndex >= 0 ? newIndex : 0,
          };
        } else {
          // Shuffle: randomize order but keep current track first
          const currentTrack = prev.currentTrack;
          const otherTracks = prev.queue.filter(
            (t) => t.id !== currentTrack?.id
          );
          const shuffled = shuffleArray(otherTracks);
          const newQueue = currentTrack
            ? [currentTrack, ...shuffled]
            : shuffled;
          return {
            ...prev,
            isShuffled: true,
            queue: newQueue,
            queueIndex: 0,
          };
        }
      });
    }, []),

    cycleRepeatMode: useCallback(() => {
      setState((prev) => {
        const modes: RepeatMode[] = ["off", "all", "one"];
        const currentIndex = modes.indexOf(prev.repeatMode);
        const nextIndex = (currentIndex + 1) % modes.length;
        return { ...prev, repeatMode: modes[nextIndex] };
      });
    }, []),

    setRepeatMode: useCallback((mode: RepeatMode) => {
      setState((prev) => ({ ...prev, repeatMode: mode }));
    }, []),

    showFullPlayer,
    setShowFullPlayer,
    showQueue,
    setShowQueue,
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cleanup();
    };
  }, [cleanup]);

  return (
    <AudioContext.Provider value={{ state, controls }}>
      {children}
    </AudioContext.Provider>
  );
}

export function useAudio() {
  const context = useContext(AudioContext);
  if (!context) {
    throw new Error("useAudio must be used within an AudioProvider");
  }
  return context;
}

export function useAudioState() {
  const { state } = useAudio();
  return state;
}

export function useAudioControls() {
  const { controls } = useAudio();
  return controls;
}

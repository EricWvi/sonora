import { create } from "zustand";

interface Track {
  id: number;
  name: string;
  singer: string;
  url: string;
  cover: string;
  duration: number;
}

interface PlayerState {
  currentTrack: Track | null;
  isPlaying: boolean;
  currentTime: number;
  play: (track: Track) => void;
  pause: () => void;
  resume: () => void;
  setCurrentTime: (time: number) => void;
  close: () => void;
}

export const usePlayer = create<PlayerState>((set) => ({
  currentTrack: null,
  isPlaying: false,
  currentTime: 0,
  play: (track) =>
    set({
      currentTrack: track,
      isPlaying: true,
      currentTime: 0,
    }),
  pause: () => set({ isPlaying: false }),
  resume: () => set({ isPlaying: true }),
  setCurrentTime: (time) => set({ currentTime: time }),
  close: () =>
    set({
      currentTrack: null,
      isPlaying: false,
      currentTime: 0,
    }),
}));

import { useEffect, useState, type ReactNode } from "react";
import { syncManager } from "@/lib/localCache";
import { useAudioState } from "@/lib/AudioContext";
import Sidebar from "./Sidebar";
import MiniPlayer from "./MiniPlayer";
import FullPlayer from "./FullPlayer";
import { AppleMusic } from "./icons";

interface LayoutProps {
  children: ReactNode;
  currentView?: string;
  onNavigate?: (path: string) => void;
  searchQuery?: string;
  onSearchChange?: (value: string) => void;
  onSearchClear?: () => void;
}

const i18nText = {
  loading: "Loading your music library...",
  initialSync: "Syncing for the first time, this may take a moment...",
};

export default function Layout({
  children,
  currentView,
  onNavigate,
  searchQuery = "",
  onSearchChange,
  onSearchClear,
}: LayoutProps) {
  const [isInitializing, setIsInitializing] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { currentTrack } = useAudioState();

  useEffect(() => {
    const initialize = async () => {
      const startTime = Date.now();
      const MIN_LOADING_TIME = 1000; // 1 second minimum

      try {
        await syncManager.initialize();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to initialize");
      } finally {
        // Ensure loading scene shows for at least 1 second
        const elapsed = Date.now() - startTime;
        const remainingTime = Math.max(0, MIN_LOADING_TIME - elapsed);
        
        setTimeout(() => {
          setIsInitializing(false);
        }, remainingTime);
      }
    };

    initialize();
  }, []);

  if (isInitializing) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-[#282828]">
        <div className="size-24 sm:size-36">
          <AppleMusic />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-[#1E1E1E]">
        <div className="max-w-md rounded-2xl border border-red-200 bg-white p-8 shadow-2xl dark:border-red-900/30 dark:bg-[#2A2A2A]">
          <h2 className="mb-3 text-2xl font-bold text-red-600 dark:text-red-400">
            Initialization Error
          </h2>
          <p className="text-gray-700 dark:text-gray-300">{error}</p>
        </div>
      </div>
    );
  }

  // Determine if mini player is visible (when there's a current track)
  const hasMiniPlayer = !!currentTrack;

  return (
    <div className="flex h-screen overflow-hidden bg-white dark:bg-[#1E1E1E]">
      {/* Sidebar */}
      <Sidebar 
        currentView={currentView} 
        onNavigate={onNavigate}
        searchQuery={searchQuery}
        onSearchChange={onSearchChange}
        onSearchClear={onSearchClear}
      />

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        <div
          className={`mx-auto max-w-7xl px-6 py-8 md:px-8 lg:px-12 ${hasMiniPlayer ? "pb-28" : ""}`}
        >
          {children}
        </div>
      </main>

      {/* Mini Player */}
      <MiniPlayer />

      {/* Full Player */}
      <FullPlayer />
    </div>
  );
}

import { useEffect, useState, type ReactNode } from "react";
import { syncManager } from "@/lib/localCache";
import Sidebar from "./Sidebar";
import { Loader2 } from "lucide-react";

interface LayoutProps {
  children: ReactNode;
  currentView?: string;
  onNavigate?: (path: string) => void;
}

const i18nText = {
  loading: "Loading your music library...",
  initialSync: "Syncing for the first time, this may take a moment...",
};

export default function Layout({
  children,
  currentView,
  onNavigate,
}: LayoutProps) {
  const [isInitializing, setIsInitializing] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initialize = async () => {
      try {
        await syncManager.initialize();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to initialize");
      } finally {
        setIsInitializing(false);
      }
    };

    initialize();
  }, []);

  if (isInitializing) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-purple-500 via-pink-500 to-blue-500 dark:from-purple-900 dark:via-pink-900 dark:to-blue-900">
        <div className="flex flex-col items-center gap-6">
          <Loader2 className="h-16 w-16 animate-spin text-white drop-shadow-lg" />
          <p className="text-xl font-semibold text-white drop-shadow-md">
            {i18nText.loading}
          </p>
          <p className="text-sm text-white/80">{i18nText.initialSync}</p>
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

  return (
    <div className="flex h-screen overflow-hidden bg-white dark:bg-[#1E1E1E]">
      {/* Sidebar */}
      <Sidebar currentView={currentView} onNavigate={onNavigate} />

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-7xl px-6 py-8 md:px-8 lg:px-12">
          {children}
        </div>
      </main>
    </div>
  );
}

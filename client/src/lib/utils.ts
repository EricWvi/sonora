import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Detect if running in Tauri environment
export const isTauri = () => {
  return typeof window !== "undefined" && "__TAURI__" in window;
};

/**
 * Format media URL based on runtime environment
 * @param mediaId - UUID of the media file
 * @returns Formatted URL for accessing the media
 */
export function formatMediaUrl(mediaId: string): string {
  if (mediaId.startsWith("/api/m/")) {
    return mediaId;
  }
  if (isTauri()) {
    // Tauri: Use file system path
    // TODO: Implement Tauri-specific path resolution when Tauri support is added
    return `tauri://media/${mediaId}`;
  }

  // Web: Use server endpoint
  return `/api/m/${mediaId}`;
}

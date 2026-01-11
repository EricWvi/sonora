import { useQuery } from "@tanstack/react-query";
import { dbClient } from "@/lib/localCache";
import { queryClient } from "@/lib/queryClient";

const keyAllTracks = () => ["player", "tracks"];
const keyTrack = (id: number) => ["player", "track", id];
const keyAlbumTracks = (albumId: number) => [
  "player",
  "tracks",
  "album",
  albumId,
];
const keySingles = () => ["player", "tracks", "singles"];
const keySearchTracks = (query: string) => [
  "player",
  "tracks",
  "search",
  query,
];
const keyLyric = (id: number) => ["player", "lyric", id];

export function useTracks() {
  return useQuery({
    queryKey: keyAllTracks(),
    queryFn: async () => {
      const tracks = await dbClient.getAllTracks();
      // Preload individual track queries
      const trackIds = tracks.map((track) => {
        queryClient.setQueryData(keyTrack(track.id), track);
        return track.id;
      });
      return trackIds;
    },
    staleTime: Infinity, // Data is managed by sync
    gcTime: Infinity,
  });
}

export function useTrack(id: number) {
  return useQuery({
    queryKey: keyTrack(id),
    queryFn: async () => {
      return await dbClient.getTrack(id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export function useAlbumTracks(albumId: number) {
  return useQuery({
    queryKey: keyAlbumTracks(albumId),
    queryFn: async () => {
      return (await dbClient.getTracksByAlbum(albumId)).map(
        (track) => track.id
      );
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export function useSingles() {
  return useQuery({
    queryKey: keySingles(),
    queryFn: async () => {
      return (await dbClient.getSingles()).map((track) => track.id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export function useSearchTracks(query: string) {
  return useQuery({
    queryKey: keySearchTracks(query),
    queryFn: async () => {
      if (!query.trim()) return [];
      return (await dbClient.searchTracks(query)).map((track) => track.id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: query.trim().length > 0,
  });
}

export function useLyric(id: number) {
  return useQuery({
    queryKey: keyLyric(id),
    queryFn: async () => {
      return await dbClient.getLyric(id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: id > 0,
  });
}

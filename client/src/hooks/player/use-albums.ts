import { useQuery } from "@tanstack/react-query";
import { dbClient } from "@/lib/localCache";
import { queryClient } from "@/lib/queryClient";

const keyAllAlbums = () => ["player", "albums"];
const keyAlbum = (id: number) => ["player", "album", id];
const keySearchAlbums = (query: string) => [
  "player",
  "albums",
  "search",
  query,
];

export function useAlbums() {
  return useQuery({
    queryKey: keyAllAlbums(),
    queryFn: async () => {
      const albums = await dbClient.getAllAlbums();
      // Preload individual album queries
      const albumIds = albums.map((album) => {
        queryClient.setQueryData(keyAlbum(album.id), album);
        return album.id;
      });
      return albumIds;
    },
    staleTime: Infinity, // Data is managed by sync
    gcTime: Infinity,
  });
}

export function useAlbum(id: number) {
  return useQuery({
    queryKey: keyAlbum(id),
    queryFn: async () => {
      return await dbClient.getAlbum(id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export function useSearchAlbums(query: string) {
  return useQuery({
    queryKey: keySearchAlbums(query),
    queryFn: async () => {
      if (!query.trim()) return [];
      return (await dbClient.searchAlbums(query)).map((album) => album.id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: query.trim().length > 0,
  });
}

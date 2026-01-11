import { useQuery } from "@tanstack/react-query";
import { dbClient } from "@/lib/localCache";
import { queryClient } from "@/lib/queryClient";

const keyAllSingers = () => ["player", "singers"];
const keySinger = (id: number) => ["player", "singer", id];
const keySearchSingers = (query: string) => [
  "player",
  "singers",
  "search",
  query,
];

export function useSingers() {
  return useQuery({
    queryKey: keyAllSingers(),
    queryFn: async () => {
      const singers = await dbClient.getAllSingers();
      // Preload individual singer queries
      const singerIds = singers.map((singer) => {
        queryClient.setQueryData(keySinger(singer.id), singer);
        return singer.id;
      });
      return singerIds;
    },
    staleTime: Infinity, // Data is managed by sync
    gcTime: Infinity,
  });
}

export function useSinger(id: number) {
  return useQuery({
    queryKey: keySinger(id),
    queryFn: async () => {
      return await dbClient.getSinger(id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export function useSearchSingers(query: string) {
  return useQuery({
    queryKey: keySearchSingers(query),
    queryFn: async () => {
      if (!query.trim()) return [];
      return (await dbClient.searchSingers(query)).map((singer) => singer.id);
    },
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: query.trim().length > 0,
  });
}

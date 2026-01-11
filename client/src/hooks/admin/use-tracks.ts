import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getRequest, postRequest, queryClient } from "@/lib/queryClient";

export type Track = {
  id: number;
  name: string;
  singer: string;
  album: number;
  cover: string;
  url: string;
  lyric: number;
  duration: number;
  year: number;
  trackNumber: number;
  genre: string;
  albumText: string;
};

export interface QueryCondition {
  operator: string;
  value: any;
}

const keyTrack = (id: number) => ["/api/track", id];
const keyAlbumTracks = (album: number) => ["/api/track/album", album];

export function useTrack(id: number) {
  return useQuery({
    queryKey: keyTrack(id),
    queryFn: async () => {
      const data = await getRequest(`/api/track?Action=GetTrack&id=${id}`);
      return data.track as Track;
    },
  });
}

export function useTracks(album: number, enabled: boolean = true) {
  return useQuery({
    queryKey: keyAlbumTracks(album),
    queryFn: async () => {
      if (album === 0) {
        const data = await getRequest("/api/track?Action=ListSingles");
        return data.tracks as Track[];
      } else {
        const data = await getRequest(
          "/api/track?Action=ListAlbumTracks&albumId=" + album
        );
        return data.tracks as Track[];
      }
    },
    enabled,
  });
}

export async function listTracks(
  page: number = 1,
  condition: QueryCondition[] = []
): Promise<[Track[], boolean]> {
  const data = await getRequest(
    `/api/track?Action=ListTracks&page=${page}&condition=${JSON.stringify(condition)}`
  );
  (data.tracks as Track[]).map((track) => {
    queryClient.setQueryData(keyTrack(track.id), track);
  });
  return [data.tracks as Track[], data.hasMore];
}

export function useCreateTrack() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<Track, "id">) => {
      const response = await postRequest("/api/track?Action=CreateTrack", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAlbumTracks(variables.album),
      });
    },
  });
}

export function useUpdateTrack() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number } & Partial<Track>) => {
      const response = await postRequest("/api/track?Action=UpdateTrack", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAlbumTracks(variables.album || 0),
      });
    },
  });
}

export function useDeleteTrack() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number; album?: number }) => {
      const response = await postRequest("/api/track?Action=DeleteTrack", {
        id: data.id,
      });
      return response;
    },
    onSuccess: async (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAlbumTracks(variables.album || 0),
      });
    },
  });
}

// Async lyric functions (no react-query)
export async function getLyric(id: number): Promise<string> {
  const response = await getRequest(`/api/track?Action=GetLyric&id=${id}`);
  return response.lyric as string;
}

export async function createLyric(content: string): Promise<{ id: number }> {
  const response = await postRequest("/api/track?Action=CreateLyric", {
    content,
  });
  return response;
}

export async function updateLyric(id: number, content: string): Promise<void> {
  await postRequest("/api/track?Action=UpdateLyric", {
    id,
    content,
  });
}

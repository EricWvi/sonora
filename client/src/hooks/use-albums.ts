import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getRequest, postRequest, queryClient } from "@/lib/queryClient";

export type Album = {
  id: number;
  name: string;
  cover: string;
  year: number;
};

export interface QueryCondition {
  operator: string;
  value: any;
}

const keyAlbum = (id: number) => ["/api/album", id];
const keyAllAlbums = () => ["/api/album/all"];

export function useAlbum(id: number) {
  return useQuery({
    queryKey: keyAlbum(id),
    queryFn: async () => {
      const data = await getRequest(`/api/album?Action=GetAlbum&id=${id}`);
      return data.album as Album;
    },
  });
}

export function useAlbums() {
  return useQuery({
    queryKey: keyAllAlbums(),
    queryFn: async () => {
      const data = await getRequest("/api/album?Action=ListAllAlbums");
      return data.albums as Album[];
    },
  });
}

export async function listAlbums(
  page: number = 1,
  condition: QueryCondition[] = []
): Promise<[Album[], boolean]> {
  const data = await getRequest(
    `/api/album?Action=ListAlbums&page=${page}&condition=${JSON.stringify(condition)}`
  );
  (data.albums as Album[]).map((album) => {
    queryClient.setQueryData(keyAlbum(album.id), album);
  });
  return [data.albums as Album[], data.hasMore];
}

export function useCreateAlbum() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<Album, "id">) => {
      const response = await postRequest("/api/album?Action=CreateAlbum", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllAlbums(),
      });
    },
  });
}

export function useUpdateAlbum() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number } & Partial<Album>) => {
      const response = await postRequest("/api/album?Action=UpdateAlbum", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllAlbums(),
      });
    },
  });
}

export function useDeleteAlbum() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number }) => {
      const response = await postRequest("/api/album?Action=DeleteAlbum", {
        id: data.id,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllAlbums(),
      });
    },
  });
}
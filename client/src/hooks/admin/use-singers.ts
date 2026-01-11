import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getRequest, postRequest, queryClient } from "@/lib/queryClient";

export type Singer = {
  id: number;
  name: string;
  avatar: string;
};

export interface QueryCondition {
  operator: string;
  value: any;
}

const keySinger = (id: number) => ["/api/singer", id];
const keyAllSingers = () => ["/api/singer/all"];

export function useSinger(id: number) {
  return useQuery({
    queryKey: keySinger(id),
    queryFn: async () => {
      const data = await getRequest(`/api/singer?Action=GetSinger&id=${id}`);
      return data.singer as Singer;
    },
  });
}

export function useSingers() {
  return useQuery({
    queryKey: keyAllSingers(),
    queryFn: async () => {
      const data = await getRequest("/api/singer?Action=ListAllSingers");
      return data.singers as Singer[];
    },
  });
}

export async function listSingers(
  page: number = 1,
  condition: QueryCondition[] = []
): Promise<[Singer[], boolean]> {
  const data = await getRequest(
    `/api/singer?Action=ListSingers&page=${page}&condition=${JSON.stringify(condition)}`
  );
  (data.singers as Singer[]).map((singer) => {
    queryClient.setQueryData(keySinger(singer.id), singer);
  });
  return [data.singers as Singer[], data.hasMore];
}

export function useCreateSinger() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<Singer, "id">) => {
      const response = await postRequest("/api/singer?Action=CreateSinger", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllSingers(),
      });
    },
  });
}

export function useUpdateSinger() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number } & Partial<Singer>) => {
      const response = await postRequest("/api/singer?Action=UpdateSinger", {
        ...data,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllSingers(),
      });
    },
  });
}

export function useDeleteSinger() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { id: number }) => {
      const response = await postRequest("/api/singer?Action=DeleteSinger", {
        id: data.id,
      });
      return response;
    },
    onSuccess: async (_data, _variables) => {
      queryClient.invalidateQueries({
        queryKey: keyAllSingers(),
      });
    },
  });
}

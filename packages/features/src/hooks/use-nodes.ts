import type { ContractsClient } from "@sonora/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createContext, useContext } from "react";
import { toast } from "sonner";

export const ClientContext = createContext<ContractsClient | null>(null);

function useClient(): ContractsClient {
  const client = useContext(ClientContext);
  if (!client) throw new Error("ClientContext not provided");
  return client;
}

function toastError(err: unknown) {
  const status = (err as { status?: number }).status;
  const message =
    status === 409
      ? "A node with that name already exists"
      : status === 404
        ? "Node not found"
        : (err as Error).message ?? "Something went wrong";
  toast.error(message);
}

export function useRootChildren() {
  const client = useClient();
  return useQuery({
    queryKey: ["nodes", "root"],
    queryFn: () => client.listRootChildren({}),
  });
}

export function useChildren(id: string | null) {
  const client = useClient();
  return useQuery({
    queryKey: ["nodes", id, "children"],
    queryFn: () => {
      if (id === null) return client.listRootChildren({});
      return client.listChildren({ id });
    },
  });
}

export function useNodePath(id: string | null) {
  const client = useClient();
  return useQuery({
    queryKey: ["nodes", id, "path"],
    queryFn: () => client.getNodePath({ id: id! }),
    enabled: id !== null,
  });
}

function invalidateChildren(queryClient: ReturnType<typeof useQueryClient>, parentId: string | null) {
  if (parentId === null) {
    void queryClient.invalidateQueries({ queryKey: ["nodes", "root"] });
  } else {
    void queryClient.invalidateQueries({ queryKey: ["nodes", parentId, "children"] });
  }
}

export function useCreateDir(parentId: string | null) {
  const client = useClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) =>
      client.createDir({ parentId, name }),
    onSuccess: () => invalidateChildren(qc, parentId),
    onError: toastError,
  });
}

export function useCreateFile(parentId: string | null) {
  const client = useClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { name: string; size: bigint | null; mimeType: string | null }) =>
      client.createFile({ parentId, ...args }),
    onSuccess: () => invalidateChildren(qc, parentId),
    onError: toastError,
  });
}

export function useRenameNode(parentId: string | null) {
  const client = useClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; newName: string; newParentId: string | null }) =>
      client.moveNode({ id: args.id, newName: args.newName, newParentId: args.newParentId }),
    onSuccess: () => {
      invalidateChildren(qc, parentId);
      void qc.invalidateQueries({ queryKey: ["nodes", "root"] });
    },
    onError: toastError,
  });
}

export function useDeleteNode(parentId: string | null) {
  const client = useClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => client.deleteNode({ id }),
    onSuccess: () => invalidateChildren(qc, parentId),
    onError: toastError,
  });
}

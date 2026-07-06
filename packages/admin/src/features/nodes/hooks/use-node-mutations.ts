import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useClient } from "../../../providers/client-context";
import { notifyError } from "../../../lib/error-toast";
import { FILE_PLACEHOLDER_SIZE } from "../constants";

// Invalidate the whole `nodes` query namespace after any structural mutation so
// the listing and breadcrumb path refresh consistently.
function useInvalidateNodes() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: ["nodes"] });
}

export function useCreateDir(parentId: string | null) {
  const client = useClient();
  const invalidate = useInvalidateNodes();
  return useMutation({
    mutationFn: async (name: string) =>
      client.createDir({ parentId, name }),
    onSuccess: invalidate,
    onError: notifyError,
  });
}

export interface CreateFileInput {
  name: string;
  mimeType: string | null;
}

export function useCreateFile(parentId: string | null) {
  const client = useClient();
  const invalidate = useInvalidateNodes();
  return useMutation({
    // `size` is hardcoded to the 4 MB placeholder; see FILE_PLACEHOLDER_SIZE.
    mutationFn: async ({ name, mimeType }: CreateFileInput) =>
      client.createFile({
        parentId,
        name,
        size: FILE_PLACEHOLDER_SIZE,
        mimeType,
      }),
    onSuccess: invalidate,
    onError: notifyError,
  });
}

export interface RenameInput {
  id: string;
  newName: string;
}

// Inline rename: moveNode with the existing parentId so the node stays in place
// while only its name changes.
export function useRenameNode(parentId: string | null) {
  const client = useClient();
  const invalidate = useInvalidateNodes();
  return useMutation({
    mutationFn: async ({ id, newName }: RenameInput) =>
      client.moveNode({ id, newParentId: parentId, newName }),
    onSuccess: invalidate,
    onError: notifyError,
  });
}

export function useDeleteNode() {
  const client = useClient();
  const invalidate = useInvalidateNodes();
  return useMutation({
    mutationFn: async (id: string) => client.deleteNode({ id }),
    onSuccess: invalidate,
    onError: notifyError,
  });
}

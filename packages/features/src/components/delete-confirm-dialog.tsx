import { AlertDialog } from "@sonora/ui";
import { useDeleteNode } from "../hooks/use-nodes.js";

interface DeleteConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeId: string;
  nodeName: string;
  parentId: string | null;
}

export function DeleteConfirmDialog({
  open,
  onOpenChange,
  nodeId,
  nodeName,
  parentId,
}: DeleteConfirmDialogProps) {
  const deleteNode = useDeleteNode(parentId);

  function handleConfirm() {
    deleteNode.mutate(nodeId, {
      onSuccess: () => onOpenChange(false),
    });
  }

  return (
    <AlertDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Delete node"
      description={`"${nodeName}" will be permanently deleted. This action cannot be undone.`}
      confirmLabel={deleteNode.isPending ? "Deleting…" : "Delete"}
      onConfirm={handleConfirm}
    />
  );
}

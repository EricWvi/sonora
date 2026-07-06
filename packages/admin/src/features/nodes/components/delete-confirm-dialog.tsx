import { Button, Modal } from "@heroui/react";
import type { NodeView } from "@sonora/contracts";
import { useDeleteNode } from "../hooks/use-node-mutations";

interface DeleteConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  node: NodeView | undefined;
}

export function DeleteConfirmDialog({ isOpen, onClose, node }: DeleteConfirmDialogProps) {
  const remove = useDeleteNode();

  function handleConfirm() {
    if (!node) {
      return;
    }
    remove.mutate(node.id, { onSuccess: onClose });
  }

  return (
    <Modal isOpen={isOpen} onOpenChange={(open) => { if (!open) onClose(); }}>
      <Modal.Backdrop variant="blur">
        <Modal.Container size="sm">
          <Modal.Dialog aria-label="Delete node">
            <Modal.Header>
              <Modal.Heading>Delete {node?.kind === "directory" ? "directory" : "file"}</Modal.Heading>
            </Modal.Header>
            <Modal.Body>
              <p>
                Are you sure you want to delete <span className="font-medium">{node?.name}</span>?
                This cannot be undone.
              </p>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="ghost" onPress={onClose} isDisabled={remove.isPending}>
                Cancel
              </Button>
              <Button
                variant="danger"
                onPress={handleConfirm}
                isPending={remove.isPending}
              >
                Delete
              </Button>
            </Modal.Footer>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal>
  );
}

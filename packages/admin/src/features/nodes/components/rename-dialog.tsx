import { useEffect, useState } from "react";
import {
  Button,
  FieldError,
  Input,
  Label,
  Modal,
  TextField,
} from "@heroui/react";
import type { NodeView } from "@sonora/contracts";
import { useRenameNode } from "../hooks/use-node-mutations";

interface RenameDialogProps {
  isOpen: boolean;
  onClose: () => void;
  // The rename target. Present whenever the dialog is open.
  node: NodeView | undefined;
  parentId: string | null;
}

export function RenameDialog({ isOpen, onClose, node, parentId }: RenameDialogProps) {
  const [name, setName] = useState("");
  const rename = useRenameNode(parentId);

  useEffect(() => {
    if (isOpen && node) {
      setName(node.name);
    }
  }, [isOpen, node]);

  const trimmed = name.trim();
  const isInvalid = name.length > 0 && trimmed.length === 0;

  function handleSubmit() {
    if (!node || trimmed.length === 0 || trimmed === node.name) {
      return;
    }
    rename.mutate({ id: node.id, newName: trimmed }, { onSuccess: onClose });
  }

  return (
    <Modal isOpen={isOpen} onOpenChange={(open) => { if (!open) onClose(); }}>
      <Modal.Backdrop variant="blur">
        <Modal.Container size="sm">
          <Modal.Dialog aria-label="Rename node">
            <Modal.Header>
              <Modal.Heading>Rename</Modal.Heading>
            </Modal.Header>
            <Modal.Body>
              <TextField
                value={name}
                onChange={setName}
                isInvalid={isInvalid}
                isRequired
              >
                <Label>New name</Label>
                <Input autoFocus />
                <FieldError>{isInvalid ? "Name is required" : null}</FieldError>
              </TextField>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="ghost" onPress={onClose} isDisabled={rename.isPending}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onPress={handleSubmit}
                isDisabled={trimmed.length === 0 || trimmed === node?.name}
                isPending={rename.isPending}
              >
                Save
              </Button>
            </Modal.Footer>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal>
  );
}

import { useEffect, useState } from "react";
import {
  Button,
  FieldError,
  Input,
  Label,
  Modal,
  TextField,
} from "@heroui/react";
import { useCreateDir } from "../hooks/use-node-mutations";

interface CreateDirDialogProps {
  isOpen: boolean;
  onClose: () => void;
  parentId: string | null;
}

export function CreateDirDialog({ isOpen, onClose, parentId }: CreateDirDialogProps) {
  const [name, setName] = useState("");
  const createDir = useCreateDir(parentId);

  // Reset the field each time the dialog opens so stale input never lingers.
  useEffect(() => {
    if (isOpen) {
      setName("");
    }
  }, [isOpen]);

  const trimmed = name.trim();
  const isInvalid = name.length > 0 && trimmed.length === 0;

  function handleSubmit() {
    if (trimmed.length === 0) {
      return;
    }
    createDir.mutate(trimmed, { onSuccess: onClose });
  }

  return (
    <Modal isOpen={isOpen} onOpenChange={(open) => { if (!open) onClose(); }}>
      <Modal.Backdrop variant="blur">
        <Modal.Container size="sm">
          <Modal.Dialog aria-label="Create directory">
            <Modal.Header>
              <Modal.Heading>Create directory</Modal.Heading>
            </Modal.Header>
            <Modal.Body>
              <TextField
                value={name}
                onChange={setName}
                isInvalid={isInvalid}
                isRequired
              >
                <Label>Directory name</Label>
                <Input placeholder="new-folder" autoFocus />
                <FieldError>{isInvalid ? "Name is required" : null}</FieldError>
              </TextField>
            </Modal.Body>
            <Modal.Footer>
              <Button
                variant="ghost"
                onPress={onClose}
                isDisabled={createDir.isPending}
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                onPress={handleSubmit}
                isDisabled={trimmed.length === 0}
                isPending={createDir.isPending}
              >
                Create
              </Button>
            </Modal.Footer>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal>
  );
}

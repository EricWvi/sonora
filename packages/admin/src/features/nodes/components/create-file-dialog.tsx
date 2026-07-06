import { useEffect, useState } from "react";
import {
  Button,
  Description,
  FieldError,
  Input,
  Label,
  Modal,
  TextField,
} from "@heroui/react";
import { useCreateFile } from "../hooks/use-node-mutations";

interface CreateFileDialogProps {
  isOpen: boolean;
  onClose: () => void;
  parentId: string | null;
}

export function CreateFileDialog({ isOpen, onClose, parentId }: CreateFileDialogProps) {
  const [name, setName] = useState("");
  const [mimeType, setMimeType] = useState("");
  const createFile = useCreateFile(parentId);

  useEffect(() => {
    if (isOpen) {
      setName("");
      setMimeType("");
    }
  }, [isOpen]);

  const trimmedName = name.trim();
  const isInvalid = name.length > 0 && trimmedName.length === 0;

  function handleSubmit() {
    if (trimmedName.length === 0) {
      return;
    }
    // MIME type is optional and intentionally not format-validated.
    createFile.mutate(
      { name: trimmedName, mimeType: mimeType.trim() || null },
      { onSuccess: onClose },
    );
  }

  return (
    <Modal isOpen={isOpen} onOpenChange={(open) => { if (!open) onClose(); }}>
      <Modal.Backdrop variant="blur">
        <Modal.Container size="sm">
          <Modal.Dialog aria-label="Create file">
            <Modal.Header>
              <Modal.Heading>Create file</Modal.Heading>
            </Modal.Header>
            <Modal.Body>
              <TextField
                value={name}
                onChange={setName}
                isInvalid={isInvalid}
                isRequired
              >
                <Label>File name</Label>
                <Input placeholder="report.pdf" autoFocus />
                <FieldError>{isInvalid ? "Name is required" : null}</FieldError>
              </TextField>
              <TextField value={mimeType} onChange={setMimeType}>
                <Label>MIME type</Label>
                <Input placeholder="application/pdf" />
                <Description>Optional. Not strictly validated.</Description>
              </TextField>
            </Modal.Body>
            <Modal.Footer>
              <Button
                variant="ghost"
                onPress={onClose}
                isDisabled={createFile.isPending}
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                onPress={handleSubmit}
                isDisabled={trimmedName.length === 0}
                isPending={createFile.isPending}
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

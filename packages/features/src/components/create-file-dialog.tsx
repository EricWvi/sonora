import { Button, Dialog, Input, Label } from "@sonora/ui";
import { useState } from "react";
import { useCreateFile } from "../hooks/use-nodes.js";
import { FILE_SIZE } from "../lib/constants.js";

interface CreateFileDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  parentId: string | null;
}

interface FormState {
  name: string;
  mimeType: string;
}

export function CreateFileDialog({ open, onOpenChange, parentId }: CreateFileDialogProps) {
  const [form, setForm] = useState<FormState>({ name: "", mimeType: "" });
  const [nameError, setNameError] = useState("");
  const createFile = useCreateFile(parentId);

  function handleSubmit() {
    if (!form.name.trim()) {
      setNameError("Name is required");
      return;
    }
    const mimeType = form.mimeType.trim() || null;
    createFile.mutate({ name: form.name.trim(), size: BigInt(FILE_SIZE), mimeType }, {
      onSuccess: () => {
        setForm({ name: "", mimeType: "" });
        setNameError("");
        onOpenChange(false);
      },
    });
  }

  function handleOpenChange(next: boolean) {
    if (!next) {
      setForm({ name: "", mimeType: "" });
      setNameError("");
    }
    onOpenChange(next);
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      title="New File"
      description="Creates a file node with metadata only — no content upload."
      footer={
        <>
          <Button variant="ghost" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={createFile.isPending}>
            {createFile.isPending ? "Creating…" : "Create"}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="file-name">File name *</Label>
          <Input
            id="file-name"
            placeholder="document.pdf"
            value={form.name}
            onChange={(e) => {
              setForm((prev) => ({ ...prev, name: e.target.value }));
              setNameError("");
            }}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
            autoFocus
            error={nameError}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="file-mime">MIME type</Label>
          <Input
            id="file-mime"
            placeholder="application/pdf"
            value={form.mimeType}
            onChange={(e) => setForm((prev) => ({ ...prev, mimeType: e.target.value }))}
          />
        </div>
      </div>
    </Dialog>
  );
}

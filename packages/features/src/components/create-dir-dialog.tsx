import { Button, Dialog, Input, Label } from "@sonora/ui";
import { useState } from "react";
import { useCreateDir } from "../hooks/use-nodes.js";

interface CreateDirDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  parentId: string | null;
}

export function CreateDirDialog({ open, onOpenChange, parentId }: CreateDirDialogProps) {
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const createDir = useCreateDir(parentId);

  function handleSubmit() {
    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    createDir.mutate(name.trim(), {
      onSuccess: () => {
        setName("");
        setError("");
        onOpenChange(false);
      },
    });
  }

  function handleOpenChange(next: boolean) {
    if (!next) {
      setName("");
      setError("");
    }
    onOpenChange(next);
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      title="New Folder"
      footer={
        <>
          <Button variant="ghost" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={createDir.isPending}>
            {createDir.isPending ? "Creating…" : "Create"}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="dir-name">Folder name</Label>
        <Input
          id="dir-name"
          placeholder="Untitled Folder"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            setError("");
          }}
          onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
          autoFocus
          error={error}
        />
      </div>
    </Dialog>
  );
}

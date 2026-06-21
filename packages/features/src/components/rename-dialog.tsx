import { Button, Dialog, Input, Label } from "@sonora/ui";
import { useEffect, useState } from "react";
import { useRenameNode } from "../hooks/use-nodes.js";

interface RenameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nodeId: string;
  currentName: string;
  parentId: string | null;
}

export function RenameDialog({
  open,
  onOpenChange,
  nodeId,
  currentName,
  parentId,
}: RenameDialogProps) {
  const [name, setName] = useState(currentName);
  const [error, setError] = useState("");
  const rename = useRenameNode(parentId);

  useEffect(() => {
    if (open) setName(currentName);
  }, [open, currentName]);

  function handleSubmit() {
    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    rename.mutate(
      { id: nodeId, newName: name.trim(), newParentId: parentId },
      {
        onSuccess: () => {
          setError("");
          onOpenChange(false);
        },
      },
    );
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) setError("");
        onOpenChange(next);
      }}
      title="Rename"
      footer={
        <>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={rename.isPending}>
            {rename.isPending ? "Renaming…" : "Rename"}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="rename-input">New name</Label>
        <Input
          id="rename-input"
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

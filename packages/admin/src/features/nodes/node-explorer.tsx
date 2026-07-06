import { useState } from "react";
import { Button } from "@heroui/react";
import { FilePlus, FolderPlus } from "lucide-react";
import type { NodeView } from "@sonora/contracts";
import { DashboardLayout } from "../../dashboard-layout";
import { useNodeChildren } from "./hooks/use-node-children";
import { NodeBreadcrumb, type Crumb } from "./components/node-breadcrumb";
import { NodeTable } from "./components/node-table";
import { CreateDirDialog } from "./components/create-dir-dialog";
import { CreateFileDialog } from "./components/create-file-dialog";
import { RenameDialog } from "./components/rename-dialog";
import { DeleteConfirmDialog } from "./components/delete-confirm-dialog";

type DialogKind = "createDir" | "createFile" | "rename" | "delete" | null;

interface DialogState {
  kind: DialogKind;
  node: NodeView | undefined;
}

// Top-level VFS browser. Holds the directory stack and which dialog (if any) is
// open, and wires the listing + CRUD dialogs together. The create actions are
// hoisted into the dashboard top bar via the layout's actions slot; the
// breadcrumb + node table fill the content area. No router: the URL stays at
// "/" and navigation lives entirely in component state.
export function NodeExplorer() {
  const [crumbs, setCrumbs] = useState<Crumb[]>([]);
  const [dialog, setDialog] = useState<DialogState>({
    kind: null,
    node: undefined,
  });

  const currentId = crumbs.length > 0 ? crumbs[crumbs.length - 1].id : null;
  const { data: nodes, isLoading, isError, refetch } =
    useNodeChildren(currentId);

  function navigate(index: number) {
    // -1 => root; otherwise truncate the stack to that segment.
    setCrumbs(index === -1 ? [] : crumbs.slice(0, index + 1));
  }

  function openDir(node: NodeView) {
    setCrumbs((prev) => [...prev, { id: node.id, name: node.name }]);
  }

  const close = () => setDialog({ kind: null, node: undefined });

  return (
    <DashboardLayout
      title="Files"
      actions={
        <>
          <Button
            variant="ghost"
            onPress={() => setDialog({ kind: "createDir", node: undefined })}
          >
            <FolderPlus /> New folder
          </Button>
          <Button
            variant="primary"
            onPress={() => setDialog({ kind: "createFile", node: undefined })}
          >
            <FilePlus /> New file
          </Button>
        </>
      }
    >
      <div className="mx-auto max-w-6xl space-y-4 p-6">
        <NodeBreadcrumb crumbs={crumbs} onNavigate={navigate} />
        <NodeTable
          nodes={nodes ?? []}
          isLoading={isLoading}
          isError={isError}
          onOpenDir={openDir}
          onRename={(node) => setDialog({ kind: "rename", node })}
          onDelete={(node) => setDialog({ kind: "delete", node })}
          onRetry={() => refetch()}
        />
      </div>

      <CreateDirDialog
        isOpen={dialog.kind === "createDir"}
        onClose={close}
        parentId={currentId}
      />
      <CreateFileDialog
        isOpen={dialog.kind === "createFile"}
        onClose={close}
        parentId={currentId}
      />
      <RenameDialog
        isOpen={dialog.kind === "rename"}
        node={dialog.node}
        onClose={close}
        parentId={currentId}
      />
      <DeleteConfirmDialog
        isOpen={dialog.kind === "delete"}
        node={dialog.node}
        onClose={close}
      />
    </DashboardLayout>
  );
}

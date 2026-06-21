import type { NodeView } from "@sonora/contracts";
import { Badge, ContextMenu, DropdownMenu, IconButton } from "@sonora/ui";
import { AnimatePresence, motion } from "motion/react";
import {
  File,
  Folder,
  FolderPlus,
  FilePlus,
  MoreHorizontal,
  Pencil,
  Trash2,
  ChevronRight,
} from "lucide-react";
import { useState } from "react";
import { useChildren } from "../hooks/use-nodes.js";
import { formatSize, formatTimestamp } from "../lib/format.js";
import { CreateDirDialog } from "./create-dir-dialog.js";
import { CreateFileDialog } from "./create-file-dialog.js";
import { RenameDialog } from "./rename-dialog.js";
import { DeleteConfirmDialog } from "./delete-confirm-dialog.js";

interface NodeListProps {
  parentId: string | null;
  onSelectDir: (id: string) => void;
  onSelectNode: (node: NodeView) => void;
  selectedNodeId: string | null;
  breadcrumbs: Array<{ id: string | null; name: string }>;
  onBreadcrumbClick: (id: string | null) => void;
}

interface ActionTarget {
  node: NodeView;
  action: "rename" | "delete";
}

export function NodeList({
  parentId,
  onSelectDir,
  onSelectNode,
  selectedNodeId,
  breadcrumbs,
  onBreadcrumbClick,
}: NodeListProps) {
  const { data, isLoading } = useChildren(parentId);
  const [createDir, setCreateDir] = useState(false);
  const [createFile, setCreateFile] = useState(false);
  const [actionTarget, setActionTarget] = useState<ActionTarget | null>(null);

  const nodes = data?.nodes ?? [];

  function rowActions(node: NodeView) {
    return [
      {
        label: "Rename",
        icon: <Pencil size={14} />,
        onSelect: () => setActionTarget({ node, action: "rename" }),
      },
      {
        label: "Delete",
        icon: <Trash2 size={14} />,
        onSelect: () => setActionTarget({ node, action: "delete" }),
        destructive: true,
      },
    ];
  }

  return (
    <div className="flex h-full flex-col">
      {/* Toolbar */}
      <div className="flex items-center justify-between border-b border-[--color-neutral-200] px-3 py-2">
        {/* Breadcrumbs */}
        <div className="flex items-center gap-1 overflow-hidden text-[13px]">
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.id ?? "root"} className="flex items-center gap-1">
              {i > 0 && <ChevronRight size={12} className="text-[--color-neutral-300] shrink-0" />}
              <button
                onClick={() => onBreadcrumbClick(crumb.id)}
                className="truncate text-[--color-neutral-500] hover:text-[--color-neutral-800] transition-colors duration-100 cursor-default"
              >
                {crumb.name}
              </button>
            </span>
          ))}
        </div>
        {/* Action buttons */}
        <div className="flex items-center gap-1 shrink-0 ml-2">
          <IconButton label="New Folder" onClick={() => setCreateDir(true)}>
            <FolderPlus size={15} />
          </IconButton>
          <IconButton label="New File" onClick={() => setCreateFile(true)}>
            <FilePlus size={15} />
          </IconButton>
        </div>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex h-32 items-center justify-center text-[13px] text-[--color-neutral-300]">
            Loading…
          </div>
        ) : nodes.length === 0 ? (
          <div className="flex h-32 items-center justify-center text-[13px] text-[--color-neutral-300]">
            This folder is empty
          </div>
        ) : (
          <AnimatePresence initial={false}>
            {nodes.map((node, i) => (
              <motion.div
                key={node.id}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.12, delay: i * 0.02, ease: "easeOut" }}
              >
                <ContextMenu items={rowActions(node)}>
                  <div
                    onClick={() => {
                      onSelectNode(node);
                      if (node.kind === "directory") onSelectDir(node.id);
                    }}
                    className={`
                      group flex items-center gap-3 px-3 py-2 cursor-default select-none
                      hover:bg-[--color-neutral-100] transition-colors duration-75
                      ${selectedNodeId === node.id ? "bg-[--color-neutral-100]" : ""}
                    `}
                  >
                    {/* Icon */}
                    <span className="shrink-0 text-[--color-neutral-400]">
                      {node.kind === "directory" ? (
                        <Folder size={16} className="text-[--color-accent]" />
                      ) : (
                        <File size={16} />
                      )}
                    </span>

                    {/* Name */}
                    <span className="flex-1 min-w-0 truncate text-[13px] text-[--color-neutral-800]">
                      {node.name}
                    </span>

                    {/* Meta */}
                    <span className="hidden sm:flex items-center gap-4 shrink-0 text-[12px] text-[--color-neutral-400]">
                      {node.kind === "file" && (
                        <>
                          <span className="w-20 text-right">{formatSize(node.size)}</span>
                          {node.mimeType && (
                            <span className="w-32 truncate">{node.mimeType}</span>
                          )}
                        </>
                      )}
                      <Badge
                        variant={node.storageStatus === "available" ? "available" : "pending"}
                      >
                        {node.storageStatus === "available" ? "Available" : "Pending"}
                      </Badge>
                      <span className="w-36 text-right">{formatTimestamp(node.updatedAt)}</span>
                    </span>

                    {/* Row action */}
                    <span className="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity duration-100">
                      <DropdownMenu items={rowActions(node)}>
                        <IconButton label="More actions">
                          <MoreHorizontal size={14} />
                        </IconButton>
                      </DropdownMenu>
                    </span>
                  </div>
                </ContextMenu>
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>

      {/* Dialogs */}
      <CreateDirDialog open={createDir} onOpenChange={setCreateDir} parentId={parentId} />
      <CreateFileDialog open={createFile} onOpenChange={setCreateFile} parentId={parentId} />

      {actionTarget?.action === "rename" && (
        <RenameDialog
          open
          onOpenChange={(o) => { if (!o) setActionTarget(null); }}
          nodeId={actionTarget.node.id}
          currentName={actionTarget.node.name}
          parentId={parentId}
        />
      )}
      {actionTarget?.action === "delete" && (
        <DeleteConfirmDialog
          open
          onOpenChange={(o) => { if (!o) setActionTarget(null); }}
          nodeId={actionTarget.node.id}
          nodeName={actionTarget.node.name}
          parentId={parentId}
        />
      )}
    </div>
  );
}

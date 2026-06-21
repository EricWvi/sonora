import type { NodeView } from "@sonora/contracts";
import { AnimatePresence, motion } from "motion/react";
import { ChevronRight, Folder } from "lucide-react";
import { useState } from "react";
import { useChildren, useRootChildren } from "../hooks/use-nodes.js";
import { cn } from "@sonora/ui";

interface NodeTreeProps {
  selectedDirId: string | null;
  onSelectDir: (id: string | null) => void;
}

interface TreeNodeProps {
  node: NodeView;
  depth: number;
  selectedDirId: string | null;
  onSelectDir: (id: string | null) => void;
}

function TreeNode({ node, depth, selectedDirId, onSelectDir }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(false);
  const { data } = useChildren(expanded ? node.id : null);
  const children = (data?.nodes ?? []).filter((n) => n.kind === "directory");

  function handleClick() {
    onSelectDir(node.id);
    if (node.kind === "directory") setExpanded((prev) => !prev);
  }

  return (
    <div>
      <motion.div
        onClick={handleClick}
        whileTap={{ scale: 0.98 }}
        transition={{ duration: 0.08 }}
        className={cn(
          "flex items-center gap-1.5 px-2 py-1.5 rounded-[--radius-md] cursor-default select-none",
          "transition-colors duration-75 text-[13px]",
          selectedDirId === node.id
            ? "bg-[--color-accent]/10 text-[--color-accent]"
            : "text-[--color-neutral-700] hover:bg-[--color-neutral-100]",
        )}
        style={{ paddingLeft: `${8 + depth * 16}px` }}
      >
        <motion.span
          className="shrink-0 text-[--color-neutral-300]"
          animate={{ rotate: expanded ? 90 : 0 }}
          transition={{ duration: 0.12 }}
        >
          <ChevronRight size={12} />
        </motion.span>
        <Folder
          size={14}
          className={cn(
            "shrink-0",
            selectedDirId === node.id ? "text-[--color-accent]" : "text-[--color-neutral-400]",
          )}
        />
        <span className="truncate">{node.name}</span>
      </motion.div>

      <AnimatePresence initial={false}>
        {expanded && children.length > 0 && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.15, ease: "easeOut" }}
            className="overflow-hidden"
          >
            {children.map((child) => (
              <TreeNode
                key={child.id}
                node={child}
                depth={depth + 1}
                selectedDirId={selectedDirId}
                onSelectDir={onSelectDir}
              />
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export function NodeTree({ selectedDirId, onSelectDir }: NodeTreeProps) {
  const { data, isLoading } = useRootChildren();
  const dirs = (data?.nodes ?? []).filter((n) => n.kind === "directory");

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="px-3 py-2 text-[11px] font-medium uppercase tracking-wide text-[--color-neutral-400] border-b border-[--color-neutral-200]">
        Files
      </div>
      <div className="flex-1 overflow-y-auto p-1">
        {/* Root row */}
        <motion.div
          onClick={() => onSelectDir(null)}
          whileTap={{ scale: 0.98 }}
          transition={{ duration: 0.08 }}
          className={cn(
            "flex items-center gap-1.5 px-2 py-1.5 rounded-[--radius-md] cursor-default select-none",
            "transition-colors duration-75 text-[13px] mb-0.5",
            selectedDirId === null
              ? "bg-[--color-accent]/10 text-[--color-accent]"
              : "text-[--color-neutral-700] hover:bg-[--color-neutral-100]",
          )}
        >
          <span className="w-3 shrink-0" />
          <Folder
            size={14}
            className={cn(
              "shrink-0",
              selectedDirId === null ? "text-[--color-accent]" : "text-[--color-neutral-400]",
            )}
          />
          <span className="truncate font-medium">Root</span>
        </motion.div>

        {isLoading ? (
          <div className="px-3 py-2 text-[12px] text-[--color-neutral-300]">Loading…</div>
        ) : (
          dirs.map((node) => (
            <TreeNode
              key={node.id}
              node={node}
              depth={0}
              selectedDirId={selectedDirId}
              onSelectDir={onSelectDir}
            />
          ))
        )}
      </div>
    </div>
  );
}

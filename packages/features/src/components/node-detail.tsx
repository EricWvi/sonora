import type { NodeView } from "@sonora/contracts";
import { Badge } from "@sonora/ui";
import { AnimatePresence, motion } from "motion/react";
import { useNodePath } from "../hooks/use-nodes.js";
import { formatSize, formatTimestamp } from "../lib/format.js";

interface NodeDetailProps {
  node: NodeView | null;
}

interface DetailRowProps {
  label: string;
  value: React.ReactNode;
}

function DetailRow({ label, value }: DetailRowProps) {
  return (
    <div className="flex items-baseline gap-3 py-1.5 border-b border-[--color-neutral-100] last:border-0">
      <span className="w-24 shrink-0 text-[11px] font-medium uppercase tracking-wide text-[--color-neutral-400]">
        {label}
      </span>
      <span className="min-w-0 truncate text-[13px] text-[--color-neutral-700]">{value}</span>
    </div>
  );
}

function NodeDetailContent({ node }: { node: NodeView }) {
  const pathQuery = useNodePath(node.id);

  const statusVariant =
    node.storageStatus === "available" ? "available" : "pending";

  return (
    <div className="flex flex-col">
      <DetailRow label="Name" value={node.name} />
      <DetailRow label="Kind" value={node.kind} />
      <DetailRow
        label="Path"
        value={
          pathQuery.data ? (
            <span className="font-mono text-[12px]">{pathQuery.data.path}</span>
          ) : (
            <span className="text-[--color-neutral-300]">Loading…</span>
          )
        }
      />
      <DetailRow label="ID" value={<span className="font-mono text-[12px]">{node.id}</span>} />
      <DetailRow
        label="Status"
        value={
          <Badge variant={statusVariant}>
            {node.storageStatus === "available" ? "Available" : "Pending upload"}
          </Badge>
        }
      />
      <DetailRow label="MD5" value={node.md5 ?? "—"} />
      <DetailRow label="Size" value={formatSize(node.size)} />
      <DetailRow label="MIME" value={node.mimeType ?? "—"} />
      <DetailRow label="Created" value={formatTimestamp(node.createdAt)} />
      <DetailRow label="Updated" value={formatTimestamp(node.updatedAt)} />
    </div>
  );
}

export function NodeDetail({ node }: NodeDetailProps) {
  return (
    <AnimatePresence mode="wait">
      {node ? (
        <motion.div
          key={node.id}
          className="flex flex-col"
          initial={{ opacity: 0, x: 8 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: 8 }}
          transition={{ duration: 0.15, ease: "easeOut" }}
        >
          <NodeDetailContent node={node} />
        </motion.div>
      ) : (
        <motion.div
          key="empty"
          className="flex h-full items-center justify-center text-[13px] text-[--color-neutral-300]"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.1 }}
        >
          Select a node to view details
        </motion.div>
      )}
    </AnimatePresence>
  );
}

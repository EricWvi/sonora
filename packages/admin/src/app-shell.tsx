import "./styles.css";
import type { ContractsClient } from "@sonora/contracts";
import { ClientContext, NodeDetail, NodeList, NodeTree } from "@sonora/features";
import { QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";
import { Toaster } from "sonner";
import { queryClient } from "./query-client.js";
import type { NodeView } from "@sonora/contracts";

interface AppShellProps {
  client: ContractsClient;
}

export function AppShell({ client }: AppShellProps) {
  const [selectedDirId, setSelectedDirId] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<NodeView | null>(null);
  // Breadcrumb trail: [{id, name}]
  const [trail, setTrail] = useState<Array<{ id: string | null; name: string }>>([
    { id: null, name: "Root" },
  ]);

  function handleSelectDir(id: string | null) {
    setSelectedDirId(id);
    setSelectedNode(null);
  }

  function handleSelectDirFromList(id: string, name?: string) {
    setSelectedDirId(id);
    setSelectedNode(null);
    setTrail((prev) => {
      const idx = prev.findIndex((c) => c.id === id);
      if (idx >= 0) return prev.slice(0, idx + 1);
      return [...prev, { id, name: name ?? id }];
    });
  }

  function handleSelectNode(node: NodeView) {
    setSelectedNode(node);
    if (node.kind === "directory") {
      handleSelectDirFromList(node.id, node.name);
    }
  }

  function handleBreadcrumbClick(id: string | null) {
    setSelectedDirId(id);
    setSelectedNode(null);
    setTrail((prev) => {
      const idx = prev.findIndex((c) => c.id === id);
      return idx >= 0 ? prev.slice(0, idx + 1) : prev;
    });
  }

  function handleTreeSelectDir(id: string | null) {
    handleSelectDir(id);
    setTrail([{ id: null, name: "Root" }]);
    if (id !== null) {
      // Will be updated when the node name is known from the list
      setTrail([{ id: null, name: "Root" }, { id, name: id }]);
    }
  }

  return (
    <QueryClientProvider client={queryClient}>
      <ClientContext.Provider value={client}>
        <div className="flex h-screen overflow-hidden bg-[--color-neutral-50]">
          {/* Left sidebar — directory tree */}
          <aside className="w-52 shrink-0 border-r border-[--color-neutral-200] bg-white overflow-hidden flex flex-col">
            <NodeTree
              selectedDirId={selectedDirId}
              onSelectDir={handleTreeSelectDir}
            />
          </aside>

          {/* Main content */}
          <main className="flex-1 flex flex-col overflow-hidden">
            {/* Node list takes up most of the space */}
            <div className="flex-1 overflow-hidden">
              <NodeList
                parentId={selectedDirId}
                onSelectDir={(id) => handleSelectDirFromList(id)}
                onSelectNode={handleSelectNode}
                selectedNodeId={selectedNode?.id ?? null}
                breadcrumbs={trail}
                onBreadcrumbClick={handleBreadcrumbClick}
              />
            </div>

            {/* Detail panel — bottom strip */}
            {selectedNode && (
              <div className="h-[220px] shrink-0 border-t border-[--color-neutral-200] bg-white overflow-y-auto px-4 py-3">
                <div className="mb-2 text-[11px] font-medium uppercase tracking-wide text-[--color-neutral-400]">
                  Details
                </div>
                <NodeDetail node={selectedNode} />
              </div>
            )}
          </main>
        </div>
        <Toaster position="bottom-right" />
      </ClientContext.Provider>
    </QueryClientProvider>
  );
}

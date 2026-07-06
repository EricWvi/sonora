import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toast } from "@heroui/react";
import type { ContractsClient } from "@sonora/contracts";
import { ClientProvider } from "./providers/client-context";
import { NodeExplorer } from "./features/nodes/node-explorer";

interface AppShellProps {
  client: ContractsClient;
}

// Composes the provider stack for the admin panel:
// - QueryClientProvider: tanstack-query server state for node listings/CRUD.
// - Toast.Provider: imperative `toast.success/danger` used by error handling.
// - ClientProvider: the typed VFS ContractsClient via React Context.
export function AppShell({ client }: AppShellProps) {
  // One QueryClient per shell instance, created once via lazy state.
  const [queryClient] = useState(() => new QueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      <ClientProvider client={client}>
        <NodeExplorer />
      </ClientProvider>
      {/* Toast.Provider hosts the global toast region; it renders toast popups
          rather than its children, so it must be a sibling of the app content
          (not wrap it). `toast.success/danger` write to a module-level queue
          that this region observes. */}
      <Toast.Provider placement="top end" />
    </QueryClientProvider>
  );
}

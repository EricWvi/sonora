import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { ContractsClient } from "@sonora/contracts";

// Provides the type-safe VFS client to the feature tree via React Context so
// hooks can reach it without prop-drilling. The host assembles the client once
// (`createContractsClient(createFetchTransport())`) and passes it to AppShell.
const ClientContext = createContext<ContractsClient | null>(null);

interface ClientProviderProps {
  client: ContractsClient;
  children: ReactNode;
}

export function ClientProvider({ client, children }: ClientProviderProps) {
  return <ClientContext.Provider value={client}>{children}</ClientContext.Provider>;
}

export function useClient(): ContractsClient {
  const client = useContext(ClientContext);
  if (client === null) {
    throw new Error("useClient must be used within a ClientProvider");
  }
  return client;
}

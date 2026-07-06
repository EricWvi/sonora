import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createContractsClient } from "@sonora/contracts";
import { createFetchTransport } from "@sonora/contracts/fetch";
import { AppShell } from "@sonora/admin";
import "./index.css";

// Relative requests ride the Vite dev `/api` proxy in development; in production
// the client is served from the same origin as the backend.
const client = createContractsClient(createFetchTransport());

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppShell client={client} />
  </StrictMode>,
);

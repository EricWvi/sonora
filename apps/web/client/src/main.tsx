import "./styles.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createContractsClient, createFetchTransport } from "@sonora/contracts";
import { AppShell } from "@sonora/admin";

const client = createContractsClient(createFetchTransport());

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppShell client={client} />
  </StrictMode>,
);

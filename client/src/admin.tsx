import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./admin.css";
import AdminBoard from "./AdminBoard.tsx";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/queryClient.ts";
import { Toaster } from "sonner";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AdminBoard />
      <Toaster position="top-right" />
    </QueryClientProvider>
  </StrictMode>
);

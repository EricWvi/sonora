import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwindcss from "@tailwindcss/vite";

// Workspace packages (@sonora/admin, @sonora/contracts) are resolved via their
// conditional `exports` (the `development` condition points at source), so no
// `resolve.alias` is needed for them — Vite activates `development` in dev mode.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  server: {
    host: "0.0.0.0",
    port: 5173,
    fs: {
      // Allow Vite to read workspace sibling sources (packages/*).
      allow: [".."],
    },
    watch: {
      // Keep watching the symlinked @sonora/* workspace packages.
      ignored: ["!**/node_modules/@sonora/**"],
    },
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false,
      },
    },
  },
  build: {
    outDir: "dist",
  },
});

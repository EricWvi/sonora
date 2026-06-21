import path from "path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5175,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@sonora/contracts": path.resolve(__dirname, "../../../packages/contracts/src/index.ts"),
      "@sonora/ui": path.resolve(__dirname, "../../../packages/ui/src/index.ts"),
      "@sonora/features": path.resolve(__dirname, "../../../packages/features/src/index.ts"),
      "@sonora/admin": path.resolve(__dirname, "../../../packages/admin/src/index.ts"),
    },
  },
});

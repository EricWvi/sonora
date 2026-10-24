import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// The console is served by the Go service under /admin/. In development
// the API is proxied to `task run:server`.
export default defineConfig({
  base: "/admin/",
  plugins: [vue()],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  server: {
    host: "127.0.0.1",
    port: 14561,
    strictPort: true,
    proxy: { "/admin/api": "http://127.0.0.1:14560" },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: { ui: ["naive-ui"], vue: ["vue", "vue-router"] },
      },
    },
  },
});

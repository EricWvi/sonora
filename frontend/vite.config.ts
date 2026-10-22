import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { NaiveUiResolver } from "unplugin-vue-components/resolvers";

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ["vue", "vue-router", "@vueuse/core"],
      dts: "auto-imports.d.ts",
    }),
    Components({ resolvers: [NaiveUiResolver()], dts: "components.d.ts" }),
  ],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  server: { host: "127.0.0.1", port: 14559, strictPort: true },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          ui: ["naive-ui"],
          vue: ["vue", "vue-router", "pinia"],
        },
      },
    },
  },
});

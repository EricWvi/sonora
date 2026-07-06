import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts", "src/fetch.ts"],
  format: ["esm"],
  dts: true,
  sourcemap: true,
  clean: true,
});

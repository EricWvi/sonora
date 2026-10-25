import { defineConfig, mergeConfig } from "vitest/config";
import viteConfig from "./vite.config";

// 单元测试只跑纯逻辑（播放控制器与队列），不需要 DOM；音频引擎与数据源换成假的实现。
export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      environment: "node",
      include: ["src/**/*.test.ts"],
      setupFiles: ["./vitest.setup.ts"],
      // 包内的 ESM 导入不带扩展名，Node 无法直接加载，交给 Vite 转换
      server: {
        deps: {
          inline: ["@material/material-color-utilities", "@imsyy/color-utils"],
        },
      },
    },
  }),
);

import vue from "eslint-plugin-vue";
import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";

export default [
  {
    ignores: [
      "node_modules/**",
      "dist/**",
      "auto-imports.d.ts",
      "components.d.ts",
    ],
  },
  ...vue.configs["flat/essential"],
  {
    files: ["**/*.ts", "**/*.vue"],
    languageOptions: { parserOptions: { parser: tsParser } },
    plugins: { "@typescript-eslint": tsPlugin },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      "vue/multi-word-component-names": "off",
      "vue/no-v-html": "off",
      // 与 SPlayer 相同：以 _ 开头的变量与参数表示有意不用
      "@typescript-eslint/no-unused-vars": [
        "error",
        { varsIgnorePattern: "^_", argsIgnorePattern: "^_" },
      ],
    },
  },
  { files: ["**/*.ts"], languageOptions: { parser: tsParser } },
];

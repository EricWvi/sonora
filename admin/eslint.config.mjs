import vue from "eslint-plugin-vue";
import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";

export default [
  { ignores: ["node_modules/**", "dist/**"] },
  ...vue.configs["flat/essential"],
  {
    files: ["**/*.ts", "**/*.vue"],
    languageOptions: { parserOptions: { parser: tsParser } },
    plugins: { "@typescript-eslint": tsPlugin },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      "vue/multi-word-component-names": "off",
    },
  },
  { files: ["**/*.ts"], languageOptions: { parser: tsParser } },
];

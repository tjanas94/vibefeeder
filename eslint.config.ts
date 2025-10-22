import eslint from "@eslint/js";
import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";
import prettierRecommended from "eslint-plugin-prettier/recommended";

export default defineConfig(
  {
    ignores: ["dist/", "node_modules/", "**/*.d.ts"],
  },
  prettierRecommended,
  {
    rules: {
      "prettier/prettier": "off",
    },
  },
  eslint.configs.recommended,
  {
    extends: [tseslint.configs.recommended],
    plugins: {
      "@typescript-eslint": tseslint.plugin,
    },
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        projectService: true,
      },
    },
    rules: {
      "@typescript-eslint/no-floating-promises": "error",
    },
  },
  {
    files: ["**/*.js", "**/*.mjs"],
    extends: [tseslint.configs.disableTypeChecked],
  },
);

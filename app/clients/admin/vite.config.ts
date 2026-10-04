/// <reference types="vitest/config" />

import { paraglideVitePlugin } from "@inlang/paraglide-js";
import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import { devtools } from "@tanstack/devtools-vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact, { reactCompilerPreset } from "@vitejs/plugin-react";
import { visualizer } from "rollup-plugin-visualizer";
import { defineConfig, loadEnv } from "vite";

const config = defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return {
    resolve: { tsconfigPaths: true },
    ssr: {
      noExternal: ["@ant-design/x-markdown"],
    },
    server: {
      proxy: {
        "^/connect/cyber\\.system\\.": {
          target: env.SYSTEM_CONNECT_API_URL,
          ws: true,
          rewrite: (path) => path.replace(/^\/connect/, ""),
        },
        "^/connect/cyber\\.agent\\.": {
          target: env.AGENT_CONNECT_API_URL,
          ws: true,
          rewrite: (path) => path.replace(/^\/connect/, ""),
        },
      },
    },
    plugins: [
      devtools(),
      paraglideVitePlugin({
        project: "./project.inlang",
        outdir: "./src/paraglide",
        strategy: ["custom-smart-preferred", "url", "cookie", "baseLocale"],
      }),
      tailwindcss(),
      tanstackStart(),
      viteReact(),
      babel({ presets: [reactCompilerPreset()] }),
      mode === "analyze" &&
        visualizer({
          open: true,
          filename: "stats.html",
          gzipSize: true,
          brotliSize: true,
        }),
    ].filter(Boolean),
    test: {
      environment: "jsdom",
      include: ["src/**/*.test.{ts,tsx}"],
    },
  };
});

export default config;

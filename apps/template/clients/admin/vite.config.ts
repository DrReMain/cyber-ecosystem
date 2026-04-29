import { paraglideVitePlugin } from "@inlang/paraglide-js"
import babel from "@rolldown/plugin-babel"
import tailwindcss from "@tailwindcss/vite"
import { devtools as TanstackDevtools } from "@tanstack/devtools-vite"
import { tanstackStart } from "@tanstack/react-start/plugin/vite"
import viteReact, { reactCompilerPreset } from "@vitejs/plugin-react"
import { defineConfig, loadEnv } from "vite"

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "")

  return {
    preview: {
      port: parseInt(env.PORT || "10000", 10),
    },
    server: {
      port: 10000,
      proxy: {
        "/connect": {
          target: env.CONNECT_API_URL || "http://localhost:13000",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/connect/, ""),
        },
        "/http": {
          target: env.HTTP_API_URL || "http://localhost:11000",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/http/, ""),
        },
        "/glitchtip": {
          target: env.PROXY_GLITCHTIP_URL || "http://localhost:8000",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/glitchtip/, ""),
        },
        "/otel": {
          target: env.PROXY_OTEL_URL || "http://localhost:4318",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/otel/, ""),
        },
      },
    },
    resolve: { tsconfigPaths: true },
    plugins: [
      TanstackDevtools(),
      paraglideVitePlugin({
        project: "./project.inlang",
        outdir: "./src/paraglide",
        strategy: ["custom-smart-preferred", "url", "baseLocale"],
      }),
      tailwindcss(),
      tanstackStart(),
      viteReact(),
      babel({ presets: [reactCompilerPreset()] }),
    ],
    build: {
      sourcemap: "hidden",
    },
    devtools:
      process.env.VITE_DEVTOOLS === "true"
        ? {
            enabled: true,
            environments: ["client"],
            host: "127.0.0.1",
            port: 9999,
            open: false,
          }
        : {
            enabled: false,
          },
  }
})

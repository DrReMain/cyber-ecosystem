import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  projects: [{ name: "chromium", use: { browserName: "chromium" } }],
  use: { baseURL: process.env.E2E_BASE_URL ?? "http://localhost:5179" },
  webServer: {
    command: "pnpm exec vite dev --port 5179 --strictPort",
    url: "http://localhost:5179",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});

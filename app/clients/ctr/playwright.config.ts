import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  // One worker, always: the fidelity seeds bench their members against
  // cohort pools keyed by birth year + sex in the shared dev database —
  // two members seeded in parallel join each other's pools and the
  // frozen percentiles drift (a 22-row bench moves 78 → 79). Every
  // seeded test must clean up before the next one seeds.
  workers: 1,
  projects: [{ name: "chromium", use: { browserName: "chromium" } }],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:5189",
    actionTimeout: 15_000,
  },
  webServer: {
    command: "pnpm exec vite dev --port 5189 --strictPort",
    url: "http://localhost:5189",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});

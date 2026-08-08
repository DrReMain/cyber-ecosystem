import { defineConfig } from "@hey-api/openapi-ts";

// OpenAPI (gen/openapi/openapi.yaml) → TypeScript SDK (this package).
// hey-api relies on the TS compiler API (removed in native TS 7), so typescript@6 is
// pinned as a devDep to run the generator in isolation from the repo's TS 7 root.
// fetch client + no runtimeConfigPath: the SDK is transport-agnostic — each consumer
// (web/RN/Taro) injects its own base URL/client at runtime.
// This file must live in the package: hey-api resolves "@hey-api/openapi-ts" from the
// config's location. clean-gen preserves it across regeneration.
export default defineConfig({
  input: "../openapi/openapi.yaml",
  output: { path: ".", clean: false },
  plugins: ["@hey-api/client-fetch", "@hey-api/sdk", "@hey-api/typescript"],
});

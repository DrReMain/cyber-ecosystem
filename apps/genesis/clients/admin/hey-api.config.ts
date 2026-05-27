import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../../../genesis/gen/oas/openapi.yaml",
  output: "src/services/openapi",
  plugins: [
    {
      name: "@hey-api/client-ky",
      runtimeConfigPath: "./src/services/http-config",
    },
  ],
})

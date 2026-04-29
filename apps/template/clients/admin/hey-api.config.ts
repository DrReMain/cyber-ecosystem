import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../../../template/gen/oas/openapi.yaml",
  output: "src/services/openapi",
  plugins: [
    {
      name: "@hey-api/client-ky",
      runtimeConfigPath: "./src/services/http-config",
    },
  ],
})

import type { Plugin } from "vite";

// The React DevTools MCP bridge must install its hook before react-dom loads.
// TanStack Start's dev client entry is a framework-generated virtual module
// running ahead of every app module, so prepending the register import there
// is the only reliable window; `serve`-only apply keeps production builds
// untouched — the injected module does not exist outside the dev server.
export function reactDevToolsMcp(): Plugin {
  return {
    name: "react-devtools-cdt-mcp",
    apply: (_config, { command }) => command === "serve",
    transform(code, id) {
      if (
        id.includes("virtual:tanstack-start-dev-client-entry") &&
        this.environment?.name === "client"
      ) {
        return `import "react-devtools-cdt-mcp/register";\n${code}`;
      }
    },
  };
}

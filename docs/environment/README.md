# Recommended agent environment

An advice channel, not working-process law: the developer's coding agent surfaces these recommendations to the developer. The repo is fully functional without any tool listed here — each entry states what its absence degrades, so a developer can decide. Setup is manual by design: follow the upstream docs; an agent recommends but never executes the installation.

| Tool | Identity | Serves | Absent → |
|---|---|---|---|
| codegraph | npm `@colbymchenry/codegraph` — MCP server (`codegraph serve --mcp`) | edit-time context: symbol source, call paths, blast radius for indexed code | agents fall back to search-and-read loops — slower, no call-path view |
| memsearch | plugin from the `zilliztech/memsearch` marketplace | cross-session recall of past decisions, per-project collection under `.memsearch/` | semantic recall of history is gone; plain markdown remains readable |
| chrome-devtools-mcp | npm `chrome-devtools-mcp` — MCP server (`npx -y chrome-devtools-mcp@latest --auto-connect`) | driving live verification of client surfaces (tanstack CONVENTIONS §11 reachable-surface rule) | live verification degrades to a manual developer walkthrough |

Per-tool notes:

- **codegraph** — upstream: github.com/colbymchenry/codegraph. Per clone: `codegraph init` builds the index under `.codegraph/` (gitignored). Agents reach it through the codegraph MCP tools.
- **memsearch** — upstream: github.com/zilliztech/memsearch. Per-project state under `.memsearch/` (gitignored); installed and updated through the plugin marketplace.
- **chrome-devtools-mcp** — upstream: github.com/ChromeDevTools/chrome-devtools-mcp. Global npm install, MCP registration in the client config, plus the Chrome-side setup its README specifies.
  - Connection reality (Chrome ≥136): `--remote-debugging-port` no longer works on the default profile and `open --args` cannot retrofit a running instance — `--auto-connect` attaches only after the in-browser switch at `chrome://inspect/#remote-debugging` is flipped once (Chrome 144+; writes DevToolsActivePort into the default profile).
  - The attached session is the user's real browser: login state survives, but extensions inject console noise on some machines (fetch-wrapper and html-stamping families are handled app-side; expect their noise elsewhere).

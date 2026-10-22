# Recommended agent environment

**Status:** ACTIVE

Optional. The repo is fully functional without any tool listed here — each entry states what its absence degrades, so a developer can decide. Setup is manual by design: follow the upstream docs; an agent recommends but never executes the installation.

| Tool | Identity | Serves | Absent → |
|---|---|---|---|
| codegraph | npm `@colbymchenry/codegraph` — MCP server (`codegraph serve --mcp`) | edit-time context: symbol source, call paths, blast radius for indexed code | agents fall back to search-and-read loops — slower, no call-path view |
| memsearch | plugin from the `zilliztech/memsearch` marketplace | cross-session recall of past decisions, per-project collection under `.memsearch/` | semantic recall of history is gone; plain markdown remains readable |
| chrome-devtools-mcp | npm `chrome-devtools-mcp` — MCP server (`npx -y chrome-devtools-mcp@latest --auto-connect`) | driving live verification of client surfaces (tanstack CONVENTIONS §11 reachable-surface e2e) | live verification degrades to a manual developer walkthrough |

Per-tool notes:

- **codegraph** — upstream: github.com/colbymchenry/codegraph. Per clone: `codegraph init` builds the index under `.codegraph/` (gitignored). Agents reach it through the codegraph MCP tools.
- **memsearch** — upstream: github.com/zilliztech/memsearch. Per-project state under `.memsearch/` (gitignored); installed and updated through the plugin marketplace.
- **chrome-devtools-mcp** — upstream: github.com/ChromeDevTools/chrome-devtools-mcp. Global npm install, MCP registration in the client config, plus the Chrome-side setup its README specifies.
 Connection reality on Chrome ≥136 (2026-10-22, first live session): `--remote-debugging-port` no longer works on the default profile and `open --args` cannot retrofit a running instance — `--auto-connect` attaches only after someone flips the in-browser switch at `chrome://inspect/#remote-debugging` (Chrome 144+; writes DevToolsActivePort into the default profile). The attached session is the user's real browser: login state survives, but extensions inject (沉浸式翻译 stamps `<html>` attributes pre-hydration; ImageAssistant wraps `window.fetch` and console-errors every rejected fetch — the fetch-wrapper family is absorbed by the shared transport's header-cancellation shield, the html-stamping family by client-side accommodations — expect their noise on other machines).

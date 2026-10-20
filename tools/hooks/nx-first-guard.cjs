#!/usr/bin/env node
/**
 * PreToolUse guard (Bash): deny high-confidence Nx bypasses (root AGENTS.md §3/§4).
 *
 * Fail-open by design — a broken guard must never brick the Bash tool: any
 * parse error or unexpected input exits 0 with no output. Denials carry the
 * lawful alternative in the reason.
 */
const BYPASSES = [
  [
    /\bbuf\s+(generate|lint|format|breaking)\b/,
    "buf workflows are Nx targets — use ./nx run proto:<generate|lint|format|breaking>",
  ],
  [
    /\bgo\s+(build|generate)\b/,
    "Go build/generate workflows are Nx targets — use the owning service's targets (ad-hoc go test/vet for scoped verification stay allowed)",
  ],
  [/\bgolangci-lint\b/, "golangci-lint runs through ./nx run tools:go:lint"],
];
const NX_RE = /(\.\/nx\b|\bnx\s+(run|run-many|show|graph)\b)/;

let raw = "";
try {
  raw = require("node:fs").readFileSync(0, "utf8");
} catch {
  process.exit(0);
}
let event;
try {
  event = JSON.parse(raw);
} catch {
  process.exit(0);
}
const command = event?.tool_input && String(event.tool_input.command || "");
if (!command || NX_RE.test(command)) process.exit(0);
for (const [re, reason] of BYPASSES) {
  if (re.test(command)) {
    process.stdout.write(
      JSON.stringify({
        hookSpecificOutput: {
          hookEventName: "PreToolUse",
          permissionDecision: "deny",
          permissionDecisionReason: `${reason} (AGENTS.md §3: MUST use Nx for workflows with declared targets)`,
        },
      }),
    );
    process.exit(0);
  }
}
process.exit(0);

// Cross-OS cleanup of generated output before buf generate.
//
// buf generate is purely additive — it never removes files whose source proto
// was deleted, leaving stale "orphan" outputs. This script wipes a single gen
// subtree's GENERATED files so regeneration reproduces a clean, orphan-free
// result — while preserving hand-maintained package files (package.json,
// tsconfig.json) that buf does NOT produce. For example, gen/ts/package.json
// declares the @cyber-ecosystem/gen-ts workspace package and its protobuf dep;
// wiping it would break the workspace, so it must survive a regen.
//
// Invoked from the proto Nx generate:* targets as:
//   node proto/scripts/clean-gen.cjs <dir>   (e.g. gen/go, gen/ts, gen/openapi)
//
// Uses fs only (Node >= 14.14): no dependencies, no shell quoting — works on
// macOS, Linux, and Windows. Exits 2 if no target directory is given.
const fs = require('node:fs')
const path = require('node:path')

// Files at the gen package root that are hand-maintained (not buf output) and
// must survive a regen. Matched only at the target root, never inside generated
// sub-packages.
const PRESERVE = new Set(['package.json', 'tsconfig.json'])

const dir = process.argv[2]
if (!dir) {
  console.error('clean-gen: missing target directory argument (e.g. gen/go)')
  process.exit(2)
}

function clean(target, isRoot) {
  // No-op when the target doesn't exist yet (fresh checkout with no gen/, or a
  // manually wiped tree). buf generate (re)creates the tree from scratch.
  if (!fs.existsSync(target)) {
    return
  }
  for (const entry of fs.readdirSync(target, { withFileTypes: true })) {
    if (isRoot && entry.isFile() && PRESERVE.has(entry.name)) {
      continue
    }
    const p = path.join(target, entry.name)
    if (entry.isDirectory()) {
      clean(p, false)
      // Remove subdirectories that became empty after wiping generated files.
      // fs.rm requires recursive:true for directories (EISDIR otherwise).
      if (fs.readdirSync(p).length === 0) {
        fs.rmSync(p, { recursive: true, force: true })
      }
    } else {
      fs.rmSync(p, { force: true })
    }
  }
}

clean(dir, true)

#!/usr/bin/env node
/**
 * Pointer-integrity sentinel for the three documentation lanes (docs/README.md 2.6).
 *
 * Checks:
 *   1. every literal docs/... or .claude/... path referenced from AGENTS.md,
 *      CLAUDE.md, and every doc/skill body exists on disk;
 *   2. every same-line "docs/<file>.md <section N>" pointer resolves to a real
 *      "## N." / "## N)" heading in the target (quoted heading names verified
 *      as substrings) -- this is the pointer grammar skills must use;
 *   3. the AGENTS.md area index and the conventions area directories agree
 *      in both directions;
 *   4. every conventions doc carries a Status line; every roadmap doc
 *      carries a destination (归宿) header near the top;
 *   5. roadmap isolation (docs/README.md 2.7): no docs/roadmap reference in
 *      AGENTS.md, CLAUDE.md, skills, or conventions.
 */
const fs = require("node:fs");
const path = require("node:path");

const ROOT = path.resolve(__dirname, "..", "..");
const fail = [];
const bad = (msg) => fail.push(msg);

const refFiles = [path.join(ROOT, "AGENTS.md"), path.join(ROOT, "CLAUDE.md")];
const skillsDir = path.join(ROOT, ".claude", "skills");
if (fs.existsSync(skillsDir)) {
  for (const name of fs.readdirSync(skillsDir).sort()) {
    const p = path.join(skillsDir, name, "SKILL.md");
    if (fs.existsSync(p)) refFiles.push(p);
  }
}
const walk = (dir) => {
  const out = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.name.endsWith(".md")) out.push(p);
  }
  return out;
};
refFiles.push(...walk(path.join(ROOT, "docs")));

// literal repo paths (docs/... or .claude/...); templates with < > * are skipped
const PATH_RE = /(?<![\w<*>])(docs\/[A-Za-z0-9_\-./]+|\.claude\/[A-Za-z0-9_\-./]+)/g;
// same-line "docs/<file>.md §N" with optional quoted heading name
const REF_RE = /(docs\/[A-Za-z0-9_\-./]+?\.md)\s*§(\d+[a-z]?)(?:\s*"([^"]+)")?/g;
const HEAD_RE = /^##+\s+(\d+[a-z]?)[.)]/gm;

for (const f of refFiles) {
  if (!fs.existsSync(f)) continue;
  const text = fs.readFileSync(f, "utf8");
  const rel = path.relative(ROOT, f);
  for (const m of text.matchAll(PATH_RE)) {
    const p = m[1].replace(/\.$/, "");
    if (!fs.existsSync(path.join(ROOT, p))) bad(`${rel}: dangling path \`${p}\``);
  }
  for (const m of text.matchAll(REF_RE)) {
    const [_ref, p, sec, quoted] = m;
    const t = path.join(ROOT, p);
    if (!fs.existsSync(t)) continue; // already flagged by check 1
    const content = fs.readFileSync(t, "utf8");
    const heads = new Set([...content.matchAll(HEAD_RE)].map((h) => h[1]));
    if (!heads.has(sec)) bad(`${rel}: \`${p} §${sec}\` -- no "## ${sec}." heading in target`);
    if (quoted && !content.includes(quoted))
      bad(`${rel}: \`${p} §${sec} "${quoted}"\` -- heading text not found in target`);
  }
}

// area index bidirectional sync
const agents = fs.readFileSync(path.join(ROOT, "AGENTS.md"), "utf8");
const indexed = new Set(
  [...agents.matchAll(/`(docs\/conventions\/[a-z0-9-]+\/CONVENTIONS\.md)`/g)].map((m) => m[1]),
);
const convRoot = path.join(ROOT, "docs", "conventions");
const actual = new Set(
  fs
    .readdirSync(convRoot, { withFileTypes: true })
    .filter((e) => e.isDirectory() && fs.existsSync(path.join(convRoot, e.name, "CONVENTIONS.md")))
    .map((e) => `docs/conventions/${e.name}/CONVENTIONS.md`),
);
for (const missing of [...actual].filter((x) => !indexed.has(x)).sort())
  bad(`AGENTS.md area index is missing ${missing}`);
for (const ghost of [...indexed].filter((x) => !actual.has(x)).sort())
  bad(`AGENTS.md area index references non-existent ${ghost}`);

// status & destination headers
for (const area of [...actual].sort()) {
  const p = path.join(ROOT, area);
  if (!fs.readFileSync(p, "utf8").includes("**Status:**"))
    bad(`${area}: missing \`**Status:**\` line`);
}
const roadmapDir = path.join(ROOT, "docs", "roadmap");
for (const name of fs
  .readdirSync(roadmapDir)
  .filter((n) => n.endsWith(".md"))
  .sort()) {
  const head = fs.readFileSync(path.join(roadmapDir, name), "utf8").slice(0, 900);
  if (!head.includes("归宿"))
    bad(`docs/roadmap/${name}: missing 归宿 (destination) header near the top`);
}

// roadmap isolation (docs/README.md 2.7): the durable lane never references
// roadmap content
const durableFiles = [path.join(ROOT, "AGENTS.md"), path.join(ROOT, "CLAUDE.md")];
for (const name of fs.readdirSync(skillsDir).sort()) {
  const p = path.join(skillsDir, name, "SKILL.md");
  if (fs.existsSync(p)) durableFiles.push(p);
}
durableFiles.push(...walk(convRoot));
for (const f of durableFiles) {
  if (/docs\/roadmap/.test(fs.readFileSync(f, "utf8")))
    bad(`${path.relative(ROOT, f)}: references roadmap content (isolation law, docs/README.md 2.7)`);
}

if (fail.length) {
  console.log("docs:check FAILED");
  for (const msg of fail) console.log(" -", msg);
  process.exitCode = 1;
} else {
  console.log(
    "docs:check OK — pointers resolve, area index synced, status/destination headers present",
  );
}

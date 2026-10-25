#!/usr/bin/env node
/**
 * Paraglide message parity guard (the i18n directory-ownership law).
 *
 * Scope: clients that declare inlang/paraglide — the presence of a
 * project.inlang settings file is the opt-in; clients on other i18n stacks
 * are skipped. Verifies every message directory: registration,
 * directory-prefixed keys, and per-locale parity (key set, key order,
 * placeholder set). The paraglide compile step resolves a key missing from
 * one locale from the base locale and exits 0 — no other gate catches this
 * class.
 *
 * Usage: node tools/scripts/paraglide-parity.cjs <clientRoot>
 */
const fs = require("node:fs");
const path = require("node:path");

const clientRoot = path.resolve(process.argv[2] || ".");
const fail = [];
const bad = (msg) => fail.push(msg);

const settingsPath = path.join(clientRoot, "project.inlang", "settings.json");
const messagesDir = path.join(clientRoot, "messages");
if (!fs.existsSync(settingsPath) || !fs.existsSync(messagesDir)) {
  console.log(
    `paraglide-parity: skipped (${path.basename(clientRoot)} — no project.inlang settings)`,
  );
  process.exit(0);
} else {
  const settings = JSON.parse(fs.readFileSync(settingsPath, "utf8"));
  const locales = settings.locales || [];
  const base = settings.baseLocale || locales[0];
  const patterns = settings["plugin.inlang.messageFormat"]?.pathPattern || [];
  const registered = patterns.map((p) =>
    p.replace(/^\.\/messages\//, "").replace(/\/\{locale\}\.json$/, ""),
  );
  const onDisk = fs
    .readdirSync(messagesDir)
    .filter((n) => fs.statSync(path.join(messagesDir, n)).isDirectory());
  for (const d of onDisk) {
    if (!registered.includes(d)) {
      bad(
        `messages/${d}/ is not registered in project.inlang settings (the three-point registration law)`,
      );
    }
  }
  const placeholders = (v) => new Set(JSON.stringify(v).match(/\{[a-zA-Z_][a-zA-Z0-9_]*\}/g) || []);
  for (const d of registered) {
    if (!onDisk.includes(d)) {
      bad(`registered directory messages/${d}/ does not exist`);
      continue;
    }
    const dir = path.join(messagesDir, d);
    const prefix = `${d.replace(/-/g, "_")}_`;
    const files = {};
    for (const loc of locales) {
      const f = path.join(dir, `${loc}.json`);
      if (!fs.existsSync(f)) {
        bad(`messages/${d}/${loc}.json missing (a registered locale)`);
        continue;
      }
      try {
        files[loc] = JSON.parse(fs.readFileSync(f, "utf8"));
      } catch (e) {
        bad(`messages/${d}/${loc}.json is not valid JSON: ${e.message}`);
      }
    }
    const baseKeys = files[base] ? Object.keys(files[base]) : [];
    for (const k of baseKeys) {
      if (!k.startsWith(prefix)) {
        bad(`messages/${d}/${base}.json key \`${k}\` lacks the directory prefix \`${prefix}\``);
      }
    }
    for (const [loc, data] of Object.entries(files)) {
      if (loc === base) continue;
      const keys = Object.keys(data);
      const miss = baseKeys.filter((k) => !keys.includes(k));
      const extra = keys.filter((k) => !baseKeys.includes(k));
      if (miss.length) bad(`messages/${d}/${loc}.json missing keys vs ${base}: ${miss.join(", ")}`);
      if (extra.length) bad(`messages/${d}/${loc}.json extra keys vs ${base}: ${extra.join(", ")}`);
      if (!miss.length && !extra.length && keys.join("\u0000") !== baseKeys.join("\u0000")) {
        bad(
          `messages/${d}/${loc}.json key order differs from ${base} (deliberate order, never re-sorted)`,
        );
      }
      for (const k of baseKeys) {
        if (!(k in data)) continue;
        const a = placeholders(files[base][k]);
        const b = placeholders(data[k]);
        for (const x of a)
          if (!b.has(x)) bad(`messages/${d}/${loc}.json \`${k}\` missing placeholder ${x}`);
        for (const x of b)
          if (!a.has(x)) bad(`messages/${d}/${loc}.json \`${k}\` extra placeholder ${x}`);
      }
    }
  }
}

if (fail.length) {
  console.log(`paraglide-parity: FAILED (${fail.length}) — ${path.basename(clientRoot)}`);
  for (const msg of fail) console.log(" -", msg);
  process.exitCode = 1;
} else {
  console.log(`paraglide-parity: OK (${path.basename(clientRoot)})`);
}

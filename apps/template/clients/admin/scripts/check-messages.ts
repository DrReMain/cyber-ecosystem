import { parse as parseYaml, stringify } from "yaml"
import {
  compileParaglide,
  deleteFromSchemaNode,
  flatKeyToPath,
  I18N_DIR,
  MESSAGES_DIR,
  readJson,
  readSchema,
  readSettings,
  SRC_DIR,
  writeSortedJson,
} from "./lib/i18n"

const M_KEY_RE = /\bm\.([a-z][a-z0-9_]*)\s*\(/

function checkDiff(
  locale: string,
  messageKeys: Set<string>,
  schemaKeys: Set<string>,
): boolean {
  const orphaned = [...messageKeys].filter((k) => !schemaKeys.has(k))
  const missing = [...schemaKeys].filter((k) => !messageKeys.has(k))
  if (orphaned.length === 0 && missing.length === 0) return false
  console.warn(`\n✗ ${locale}:`)
  for (const k of orphaned) console.warn(`  orphaned: ${k}`)
  for (const k of missing) console.warn(`  missing:  ${k}`)
  return true
}

function warnEmpty(locale: string, messages: Record<string, unknown>): void {
  const emptyKeys = Object.entries(messages)
    .filter(([, v]) => typeof v === "string" && v === "")
    .map(([k]) => k)
  if (emptyKeys.length > 0) {
    console.log(
      `\n⚠ ${locale}: ${emptyKeys.length} empty value(s): ${emptyKeys.join(", ")}`,
    )
  }
}

async function findUsedKeys(): Promise<Set<string>> {
  const glob = new Bun.Glob("**/*.{ts,tsx}")
  const used = new Set<string>()
  for await (const file of glob.scan({ cwd: SRC_DIR })) {
    if (file.startsWith("paraglide/")) continue
    const content = await Bun.file(`${SRC_DIR}/${file}`).text()
    for (const line of content.split("\n")) {
      const m = line.match(M_KEY_RE)
      if (m) used.add(m[1])
    }
  }
  return used
}

async function removeUnusedKeys(keys: string[]): Promise<void> {
  const schemaPath = `${I18N_DIR}/schema.yaml`
  const schemaRaw = await Bun.file(schemaPath).text()
  const headerMatch = schemaRaw.match(
    /^([\s\S]*?# ==================== 应用全局)/,
  )
  const header = headerMatch?.[1] ?? ""

  const schemaTree = parseYaml(schemaRaw) as Parameters<
    typeof deleteFromSchemaNode
  >[0]
  for (const key of keys) {
    deleteFromSchemaNode(schemaTree, flatKeyToPath(key))
  }
  await Bun.write(schemaPath, `${header}\n${stringify(schemaTree)}\n`)

  const { locales } = await readSettings()
  for (const locale of locales) {
    const path = `${MESSAGES_DIR}/${locale}.json`
    const raw = await readJson<Record<string, unknown>>(path)
    const { $schema, ...messages } = raw
    for (const key of keys) {
      delete messages[key]
    }
    await writeSortedJson(path, { $schema: $schema as string, ...messages })
  }

  await compileParaglide()
}

async function ask(question: string): Promise<string> {
  process.stdout.write(question)
  const rl = Bun.stdin.stream()
  const reader = rl.getReader()
  const { value } = await reader.read()
  reader.releaseLock()
  return new TextDecoder().decode(value).trim()
}

async function handleUnusedKeys(unusedKeys: string[]): Promise<void> {
  console.warn(`\n⚠ Unused keys (${unusedKeys.length}):`)
  for (const [i, k] of unusedKeys.entries()) {
    console.warn(`  ${i + 1}. ${k}`)
  }

  if (!process.stdin.isTTY) return

  const answer = await ask(
    `\n  Remove unused keys? Enter numbers (e.g. "1,3") or "all" or "n": `,
  )
  if (answer.toLowerCase() === "all") {
    await removeUnusedKeys(unusedKeys)
    console.log(`  ✓ Removed ${unusedKeys.length} unused key(s).`)
  } else if (answer.toLowerCase() !== "n" && answer !== "") {
    const indices = answer
      .split(",")
      .map((s) => Number.parseInt(s.trim(), 10) - 1)
      .filter((i) => i >= 0 && i < unusedKeys.length)
    if (indices.length > 0) {
      const selected = indices.map((i) => unusedKeys[i])
      await removeUnusedKeys(selected)
      console.log(
        `  ✓ Removed ${selected.length} key(s): ${selected.join(", ")}`,
      )
    }
  }
}

async function check(): Promise<void> {
  const { locales } = await readSettings()
  const schemaKeys = await readSchema()
  const schemaKeySet = new Set(schemaKeys.keys())

  let hasError = false
  for (const locale of locales) {
    const raw = await readJson<Record<string, unknown>>(
      `${MESSAGES_DIR}/${locale}.json`,
    )
    const { $schema, ...messages } = raw
    const msg = messages as Record<string, unknown>
    const messageKeys = new Set(Object.keys(msg))

    if (checkDiff(locale, messageKeys, schemaKeySet)) hasError = true
    warnEmpty(locale, msg)
  }

  if (hasError) {
    console.error("\nOrphan/missing keys detected. Run `messages:gen` to sync.")
    process.exit(1)
  }

  const usedKeys = await findUsedKeys()
  const unusedKeys = [...schemaKeySet].filter((k) => !usedKeys.has(k))
  if (unusedKeys.length > 0) {
    await handleUnusedKeys(unusedKeys)
  }

  console.log(
    `\n✓ All ${locales.length} locale(s) are in sync with schema (${schemaKeySet.size} keys).`,
  )
}

await check()

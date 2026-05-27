import {
  compileParaglide,
  MESSAGES_DIR,
  readJson,
  readSchema,
  readSettings,
  writeSortedJson,
} from "./lib/i18n"

async function readMessages(locale: string): Promise<Record<string, unknown>> {
  const raw = await readJson<Record<string, unknown>>(
    `${MESSAGES_DIR}/${locale}.json`,
  )
  const { $schema, ...messages } = raw
  return messages as Record<string, unknown>
}

function isNonEmpty(value: unknown): boolean {
  if (value == null) return false
  if (typeof value === "string" && value !== "") return true
  return Array.isArray(value) && value.length > 0
}

function isVariantAllEmpty(variants: unknown[]): boolean {
  return variants.some((v) => {
    if (typeof v !== "object" || v === null) return true
    const match = (v as Record<string, unknown>).match
    if (typeof match !== "object" || match === null) return true
    return Object.values(match as Record<string, unknown>).every(
      (val) => typeof val === "string" && val === "",
    )
  })
}

async function processLocale(
  locale: string,
  keys: Map<string, string>,
): Promise<void> {
  const existing = await readMessages(locale)
  const generated: Record<string, unknown> = {}
  let added = 0
  let preserved = 0
  const warnings: string[] = []

  for (const [key] of keys) {
    const existingVal = existing[key]
    if (isNonEmpty(existingVal)) {
      generated[key] = existingVal
      if (Array.isArray(existingVal) && isVariantAllEmpty(existingVal)) {
        warnings.push(`  ⚠ ${key}: variant with empty match values`)
      }
      preserved++
    } else {
      generated[key] = ""
      added++
    }
  }

  const removedKeys = Object.keys(existing).filter((k) => !(k in generated))
  for (const k of removedKeys) {
    warnings.push(`  ⚠ ${k}: in ${locale}.json but not in schema`)
  }

  await writeSortedJson(`${MESSAGES_DIR}/${locale}.json`, {
    $schema: "https://inlang.com/schema/inlang-message-format",
    ...generated,
  })
  console.log(
    `✓ ${locale}: ${preserved} preserved, ${added} new (empty), ${removedKeys.length} removed`,
  )
  for (const w of warnings) {
    console.log(w)
  }
}

async function generate(): Promise<void> {
  const { locales } = await readSettings()
  const keys = await readSchema()

  for (const locale of locales) {
    await processLocale(locale, keys)
  }

  console.log(`\nTotal keys in schema: ${keys.size}`)

  await compileParaglide()
  console.log("✓ Paraglide compiled")
}

await generate()

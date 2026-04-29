import { parse as parseYaml } from "yaml"

export const ROOT = import.meta.dir.replace(/\/scripts(\/lib)?$/, "")
export const SRC_DIR = `${ROOT}/src`
export const I18N_DIR = `${ROOT}/i18n`
export const INLANG_DIR = `${ROOT}/project.inlang`
export const MESSAGES_DIR = `${ROOT}/messages`

export interface Settings {
  baseLocale: string
  locales: string[]
}

export interface SchemaNode {
  [key: string]: string | SchemaNode
}

export function isLeaf(value: string | SchemaNode): value is string {
  return typeof value === "string"
}

export function flattenSchema(
  node: SchemaNode,
  prefix = "",
): Map<string, string> {
  const result = new Map<string, string>()
  for (const [key, value] of Object.entries(node)) {
    if (key === "") continue
    const flatKey = prefix ? `${prefix}_${key}` : key
    if (isLeaf(value)) {
      if (value.trim() !== "") result.set(flatKey, value)
    } else if (value !== null) {
      for (const [k, v] of flattenSchema(value, flatKey)) {
        result.set(k, v)
      }
    }
  }
  return result
}

export async function readJson<T>(path: string): Promise<T> {
  const file = Bun.file(path)
  if (!(await file.exists())) return {} as T
  return await file.json()
}

export async function readSettings(): Promise<Settings> {
  const settings = await readJson<Settings>(`${INLANG_DIR}/settings.json`)
  if (!settings.baseLocale || !settings.locales?.length) {
    console.error("Invalid settings.json: missing baseLocale or locales")
    process.exit(1)
  }
  return settings
}

export async function readSchema(): Promise<Map<string, string>> {
  const schemaPath = `${I18N_DIR}/schema.yaml`
  const schemaFile = Bun.file(schemaPath)
  if (!(await schemaFile.exists())) {
    console.error(`Schema not found: ${schemaPath}`)
    process.exit(1)
  }
  const schema = parseYaml(await schemaFile.text()) as SchemaNode
  if (typeof schema !== "object" || schema === null) {
    console.error("Schema is not a valid YAML object")
    process.exit(1)
  }
  const keys = flattenSchema(schema)
  if (keys.size === 0) {
    console.error("Schema contains no keys. Aborting.")
    process.exit(1)
  }
  return keys
}

export function flatKeyToPath(key: string): string[] {
  return key.split("_")
}

export function deleteFromSchemaNode(
  node: SchemaNode,
  path: string[],
): boolean {
  if (path.length === 1) {
    if (node[path[0]] !== undefined) {
      delete node[path[0]]
      return true
    }
    return false
  }
  const [head, ...rest] = path
  const child = node[head]
  if (!child || isLeaf(child)) return false
  const deleted = deleteFromSchemaNode(child as SchemaNode, rest)
  if (deleted && Object.keys(child as SchemaNode).length === 0) {
    delete node[head]
  }
  return deleted
}

export async function writeSortedJson(
  path: string,
  data: Record<string, unknown>,
): Promise<void> {
  const sorted = Object.fromEntries(
    Object.entries(data).sort(([a], [b]) => a.localeCompare(b)),
  )
  await Bun.write(path, `${JSON.stringify(sorted, null, 2)}\n`)
}

export async function compileParaglide(): Promise<void> {
  const proc = Bun.spawn(
    [
      "npx",
      "@inlang/paraglide-js",
      "compile",
      "--project",
      "./project.inlang",
      "--outdir",
      "./src/paraglide",
    ],
    { cwd: ROOT, stdout: "pipe", stderr: "pipe" },
  )
  await proc.exited
}

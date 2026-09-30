import { getPartition } from "./schema";

export interface SearchParserSchema<T> {
  shape: Record<string, unknown>;
  parse: (input: unknown) => T;
}

export function makeSearchParser<T extends Record<string, unknown>>(
  schema: SearchParserSchema<T>,
  extraCoerceKeys: readonly string[] = [],
): (input: Record<string, unknown>) => T {
  const coerceKeys = new Set(extraCoerceKeys);
  const { pageNoKey, pageSizeKey } = getPartition(schema);
  if (pageNoKey) coerceKeys.add(pageNoKey);
  if (pageSizeKey) coerceKeys.add(pageSizeKey);

  return (input) => {
    const coerced = Object.fromEntries(
      Object.entries(input).map(([key, value]) => [
        key,
        coerceKeys.has(key) && typeof value === "string" && value !== "" ? Number(value) : value,
      ]),
    );
    return schema.parse(coerced);
  };
}

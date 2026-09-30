import { AREA_PATH } from "../area";

export function chainKeysOf(
  matches: ReadonlyArray<{
    pathname?: string;
    staticData?: { title?: unknown };
  }>,
): string[] {
  const keys: string[] = [];
  for (const mt of matches) {
    const title = mt.staticData?.title;
    const path = mt.pathname?.replace(/\/$/, "");
    if (title && path?.startsWith(`${AREA_PATH}/`)) {
      keys.push(path);
    }
  }
  return keys;
}

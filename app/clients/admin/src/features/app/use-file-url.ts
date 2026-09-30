import { useQuery } from "@connectrpc/connect-query";
import { getFileUrls } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file-FileService_connectquery";
import { useMemo } from "react";

const STALE_MS = 10 * 60 * 1000;

const EMPTY: readonly string[] = [];

export function useFileUrls(ids: readonly string[]): Map<string, string> {
  const sig = ids.join("\u0000");
  const parsed = useMemo(
    () => [...new Set(sig.split("\u0000"))].filter((v) => v !== "").sort(),
    [sig],
  );
  const { data } = useQuery(
    getFileUrls,
    { ids: parsed },
    { enabled: parsed.length > 0, staleTime: STALE_MS },
  );
  return useMemo(() => {
    const raw = (data?.urls ?? {}) as Record<string, { url?: string }>;
    const map = new Map<string, string>();
    for (const [id, entry] of Object.entries(raw)) {
      if (entry?.url) map.set(id, entry.url);
    }
    return map;
  }, [data]);
}

export function useFileUrl(id: string | undefined): string | undefined {
  const urls = useFileUrls(id === undefined || id === "" ? EMPTY : [id]);
  return id === undefined || id === "" ? undefined : urls.get(id);
}

import { useQuery } from "@connectrpc/connect-query";
import { getFileUrls } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file-FileService_connectquery";
import { useEffect, useMemo, useState } from "react";

const STALE_MS = 10 * 60 * 1000; // must stay under the s3 presign TTL (15m default)

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
  const [hydrated, setHydrated] = useState(false);
  useEffect(() => setHydrated(true), []);
  if (!hydrated || id === undefined || id === "") return undefined;
  return urls.get(id);
}

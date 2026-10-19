import {
  buildSearchPatch,
  type SearchStore,
  useUrlSearchStore,
} from "@cyber-ecosystem/shared-antd/use-search";
import { useNavigate } from "@tanstack/react-router";

export function useAreaSearchStore<T extends Record<string, unknown>>(
  area: string,
  search: T,
  parse: (input: Record<string, unknown>) => T,
): SearchStore<T> {
  const navigate = useNavigate();
  return useUrlSearchStore(search, {
    onNavigate: (patch) =>
      navigate({
        to: area,
        search: (prev) => buildSearchPatch(parse(prev ?? {}), patch),
        replace: true,
      }),
  });
}

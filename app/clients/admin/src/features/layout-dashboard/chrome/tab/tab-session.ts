import type { LucideIcon } from "lucide-react";
import { AREA_PATH } from "../../area";
import { isHome, type NavNode } from "../../protocol/nav";
import type { NoParamMessageKey, RouteMenuMeta } from "../../protocol/route-meta";

export interface TabRecord {
  affix: boolean;
  href?: string;
  icon?: LucideIcon;
  key: string;
  title: NoParamMessageKey;
}

const SESSION_KEY = "session_tabbar";
const SESSION_UID_KEY = "session_tabbar_uid";

const recordOf = (path: string, title: NoParamMessageKey, icon?: LucideIcon): TabRecord => ({
  affix: isHome(path),
  icon,
  key: path,
  title,
});

export function tabFromNav(node: NavNode): TabRecord {
  return recordOf(node.path, node.title, node.meta?.icon);
}

export function tabFromMatches(
  matches: ReadonlyArray<{
    pathname?: string;
    staticData?: { menu?: RouteMenuMeta; title?: NoParamMessageKey };
  }>,
): TabRecord | null {
  for (let i = matches.length - 1; i >= 0; i--) {
    const match = matches[i];
    const data = match?.staticData;
    if (match && data?.title) {
      return recordOf(match.pathname?.replace(/\/$/, "") || "/", data.title, data.menu?.icon);
    }
  }
  return null;
}

export function closedTabs(
  tabs: readonly TabRecord[],
  key: string,
): { fallback: string | null; tabs: TabRecord[] } | null {
  const index = tabs.findIndex((t) => t.key === key);
  const target = index === -1 ? undefined : tabs[index];
  if (!target || target.affix || tabs.length <= 1) return null;
  const next = tabs.toSpliced(index, 1);
  const neighbor = next[index] ?? next[index - 1];
  return { fallback: neighbor?.key ?? null, tabs: next };
}

export type BulkScope = "all" | "left" | "other" | "right";

export function closedBulk(
  tabs: readonly TabRecord[],
  anchorKey: string,
  scope: BulkScope,
  activeKey: string,
): { fallback: string | null; tabs: TabRecord[] } | null {
  const anchorIndex = tabs.findIndex((t) => t.key === anchorKey);
  if (anchorIndex === -1) return null;
  const doomed = (t: TabRecord, i: number) => {
    if (t.affix) return false;
    if (scope === "all") return true;
    if (i === anchorIndex) return false;
    if (scope === "left") return i < anchorIndex;
    if (scope === "right") return i > anchorIndex;
    return true; // other
  };
  const set = tabs.filter(doomed);
  if (set.length === 0) return null;
  const closedKeys = new Set(set.map((t) => t.key));
  const next = tabs.filter((t) => !closedKeys.has(t.key));
  const fallbackKey = closedKeys.has(anchorKey) ? next[0]?.key : anchorKey;
  return { fallback: closedKeys.has(activeKey) ? (fallbackKey ?? null) : null, tabs: next };
}

export function pinnedTabs(tabs: readonly TabRecord[], key: string): TabRecord[] | null {
  const index = tabs.findIndex((t) => t.key === key);
  const target = index === -1 ? undefined : tabs[index];
  if (!target || target.affix) return null;
  const moved = { ...target, affix: true };
  const rest = tabs.toSpliced(index, 1);
  let at = rest.length;
  for (let i = 0; i < rest.length; i++) {
    if (!rest[i]?.affix) {
      at = i;
      break;
    }
  }
  return rest.toSpliced(at, 0, moved);
}

export function unpinnedTabs(tabs: readonly TabRecord[], key: string): TabRecord[] | null {
  const index = tabs.findIndex((t) => t.key === key);
  const target = index === -1 ? undefined : tabs[index];
  if (!target?.affix || isHome(key)) return null;
  const moved = { ...target, affix: false };
  const rest = tabs.toSpliced(index, 1);
  let at = 0;
  for (let i = 0; i < rest.length; i++) {
    if (rest[i]?.affix) at = i + 1;
    else break;
  }
  return rest.toSpliced(at, 0, moved);
}

export function movedTabs(
  tabs: readonly TabRecord[],
  fromKey: string,
  toKey: string,
): TabRecord[] | null {
  const from = tabs.findIndex((t) => t.key === fromKey);
  const to = tabs.findIndex((t) => t.key === toKey);
  const a = from === -1 ? undefined : tabs[from];
  const b = to === -1 ? undefined : tabs[to];
  if (!(a && b) || from === to || a.affix !== b.affix) return null;
  const rest = tabs.toSpliced(from, 1);
  return rest.toSpliced(to, 0, a);
}

interface TabSession {
  keys: string[];
  pinned: string[];
  hrefs: Record<string, string>;
}

export function readTabSession(): TabSession {
  try {
    const parsed: unknown = JSON.parse(sessionStorage.getItem(SESSION_KEY) ?? "null");
    if (parsed && typeof parsed === "object") {
      const { keys, pinned, hrefs } = parsed as {
        keys?: unknown;
        pinned?: unknown;
        hrefs?: unknown;
      };
      const list = (v: unknown) =>
        Array.isArray(v) ? v.filter((k): k is string => typeof k === "string") : [];
      const hrefMap = (v: unknown) => {
        const out: Record<string, string> = {};
        if (v && typeof v === "object") {
          for (const [k, val] of Object.entries(v)) {
            if (typeof val === "string") out[k] = val;
          }
        }
        return out;
      };
      return { keys: list(keys), pinned: list(pinned), hrefs: hrefMap(hrefs) };
    }
    return { keys: [], pinned: [], hrefs: {} };
  } catch {
    return { keys: [], pinned: [], hrefs: {} };
  }
}

export function writeTabSession(tabs: readonly TabRecord[]): void {
  const keys = tabs.map((t) => t.key);
  const pinned = tabs.filter((t) => t.affix && !isHome(t.key)).map((t) => t.key);
  const hrefs = Object.fromEntries(tabs.flatMap((t) => (t.href ? [[t.key, t.href]] : [])));
  sessionStorage.setItem(SESSION_KEY, JSON.stringify({ keys, pinned, hrefs }));
}

export function reconcileTabSession(userId: string): void {
  try {
    if (sessionStorage.getItem(SESSION_UID_KEY) !== userId) {
      sessionStorage.removeItem(SESSION_KEY);
      sessionStorage.setItem(SESSION_UID_KEY, userId);
    }
  } catch {
    // storage unavailable - same tolerance as readTabSession
  }
}

export function restoredTabs(
  navIndex: Map<string, NavNode>,
  current: TabRecord | null,
): TabRecord[] {
  const home = navIndex.get(AREA_PATH);
  const tabs = [home ? tabFromNav(home) : recordOf(AREA_PATH, "layout_dashboard_workbench")];
  const { keys, pinned, hrefs } = readTabSession();
  const pinnedSet = new Set(pinned);
  const restore = (key: string): TabRecord | null => {
    const node = navIndex.get(key);
    if (!node || node.children.length > 0 || isHome(key)) return null;
    return { ...tabFromNav(node), affix: pinnedSet.has(key), href: hrefs[key] };
  };
  const records = keys.map(restore).filter((r): r is TabRecord => r !== null);
  tabs.push(...records.filter((r) => r.affix), ...records.filter((r) => !r.affix));
  const currentRestorable =
    current !== null &&
    !isHome(current.key) &&
    !keys.includes(current.key) &&
    (() => {
      const node = navIndex.get(current.key);
      return node === undefined || node.children.length === 0;
    })();
  if (currentRestorable && current) tabs.push(current);
  return tabs;
}

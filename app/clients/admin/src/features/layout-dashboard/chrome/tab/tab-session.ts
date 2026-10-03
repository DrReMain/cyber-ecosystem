import type { LucideIcon } from "lucide-react";
import { m } from "#/paraglide/messages";
import { AREA_PATH } from "../../area";
import { isHome, type NavNode } from "../../protocol/nav";
import { compactParams, type MessageKey, type RouteMenuMeta } from "../../protocol/route-meta";

export interface TabRecord {
  affix: boolean;
  href?: string;
  icon?: LucideIcon;
  key: string;
  pattern: string;
  title: MessageKey;
  titleParams?: Record<string, string>;
}

const SESSION_KEY = "session_tabbar";
const SESSION_UID_KEY = "session_tabbar_uid";

export const stripSlash = (path: string) => path.replace(/\/$/, "") || "/";

const recordOf = (path: string, title: MessageKey, icon?: LucideIcon): TabRecord => ({
  affix: isHome(path),
  icon,
  key: path,
  pattern: path,
  title,
});

export function tabFromNav(node: NavNode): TabRecord {
  return recordOf(node.path, node.title, node.meta?.icon);
}

interface MatchLike {
  fullPath?: string;
  params?: Record<string, string | undefined>;
  pathname?: string;
  staticData?: { menu?: RouteMenuMeta; title?: MessageKey };
}

export function matchesAgreeWith(matches: ReadonlyArray<MatchLike>, pathname: string): boolean {
  const deepest = matches[matches.length - 1]?.pathname ?? "";
  return stripSlash(deepest) === stripSlash(pathname);
}

export function activeTabOf(matches: ReadonlyArray<MatchLike>, pathname: string): TabRecord | null {
  return matchesAgreeWith(matches, pathname) ? tabFromMatches(matches) : null;
}

export function tabFromMatches(matches: ReadonlyArray<MatchLike>): TabRecord | null {
  for (let i = matches.length - 1; i >= 0; i--) {
    const match = matches[i];
    const data = match?.staticData;
    if (match && data?.title) {
      const key = stripSlash(match.pathname ?? "/");
      const titleParams = compactParams(match.params);
      return {
        affix: isHome(key),
        icon: data.menu?.icon,
        key,
        pattern: stripSlash(match.fullPath ?? key),
        title: data.title,
        ...(titleParams && { titleParams }),
      };
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

interface SessionRecord {
  href?: string;
  key: string;
  pattern: string;
  pinned?: boolean;
  title: string;
  titleParams?: Record<string, string>;
}

interface TabSession {
  records: SessionRecord[];
}

const EMPTY_SESSION: TabSession = { records: [] };

const isParamRecord = (v: unknown): v is Record<string, string> =>
  typeof v === "object" &&
  v !== null &&
  !Array.isArray(v) &&
  Object.values(v).every((x) => typeof x === "string");

export function readTabSession(): TabSession {
  try {
    const parsed: unknown = JSON.parse(sessionStorage.getItem(SESSION_KEY) ?? "null");
    if (!parsed || typeof parsed !== "object") return EMPTY_SESSION;
    const { records } = parsed as { records?: unknown };
    if (!Array.isArray(records)) return EMPTY_SESSION; // legacy payload: degrade to empty
    const out: SessionRecord[] = [];
    for (const entry of records) {
      if (!entry || typeof entry !== "object") continue;
      const r = entry as Record<string, unknown>;
      if (
        typeof r.key !== "string" ||
        typeof r.pattern !== "string" ||
        typeof r.title !== "string"
      ) {
        continue;
      }
      out.push({
        key: r.key,
        pattern: r.pattern,
        title: r.title,
        ...(typeof r.href === "string" && { href: r.href }),
        ...(r.pinned === true && { pinned: true }),
        ...(isParamRecord(r.titleParams) && { titleParams: r.titleParams }),
      });
    }
    return { records: out };
  } catch {
    return EMPTY_SESSION;
  }
}

export function writeTabSession(tabs: readonly TabRecord[]): void {
  try {
    const records: SessionRecord[] = tabs.map((t) => ({
      key: t.key,
      pattern: t.pattern,
      title: t.title,
      ...(t.titleParams && { titleParams: t.titleParams }),
      ...(t.href && { href: t.href }),
      ...(t.affix && !isHome(t.key) && { pinned: true }),
    }));
    sessionStorage.setItem(SESSION_KEY, JSON.stringify({ records }));
  } catch {
    // storage unavailable - same tolerance as readTabSession
  }
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

// Hidden-route records (menu.hide, e.g. the profile group) restore entirely
// from the persisted payload: pattern absent from every index, title kept
// only if it is still a message key.
const hiddenRecord = (r: SessionRecord): TabRecord => ({
  affix: r.pinned === true,
  ...(r.href && { href: r.href }),
  key: r.key,
  pattern: r.pattern,
  title: r.title as MessageKey,
  ...(r.titleParams && { titleParams: r.titleParams }),
});

// Membership is rebuilt from storage + nav; hrefs are live per-key state and
// carried over from the previous list so a rebuild never discards a href the
// capture effect just wrote for the currently-shown tab.
export function restoredTabs(
  navIndex: Map<string, NavNode>,
  allIndex: Map<string, NavNode>,
  current: TabRecord | null,
  prev: readonly TabRecord[] = [],
): TabRecord[] {
  const homeNode = navIndex.get(AREA_PATH);
  const tabs: TabRecord[] =
    homeNode && homeNode.children.length === 0 ? [tabFromNav(homeNode)] : [];
  const { records } = readTabSession();
  const restore = (r: SessionRecord): TabRecord | null => {
    if (isHome(r.key)) return null;
    const node = navIndex.get(r.pattern);
    if (node) {
      if (node.children.length > 0) return null; // menu groups never become tabs
      // Title/icon re-derived from the live tree; instance key/params/href
      // from the payload.
      return {
        ...tabFromNav(node),
        key: r.key,
        ...(r.titleParams && { titleParams: r.titleParams }),
        ...(r.href && { href: r.href }),
        ...(r.pinned && { affix: true }),
      };
    }
    // A pattern present in the unfiltered index is a permission-dropped
    // visible route, not a hidden one; it must stay closed.
    if (allIndex.has(r.pattern)) return null;
    return r.title in m ? hiddenRecord(r) : null;
  };
  const restored = records.map(restore).filter((t): t is TabRecord => t !== null);
  tabs.push(...restored.filter((t) => t.affix), ...restored.filter((t) => !t.affix));
  const currentNode = current === null ? undefined : navIndex.get(current.pattern);
  if (
    current &&
    !isHome(current.key) &&
    (currentNode !== undefined || !allIndex.has(current.pattern)) &&
    !(currentNode && currentNode.children.length > 0) &&
    !tabs.some((t) => t.key === current.key)
  ) {
    tabs.push(current);
  }
  const liveHref = (key: string) => prev.find((t) => t.key === key)?.href;
  return tabs.map((t) => (t.href || !liveHref(t.key) ? t : { ...t, href: liveHref(t.key) }));
}

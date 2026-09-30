import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  MeasuringFrequency,
  MeasuringStrategy,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { restrictToHorizontalAxis } from "@dnd-kit/modifiers";
import { horizontalListSortingStrategy, SortableContext } from "@dnd-kit/sortable";
import { useMatches, useRouter, useRouterState } from "@tanstack/react-router";
import { Dropdown, type MenuProps } from "antd";
import clsx from "clsx";
import { useAtomValue, useSetAtom } from "jotai";
import {
  ArrowLeftToLine,
  ArrowRightLeft,
  ArrowRightToLine,
  ChevronsLeft,
  ChevronsRight,
  FoldHorizontal,
  Maximize2,
  Minimize2,
  MoreHorizontal,
  Pin,
  PinOff,
  RotateCw,
  X,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { dashboardPreferencesUiStore } from "#/stores/dashboard-preferences/ui-store";
import { isHome, navTo, useNavIndex } from "../../protocol/nav";
import { ScrollFade } from "../scroll-fade";
import { SortableTabChip } from "./sortable-tab-chip";
import { TabChip } from "./tab-chip";
import {
  closedBulk,
  closedTabs,
  movedTabs,
  pinnedTabs,
  restoredTabs,
  type TabRecord,
  tabFromMatches,
  unpinnedTabs,
  writeTabSession,
} from "./tab-session";
import { useIsomorphicLayoutEffect, useTabScroll } from "./use-tab-scroll";

const SCROLL_BTN =
  "flex h-full w-6.5 flex-none cursor-pointer items-center justify-center text-ink-tertiary hover:bg-fill-secondary";
const TOOL_BTN =
  "flex size-6.5 flex-none cursor-pointer items-center justify-center rounded-md text-ink-tertiary hover:bg-fill-secondary hover:text-ink";

const BULK_SCOPES = {
  "close-all": "all",
  "close-left": "left",
  "close-other": "other",
  "close-right": "right",
} as const;

const contextMenuItems = (
  tab: TabRecord,
  tabs: readonly TabRecord[],
  activeKey: string,
): NonNullable<MenuProps["items"]> => {
  const closable = (t: TabRecord) => !t.affix;
  const idx = tabs.findIndex((t) => t.key === tab.key);
  const leftHas = tabs.slice(0, Math.max(idx, 0)).some(closable);
  const rightHas = idx >= 0 ? tabs.slice(idx + 1).some(closable) : false;
  const otherHas = tabs.filter((t) => t.key !== tab.key).some(closable);
  return [
    {
      disabled: !closable(tab) || tabs.length <= 1,
      icon: <X aria-hidden className="size-3.5" />,
      key: "close",
      label: m.layout_dashboard_tab_close(),
    },
    isHome(tab.key)
      ? {
          disabled: true,
          icon: <Pin aria-hidden className="size-3.5" />,
          key: "pin",
          label: m.layout_dashboard_tab_pin(),
        }
      : tab.affix
        ? {
            icon: <PinOff aria-hidden className="size-3.5" />,
            key: "unpin",
            label: m.layout_dashboard_tab_unpin(),
          }
        : {
            icon: <Pin aria-hidden className="size-3.5" />,
            key: "pin",
            label: m.layout_dashboard_tab_pin(),
          },
    {
      disabled: tab.key !== activeKey,
      icon: <RotateCw aria-hidden className="size-3.5" />,
      key: "refresh",
      label: m.layout_dashboard_tab_refresh(),
    },
    { type: "divider" } as const,
    {
      disabled: !leftHas,
      icon: <ArrowLeftToLine aria-hidden className="size-3.5 rtl:-scale-x-100" />,
      key: "close-left",
      label: m.layout_dashboard_tabs_close_left(),
    },
    {
      disabled: !rightHas,
      icon: <ArrowRightToLine aria-hidden className="size-3.5 rtl:-scale-x-100" />,
      key: "close-right",
      label: m.layout_dashboard_tabs_close_right(),
    },
    {
      disabled: !otherHas,
      icon: <FoldHorizontal aria-hidden className="size-3.5" />,
      key: "close-other",
      label: m.layout_dashboard_tabs_close_other(),
    },
    {
      disabled: !tabs.some(closable),
      icon: <ArrowRightLeft aria-hidden className="size-3.5 rtl:-scale-x-100" />,
      key: "close-all",
      label: m.layout_dashboard_tabs_close_all(),
    },
  ];
};

interface TabbarProps {
  onRefresh: () => void;
  scrolled?: boolean;
}

export function Tabbar({ onRefresh, scrolled }: Readonly<TabbarProps>) {
  const router = useRouter();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const href = useRouterState({ select: (s) => s.location.href });
  const matches = useMatches();
  const { maximized } = useAtomValue(dashboardPreferencesUiStore.atom);
  const setUi = useSetAtom(dashboardPreferencesUiStore.atom);
  const [tabs, setTabs] = useState<TabRecord[]>([]);
  const [ready, setReady] = useState(false);
  const navIndex = useNavIndex();

  useEffect(() => {
    if (!ready || tabs.length === 0) return;
    writeTabSession(tabs);
  }, [ready, tabs]);

  useEffect(() => {
    setTabs((prev) =>
      prev.some((t) => t.key === pathname && t.href !== href)
        ? prev.map((t) => (t.key === pathname ? { ...t, href } : t))
        : prev,
    );
  }, [href, pathname]);

  const navKey = useMemo(() => [...navIndex.keys()].sort().join("|"), [navIndex]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: navIndex is read fresh inside; the dep is its content digest
  useEffect(() => {
    setTabs(restoredTabs(navIndex, tabFromMatches(matches)));
    setReady(true);
  }, [navKey]);

  useIsomorphicLayoutEffect(() => {
    const record = tabFromMatches(matches);
    if (!record || record.key !== pathname) return;
    const node = navIndex.get(record.key);
    if (node && node.children.length > 0) return;
    setTabs((prev) => (prev.some((t) => t.key === record.key) ? prev : [...prev, record]));
  }, [matches, pathname, navIndex]);

  const hrefOf = (key: string | null): string | null => {
    if (key === null) return null;
    return tabs.find((t) => t.key === key)?.href ?? key;
  };

  const close = (key: string) => {
    const result = closedTabs(tabs, key);
    if (!result) return;
    setTabs(result.tabs);
    const fallback = hrefOf(result.fallback);
    if (key === pathname && fallback) navTo(router, fallback);
    setPendingFocus(key === pathname ? result.fallback : pathname);
  };

  const runBulk = (action: string, tab: TabRecord) => {
    const scope = BULK_SCOPES[action as keyof typeof BULK_SCOPES];
    if (!scope) return;
    const result = closedBulk(tabs, tab.key, scope, pathname);
    if (!result) return;
    setTabs(result.tabs);
    const fallback = hrefOf(result.fallback);
    if (fallback && result.fallback !== pathname) {
      navTo(router, fallback);
    }
    setPendingFocus(result.fallback ?? pathname);
  };

  const runAction = (action: string, tab: TabRecord) => {
    if (action === "close") return close(tab.key);
    if (action === "refresh") return onRefresh();
    if (action === "pin") {
      setTabs((prev) => pinnedTabs(prev, tab.key) ?? prev);
      return;
    }
    if (action === "unpin") {
      setTabs((prev) => unpinnedTabs(prev, tab.key) ?? prev);
      return;
    }
    runBulk(action, tab);
  };

  const { atEnd, atStart, overflow, ref, scroll } = useTabScroll({
    activeKey: pathname,
    count: tabs.length,
  });

  const activeRecord = tabFromMatches(matches);
  const activeMenu = activeRecord
    ? (() => {
        const tab = tabs.find((t) => t.key === activeRecord.key) ?? activeRecord;
        return {
          items: contextMenuItems(tab, tabs, activeRecord.key),
          onClick: ({ key }: { key: string }) => runAction(key, tab),
        };
      })()
    : undefined;

  const didDragRef = useRef(false);
  const releaseDragFlag = () => {
    setTimeout(() => {
      didDragRef.current = false;
    }, 0);
  };
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 4 } }));
  const affixTabs = useMemo(() => tabs.filter((tab) => tab.affix), [tabs]);
  const liveTabs = useMemo(() => tabs.filter((tab) => !tab.affix), [tabs]);

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    releaseDragFlag();
    if (!over || active.id === over.id) return;
    const next = movedTabs(tabs, String(active.id), String(over.id));
    if (next) setTabs(next);
  };

  const [pendingFocus, setPendingFocus] = useState<string | null>(null);
  useIsomorphicLayoutEffect(() => {
    if (pendingFocus === null) return;
    ref.current
      ?.querySelector<HTMLElement>(`[data-tab-key="${CSS.escape(pendingFocus)}"]`)
      ?.focus();
    setPendingFocus(null);
  }, [pendingFocus, ref, tabs]);

  const chipProps = (tab: TabRecord) => ({
    isActive: tab.key === pathname,
    menu: {
      items: contextMenuItems(tab, tabs, pathname),
      onClick: ({ key }: { key: string }) => runAction(key, tab),
    },
    onActivate: () => {
      if (didDragRef.current) {
        didDragRef.current = false;
        return;
      }
      if (tab.key !== pathname) navTo(router, tab.href ?? tab.key);
    },
    onClose: () => close(tab.key),
    siblingCount: tabs.length,
    tab,
  });

  return (
    <div
      className={clsx(
        "sticky z-30 flex h-9.5 flex-none items-center border-line-soft border-b bg-container",
        maximized ? "top-0" : "top-14",
      )}
    >
      <style>{`
@keyframes nc-sheen-drift {
  from { background-position: 0% 0; }
  to { background-position: 200% 0; }
}
.nc-sheen {
  background: linear-gradient(100deg,
    rgba(192, 132, 252, 0.15) 0%,
    rgba(96, 165, 250, 0.16) 25%,
    rgba(34, 211, 238, 0.14) 50%,
    rgba(52, 211, 153, 0.14) 75%,
    rgba(192, 132, 252, 0.15) 100%);
  background-size: 300% 100%;
  animation: nc-sheen-drift 9s linear infinite;
}
@media (prefers-reduced-motion: reduce) {
  .nc-sheen { animation: none; }
}`}</style>
      <span
        aria-hidden
        className={clsx(
          "pointer-events-none absolute inset-x-0 top-[calc(100%+1px)] h-10 bg-linear-to-b from-container/85 via-container/40 opacity-0 transition-opacity",
          scrolled && "opacity-100",
        )}
      />
      {overflow ? (
        <button
          aria-label={m.layout_dashboard_tabs_scroll_start()}
          className={clsx(SCROLL_BTN, atStart && "pointer-events-none opacity-30")}
          onClick={() => scroll(-1)}
          type="button"
        >
          <ChevronsLeft aria-hidden className="size-3.5 rtl:-scale-x-100" />
        </button>
      ) : null}
      <div className="relative h-full min-w-0 flex-1">
        {ready ? null : <div aria-hidden className="nc-sheen h-full w-full rounded-[3px]" />}
        <div
          aria-label={m.layout_dashboard_tabs()}
          className={clsx(
            "flex h-full items-center overflow-x-auto overscroll-x-contain [scrollbar-width:none] [&::-webkit-scrollbar]:hidden",
            !ready && "invisible",
          )}
          onKeyDown={(e) => {
            if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;
            const flip = getTextDirection() === "rtl" ? -1 : 1;
            const step = (e.key === "ArrowRight" ? 1 : -1) * flip;
            const items = [...e.currentTarget.querySelectorAll<HTMLElement>('[role="tab"]')];
            const idx = items.indexOf(document.activeElement as HTMLElement);
            const next = items[idx + step];
            if (next) {
              e.preventDefault();
              next.focus();
            }
          }}
          ref={ref}
          role="tablist"
        >
          <DndContext
            collisionDetection={closestCenter}
            measuring={{
              droppable: {
                frequency: MeasuringFrequency.Optimized,
                strategy: MeasuringStrategy.Always,
              },
            }}
            modifiers={[restrictToHorizontalAxis]}
            onDragCancel={releaseDragFlag}
            onDragEnd={handleDragEnd}
            onDragStart={() => {
              didDragRef.current = true;
            }}
            sensors={sensors}
          >
            {affixTabs.map((tab) => (
              <TabChip key={tab.key} {...chipProps(tab)} />
            ))}
            <SortableContext
              items={liveTabs.map((tab) => tab.key)}
              strategy={horizontalListSortingStrategy}
            >
              {liveTabs.map((tab) => (
                <SortableTabChip key={tab.key} {...chipProps(tab)} />
              ))}
            </SortableContext>
          </DndContext>
        </div>
        <ScrollFade axis="x" show={overflow && !atStart} side="start" />
        <ScrollFade axis="x" show={overflow && !atEnd} side="end" />
      </div>
      {overflow ? (
        <button
          aria-label={m.layout_dashboard_tabs_scroll_end()}
          className={clsx(SCROLL_BTN, atEnd && "pointer-events-none opacity-30")}
          onClick={() => scroll(1)}
          type="button"
        >
          <ChevronsRight aria-hidden className="size-3.5 rtl:-scale-x-100" />
        </button>
      ) : null}

      <div className="flex h-full flex-none items-center gap-0.5 border-line-soft border-s px-1">
        <Dropdown disabled={!activeMenu} menu={activeMenu ?? { items: [] }} trigger={["click"]}>
          <button
            aria-label={m.layout_dashboard_tab_more()}
            className={clsx(TOOL_BTN, "disabled:cursor-not-allowed disabled:opacity-40")}
            disabled={!activeMenu}
            type="button"
          >
            <MoreHorizontal aria-hidden className="size-3.5" />
          </button>
        </Dropdown>
        <button
          aria-label={m.layout_dashboard_tab_refresh()}
          className={TOOL_BTN}
          onClick={onRefresh}
          type="button"
        >
          <RotateCw aria-hidden className="size-3.5" />
        </button>
        <button
          aria-label={
            maximized ? m.layout_dashboard_tab_restore() : m.layout_dashboard_tab_maximize()
          }
          className={TOOL_BTN}
          onClick={() =>
            setUi((d) => {
              d.maximized = !d.maximized;
            })
          }
          type="button"
        >
          {maximized ? (
            <Minimize2 aria-hidden className="size-3.5" />
          ) : (
            <Maximize2 aria-hidden className="size-3.5" />
          )}
        </button>
      </div>
    </div>
  );
}

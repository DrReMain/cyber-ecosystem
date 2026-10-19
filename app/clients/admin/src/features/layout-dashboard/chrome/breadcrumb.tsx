import { useMatches, useRouter } from "@tanstack/react-router";
import clsx from "clsx";
import { LayoutDashboard } from "lucide-react";
import { useMemo } from "react";
import { AREA_PATH } from "../area";
import { firstLeafOf, navTo, text, useNavIndex } from "../protocol/nav";
import { compactParams, type MessageKey } from "../protocol/route-meta";

interface Crumb {
  params?: Record<string, string>;
  title: MessageKey;
  path: string;
}

const LABEL = "flex h-7 items-center font-medium font-sans text-[12px] leading-none";
const TIP =
  "[clip-path:polygon(0_0,100%_50%,0_100%)] rtl:[clip-path:polygon(100%_0,0_50%,100%_100%)]";
const NOTCH =
  "[clip-path:polygon(0_0,100%_0,100%_100%,0_100%,50%_50%)] rtl:[clip-path:polygon(100%_0,0_0,0_100%,100%_100%,50%_50%)]";

const capFor = (isFirst: boolean, isLast: boolean) =>
  clsx(
    isFirst && "rounded-ss-[4px] rounded-es-[4px]",
    isLast && "rounded-se-[4px] rounded-ee-[4px]",
  );

const appendages = (isFirst: boolean, isLast: boolean) => (
  <>
    {isFirst ? null : (
      <span
        aria-hidden
        className={clsx("absolute -inset-s-7 inset-y-0 w-7.25 bg-inherit", NOTCH)}
      />
    )}
    {isLast ? null : (
      <span
        aria-hidden
        className={clsx("absolute -inset-e-3.5 inset-y-0 w-3.75 bg-inherit", TIP)}
      />
    )}
  </>
);

export function Breadcrumb() {
  const router = useRouter();
  const matches = useMatches();

  const crumbs = useMemo<Crumb[]>(() => {
    const trail: Crumb[] = [{ title: "layout_dashboard_workbench", path: AREA_PATH }];
    for (const mt of matches) {
      const title = mt.staticData.title;
      const path = mt.pathname.replace(/\/$/, "");
      if (title && path.startsWith(`${AREA_PATH}/`)) {
        trail.push({ params: compactParams(mt.params), title, path });
      }
    }
    return trail;
  }, [matches]);

  const navIndex = useNavIndex();
  const targets = useMemo(() => {
    const resolve = (path: string) => {
      const node = navIndex.get(path);
      return node ? (firstLeafOf([node])?.path ?? path) : path;
    };
    return new Map(crumbs.map((c) => [c.path, resolve(c.path)]));
  }, [navIndex, crumbs]);

  if (crumbs.length <= 1) return null;

  const label = (crumb: Crumb) => text(crumb.title, crumb.params);

  return (
    <nav aria-label="breadcrumb" className="hidden items-center leading-none lg:flex">
      <ol className="flex items-center gap-9">
        {crumbs.map((crumb, i) => {
          const isLast = i === crumbs.length - 1;
          const isFirst = i === 0;
          return (
            <li
              aria-current={isLast ? "page" : undefined}
              className={clsx(
                "relative flex h-7 select-none items-center gap-1.5 bg-layout text-ink-secondary",
                capFor(isFirst, isLast),
                isFirst ? "ps-3.75" : "ps-1.25",
                isLast ? "pe-3.75" : "pe-2",
                isLast ? "cursor-default" : "hover:bg-elevated",
              )}
              key={crumb.path}
            >
              {appendages(isFirst, isLast)}
              {isFirst ? <LayoutDashboard aria-hidden className="size-3 flex-none" /> : null}
              {isLast ? (
                <span className={LABEL}>{label(crumb)}</span>
              ) : (
                <button
                  className={clsx(
                    LABEL,
                    "cursor-pointer text-start outline-none focus-visible:bg-fill-secondary",
                  )}
                  onClick={() => navTo(router, targets.get(crumb.path) ?? crumb.path)}
                  type="button"
                >
                  {label(crumb)}
                </button>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}

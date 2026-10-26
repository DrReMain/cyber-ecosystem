import clsx from "clsx";
import type { ReactNode } from "react";
import { AmbientBackdrop, type AmbientOrb } from "./ambient-backdrop";

const maxWidthCls = {
  md: "max-w-md",
  lg: "max-w-lg",
  xl: "max-w-xl",
  "2xl": "max-w-2xl",
} as const;

export type AmbientMaxWidth = keyof typeof maxWidthCls;

interface AmbientLayoutProps {
  children: ReactNode;
  maxWidth?: AmbientMaxWidth;
  /**
   * "page" — the document grows and scrolls (default).
   * "inside" — multi-step flows: the viewport is height-capped and the
   * content column scrolls within it, keeping ambient chrome fixed.
   */
  scroll?: "page" | "inside";
  /** Login-type tight screens: less outer padding, content vertically
   * centered with an inner scroll fallback — same height cap as "inside". */
  dense?: boolean;
  /** Per-page orb art direction; absent = the shared login trio. */
  orbs?: readonly AmbientOrb[];
  /** Frame-level chrome (the tab bar): the frame's own flow child, below
   * the scrolling column — the desktop frame docks its bar to the frame
   * bottom, never to the scroll extent (an in-scroller absolute would
   * ride below the fold mid-scroll). */
  chrome?: ReactNode;
}

export function AmbientLayout({
  children,
  maxWidth = "md",
  scroll = "page",
  dense = false,
  orbs,
  chrome,
}: Readonly<AmbientLayoutProps>) {
  const capped = scroll === "inside" || dense;
  return (
    <div
      className={clsx(
        "safe-pb safe-pt bg-canvas text-ink relative isolate flex flex-col overflow-hidden",
        capped ? "h-app" : "min-h-app",
        // ≥md the app presents as a bounded H5 frame, centered on the page:
        // a phone-height canvas with breathing room, content scrolling inside
        // the frame and chrome docking to the frame — not the viewport.
        "md:h-frame md:border-line-soft md:mx-auto md:my-auto md:min-h-0 md:w-full md:max-w-md md:rounded-3xl md:border md:pb-0 md:shadow-2xl",
      )}
    >
      <AmbientBackdrop orbs={orbs} />
      <div
        className={clsx(
          "relative z-[1] flex flex-1 items-center justify-center",
          dense
            ? "min-h-0 flex-col overflow-y-auto px-3 py-4 sm:p-6"
            : clsx(
                "flex-col p-4 sm:p-6",
                scroll === "inside" && "min-h-0 overflow-y-auto",
                // Desktop frame scroll: justify-center clips a overflowing
                // column's top, so centering moves to the column's own
                // my-auto and the container scrolls from the start edge.
                "md:min-h-0 md:justify-start md:overflow-y-auto",
              ),
        )}
      >
        <div
          className={clsx(
            "mx-auto w-full md:my-auto",
            maxWidthCls[maxWidth],
            dense && scroll !== "inside" && "shrink-0",
            scroll === "inside" && "flex min-h-0 flex-col",
          )}
        >
          {children}
        </div>
      </div>
      {chrome !== undefined && <div className="relative z-[1] shrink-0">{chrome}</div>}
    </div>
  );
}

import clsx from "clsx";
import type { ReactNode } from "react";
import { AmbientBackdrop } from "./ambient-backdrop";

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
}

export function AmbientLayout({
  children,
  maxWidth = "md",
  scroll = "page",
  dense = false,
}: Readonly<AmbientLayoutProps>) {
  const capped = scroll === "inside" || dense;
  return (
    <div
      className={clsx(
        "safe-pb safe-pt bg-canvas text-ink relative isolate flex flex-col overflow-hidden",
        capped ? "h-app" : "min-h-app",
      )}
    >
      <AmbientBackdrop />
      <div
        className={clsx(
          "relative z-[1] flex flex-1 items-center justify-center",
          dense
            ? "min-h-0 flex-col overflow-y-auto px-3 py-4 sm:p-6"
            : clsx("flex-col p-4 sm:p-6", scroll === "inside" && "min-h-0 overflow-y-auto"),
        )}
      >
        <div
          className={clsx(
            "mx-auto w-full",
            maxWidthCls[maxWidth],
            dense && scroll !== "inside" && "shrink-0",
            scroll === "inside" && "flex min-h-0 flex-col",
          )}
        >
          {children}
        </div>
      </div>
    </div>
  );
}

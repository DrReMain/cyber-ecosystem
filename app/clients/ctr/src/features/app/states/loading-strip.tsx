import clsx from "clsx";

type LoadingStripTone = "inherit" | "ink";

/** Class triple per tone: mono label, bar track, sweep highlight. */
const TONE_CLASSES: Record<LoadingStripTone, { label: string; track: string; sweep: string }> = {
  inherit: {
    label: "text-current/50",
    track: "bg-current/12",
    sweep: "via-current/70",
  },
  ink: {
    label: "text-ink-tertiary",
    track: "bg-ink/12",
    sweep: "via-ink/70",
  },
};

interface LoadingStripProps {
  /** "inherit" rides the ambient text color (route fallbacks on arbitrary
   *  surfaces); "ink" pins the page-ink tokens (page-scale state screens). */
  tone: LoadingStripTone;
}

/**
 * The mono LOADING strip both loading scales share — PendingFallback
 * (route-pending) and StateScreen's loading variant (page-scale) render it
 * so the two read as one system. Decorative by design: the label is English
 * mono like the ERR chips, `dir="ltr"` keeps the sweep anchored, and callers
 * carry the accessible label (sr-only) under their own role="status".
 */
export function LoadingStrip({ tone }: Readonly<LoadingStripProps>) {
  const classes = TONE_CLASSES[tone];

  return (
    <div aria-hidden className={"justify-center"} dir="ltr">
      <div
        className={clsx(
          "text-center font-mono text-[11px] uppercase tracking-[0.18em]",
          classes.label,
        )}
      >
        LOADING
      </div>
      <span className={clsx("relative mt-2 block h-0.5 overflow-hidden rounded", classes.track)}>
        <span
          className={clsx(
            "absolute inset-s-[-40%] inset-y-0 w-2/5 bg-linear-to-r from-transparent to-transparent motion-safe:animate-sweep rtl:motion-safe:animate-sweep-rev",
            classes.sweep,
          )}
        />
      </span>
    </div>
  );
}

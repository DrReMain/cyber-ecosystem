import clsx from "clsx";
import { Inbox } from "lucide-react";
import type { ReactNode } from "react";
import { m } from "#/paraglide/messages";

type StateScreenVariant = "empty" | "loading";

interface StateScreenProps {
  variant: StateScreenVariant;
  /** Overrides the variant's default copy. */
  title?: string;
  /** Empty defaults to the generic empty description; loading has none. */
  description?: string;
  /** Optional action slot (retry link, primary button) below the copy. */
  action?: ReactNode;
}

/**
 * Page-scale state presentation for the two quiet states a content area
 * renders itself into: "empty" (no data yet) and "loading" (fetching).
 * Errors stay with the error family (ErrorPage / ErrorFallback); the
 * loading variant keeps the PendingFallback lineage (mono label + sweep
 * bar) so both loading scales read as one system.
 */
export function StateScreen({ variant, title, description, action }: Readonly<StateScreenProps>) {
  const loading = variant === "loading";
  const heading = title ?? (loading ? m.common_loading() : m.common_state_empty_title());
  const body = description ?? (loading ? undefined : m.common_state_empty_desc());

  return (
    <div
      className="flex w-full flex-col items-center justify-center px-6 py-12 text-center"
      role="status"
    >
      {loading ? (
        <div className="w-[min(200px,60%)]" dir="ltr">
          <div className="text-center font-mono text-[11px] text-ink-tertiary uppercase tracking-[0.18em]">
            LOADING
          </div>
          <span className="relative mt-2 block h-0.5 overflow-hidden rounded bg-ink/12">
            <span className="absolute inset-s-[-40%] inset-y-0 w-2/5 bg-linear-to-r from-transparent via-ink/70 to-transparent motion-safe:animate-sweep rtl:motion-safe:animate-sweep-rev" />
          </span>
        </div>
      ) : (
        <span className="flex size-14 items-center justify-center rounded-card border border-line-soft bg-surface-content text-ink-tertiary">
          <Inbox aria-hidden className="size-6" strokeWidth={1.5} />
        </span>
      )}

      <h2
        className={clsx(
          "text-balance font-semibold text-ink tracking-[0.02em] rtl:tracking-normal",
          loading ? "mt-5 text-base text-ink-secondary" : "mt-5 text-lg",
        )}
      >
        {heading}
      </h2>
      {body ? (
        <p className="mt-2 max-w-sm text-balance text-ink-secondary text-sm leading-relaxed">
          {body}
        </p>
      ) : null}
      {action ? <div className="mt-7">{action}</div> : null}
    </div>
  );
}

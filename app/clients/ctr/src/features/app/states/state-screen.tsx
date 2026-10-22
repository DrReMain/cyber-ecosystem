import clsx from "clsx";
import { Inbox } from "lucide-react";
import type { ReactNode } from "react";
import { m } from "#/paraglide/messages";
import { LoadingStrip } from "./loading-strip";

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
      className="flex w-full grow flex-col items-center justify-center px-6 py-12 text-center"
      role="status"
    >
      {loading ? (
        <>
          {/* The localized label is what assistive tech reads under the
              container's role="status"; the strip itself is decorative. */}
          <span className="sr-only">{m.common_loading()}</span>
          <LoadingStrip tone="ink" />
        </>
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

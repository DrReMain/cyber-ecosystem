import { ChevronLeft } from "lucide-react";
import type { ReactNode } from "react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";

const TOTAL_STEPS = 3;
// Static segment positions double as keys (an index key would trip the
// array-index lint; the values are the identity here).
const PROGRESS_SEGS = [1, 2, 3] as const;

interface StepFrameProps {
  step: 1 | 2 | 3;
  onBack: () => void;
  children: ReactNode;
}

/**
 * The shared chrome of the three onboarding steps: back fab, step counter,
 * and the three-segment progress rail, over the ambient inside-scroll shape
 * (viewport-capped so the ambient chrome stays fixed while the column
 * scrolls).
 */
export function StepFrame({ step, onBack, children }: Readonly<StepFrameProps>) {
  return (
    <AmbientLayout scroll="inside">
      <div className="flex w-full flex-col gap-3">
        <div className="flex items-center gap-3">
          <button
            aria-label={m.common_back()}
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border border-line text-ink-secondary active:opacity-60"
            onClick={onBack}
            type="button"
          >
            <ChevronLeft aria-hidden size={18} />
          </button>
          <h1 className="font-extrabold text-ink text-lg">{m.profile_setup_title()}</h1>
          <span className="ml-auto text-ink-tertiary text-xs">
            {m.profile_setup_step({ n: step, total: TOTAL_STEPS })}
          </span>
        </div>
        <div className="flex gap-1.5">
          {PROGRESS_SEGS.map((seg) => (
            <span
              className={`h-1 flex-1 rounded-full ${
                seg < step ? "bg-focus" : seg === step ? "bg-cta-primary" : "bg-line"
              }`}
              key={seg}
            />
          ))}
        </div>
        <div className="flex min-h-0 flex-col gap-3 py-2">{children}</div>
      </div>
    </AmbientLayout>
  );
}

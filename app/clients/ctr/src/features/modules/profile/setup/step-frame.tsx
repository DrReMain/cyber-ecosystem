import { ChevronLeft } from "lucide-react";
import type { ReactNode } from "react";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";

const TOTAL_STEPS = 3;
// Static segment positions double as keys (an index key would trip the
// array-index lint; the values are the identity here).
const PROGRESS_SEGS = [1, 2, 3] as const;

/* Per-step orb art direction, straight off the prototype's three wizard
 * pages: nickname carries two, age and games one each. */
const STEP_ORBS = {
  1: [
    { className: "top-[-60px] right-[-70px] size-[min(62vw,240px)]", tone: "blue" },
    { className: "bottom-[-30px] left-[-50px] size-[min(52vw,200px)]", tone: "cyan" },
  ],
  2: [{ className: "top-[-60px] left-[-60px] size-[min(57vw,220px)]", tone: "purple" }],
  3: [{ className: "top-[-60px] right-[-60px] size-[min(57vw,220px)]", tone: "cyan" }],
} as const satisfies Record<1 | 2 | 3, readonly AmbientOrb[]>;

interface StepFrameProps {
  step: 1 | 2 | 3;
  onBack: () => void;
  children: ReactNode;
}

/**
 * The shared chrome of the three onboarding steps: back fab, step counter,
 * and the three-segment progress rail (§4.7 tri-state — current amber with
 * glow, done cyan→blue, upcoming surface), over the ambient inside-scroll
 * shape (viewport-capped so the ambient chrome stays fixed while the column
 * scrolls).
 */
export function StepFrame({ step, onBack, children }: Readonly<StepFrameProps>) {
  return (
    <AmbientLayout orbs={STEP_ORBS[step]} scroll="inside">
      <div className="flex w-full flex-col gap-3">
        <div className="flex items-center gap-3">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={onBack}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{m.profile_setup_title()}</h1>
          <span className="text-ink-tertiary ml-auto text-xs">
            {m.profile_setup_step({ n: step, total: TOTAL_STEPS })}
          </span>
        </div>
        <div className="flex gap-[5px]">
          {PROGRESS_SEGS.map((seg) => (
            <span
              className={`h-[5px] flex-1 rounded-full ${
                seg < step
                  ? "bg-progress-done"
                  : seg === step
                    ? "bg-progress-cur shadow-progress-cur"
                    : "bg-surface-input"
              }`}
              key={seg}
            />
          ))}
        </div>
        <div className="flex min-h-0 flex-col gap-3">{children}</div>
      </div>
    </AmbientLayout>
  );
}

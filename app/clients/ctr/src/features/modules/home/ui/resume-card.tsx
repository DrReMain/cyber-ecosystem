import { m } from "#/paraglide/messages";
import { resumeDescVars, type WireStamp } from "../resume";

/** The resume card: the peeked session's progress line and the jump
 * back into /match (the navigation itself stays with the page). */
export function ResumeCard({
  createdAt,
  done,
  onOpen,
  total,
}: Readonly<{
  createdAt: WireStamp;
  done: number;
  onOpen: () => void;
  total: number;
}>) {
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card flex items-center gap-3 border p-3.5 backdrop-blur-md">
      <div className="min-w-0 flex-1">
        <div className="text-ink text-[13.5px] font-extrabold">{m.home_resume_title()}</div>
        <div className="text-ink-tertiary mt-0.75 text-[11.5px] leading-[1.7]">
          {m.home_resume_desc(resumeDescVars(createdAt, done, total))}
        </div>
      </div>
      <button
        className="bg-cta-practice shadow-cta-practice rounded-control shrink-0 px-3.5 py-2 text-[12.5px] font-bold text-white active:opacity-80"
        onClick={onOpen}
        type="button"
      >
        {m.home_resume_cta()}
      </button>
    </div>
  );
}

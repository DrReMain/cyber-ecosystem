import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { m } from "#/paraglide/messages";

/* The resume card's desc vars: the progress counts plus the age face —
 * the prototype's inclusive day count (started today reads 已过 1 天) and
 * a local-date stamp. Queries deliver proto-JSON shapes (CONVENTIONS
 * §1): createdAt arrives as its RFC 3339 wire string, never the message
 * object the typed surface claims — routing it through timestampDate
 * reads .seconds off a string (undefined → NaN) and the stamp renders
 * NaN-NaN-NaN. The wire string is the one read mechanism. */
function resumeDescVars(createdAt: Timestamp | undefined, done: number, total: number) {
  const wire = createdAt as unknown as string | undefined;
  const created = wire !== undefined ? Date.parse(wire) : Date.now();
  const days = Math.max(1, Math.floor((Date.now() - created) / 86_400_000) + 1);
  const d = new Date(created);
  const pad = (n: number) => String(n).padStart(2, "0");
  const date = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  return { done, total, date, days };
}

/** The resume card: the peeked session's progress line and the jump
 * back into /match (the navigation itself stays with the page). */
export function ResumeCard({
  createdAt,
  done,
  onOpen,
  total,
}: Readonly<{
  createdAt: Timestamp | undefined;
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

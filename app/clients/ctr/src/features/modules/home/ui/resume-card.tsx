import { type Timestamp, timestampDate } from "@bufbuild/protobuf/wkt";
import { m } from "#/paraglide/messages";

/* The resume card's desc vars: the progress counts plus the age face —
 * the prototype's inclusive day count (started today reads 已过 1 天) and
 * a local-date stamp; createdAt is the proto Timestamp itself (the typed
 * connect surface hands the message object, not a flattened string). */
function resumeDescVars(createdAt: Timestamp | undefined, done: number, total: number) {
  const created = createdAt !== undefined ? timestampDate(createdAt).getTime() : Date.now();
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
    <div className="border-line bg-surface-card shadow-card rounded-card flex items-center gap-3 border p-3.5">
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

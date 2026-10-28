import { useNavigate } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";
import { type HistorySlot, historyDescOf } from "../overview";

// One genre's history row: the latest session line's progress words and
// its action — never played offers the way in, an unfinished session the
// resume entry, a finished one its own read-back face (the anchor's
// report when it scored, the session-done page's honest no-scores face
// otherwise).
function HistoryRow({ slot }: Readonly<{ slot: HistorySlot }>) {
  const navigate = useNavigate();
  const line = slot.line;
  const desc = historyDescOf(slot);
  const openRead =
    line === undefined || !line.finished
      ? undefined
      : line.anchorRecordId !== undefined
        ? ({ to: "/report/$recordId", params: { recordId: line.anchorRecordId } } as const)
        : ({ to: "/report/session/$sessionId", params: { sessionId: line.sessionId } } as const);
  return (
    <div className="flex items-center justify-between border-b border-[rgba(55,65,81,0.4)] py-[9px] last:border-b-0">
      <div className="flex min-w-0 items-center gap-2.5">
        <span
          aria-hidden
          className={`flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-[9px] text-[9px] font-bold text-white ${
            slot.genre === "fps" ? "bg-[rgba(79,70,229,0.25)]" : "bg-[rgba(29,78,216,0.25)]"
          }`}
        >
          {slot.genre.toUpperCase()}
        </span>
        <div className="min-w-0">
          <div
            className={`text-[12.5px] font-extrabold ${line === undefined ? "text-[#9ca3af]" : "text-ink"}`}
          >
            {slot.genre === "fps" ? m.profile_hist_fps() : m.profile_hist_moba()}
          </div>
          <div className="text-ink-tertiary mt-0.5 text-[10.5px]">{desc}</div>
        </div>
      </div>
      {line === undefined ? (
        <button
          className="text-ink-secondary tap-zone rounded-control shrink-0 border-[1.5px] border-[#4b5563] px-3 py-1.5 text-[11.5px] font-medium active:opacity-60"
          onClick={() => void navigate({ to: "/match" })}
          type="button"
        >
          {m.profile_hist_go()}
        </button>
      ) : !line.finished ? (
        <button
          className="bg-cta-practice shadow-cta-practice tap-zone rounded-control shrink-0 px-3 py-1.5 text-[11.5px] font-bold text-white active:opacity-80"
          onClick={() => void navigate({ to: "/match" })}
          type="button"
        >
          {m.profile_hist_resume()}
        </button>
      ) : (
        <button
          className="text-ink-secondary tap-zone rounded-control shrink-0 border-[1.5px] border-[#4b5563] px-3 py-1.5 text-[11.5px] font-medium active:opacity-60"
          onClick={() => openRead !== undefined && void navigate(openRead)}
          type="button"
        >
          {m.profile_hist_report()}
        </button>
      )}
    </div>
  );
}

/** The assessment history (prototype profile-overview ⑤): one row per
 * seeded genre with its latest session line — the digest's truth, never
 * a fabricated session. */
export function HistoryCard({ slots }: Readonly<{ slots: readonly HistorySlot[] }>) {
  return (
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
      <div className="text-ink mb-[11px] text-[14.5px] font-extrabold">
        {m.profile_hist_title()}
      </div>
      {slots.map((slot) => (
        <HistoryRow key={slot.genre} slot={slot} />
      ))}
    </div>
  );
}

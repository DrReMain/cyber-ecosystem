import { m } from "#/paraglide/messages";

/** The retake receipt chip (prototype .st): grey while chances remain,
 * amber when spent. */
export function RetakeChip({ left, max }: Readonly<{ left: number; max: number }>) {
  const spent = left <= 0;
  return (
    <span
      className={`rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold ${
        spent ? "text-status-pending bg-status-pending-bg" : "text-ink-muted bg-status-lock-bg"
      }`}
    >
      {m.test_result_retake_chip({ left, max })}
    </span>
  );
}

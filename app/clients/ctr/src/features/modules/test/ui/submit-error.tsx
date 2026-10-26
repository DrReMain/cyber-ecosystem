import { m } from "#/paraglide/messages";

/** The submit failure's face: the error word with its inline retry. */
export function SubmitError({ onResubmit }: Readonly<{ onResubmit: () => void }>) {
  return (
    <div className="mt-3">
      <div className="text-status-pending text-sm font-bold">{m.test_submit_error()}</div>
      <button
        className="text-ink-tertiary cursor-pointer text-xs underline active:opacity-60"
        onClick={onResubmit}
        type="button"
      >
        {m.common_retry()}
      </button>
    </div>
  );
}

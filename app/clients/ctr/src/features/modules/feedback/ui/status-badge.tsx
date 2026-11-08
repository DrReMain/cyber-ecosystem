import { m } from "#/paraglide/messages";
import { STATUS_CLOSED, STATUS_PENDING, STATUS_REPLIED } from "../wire";

// One badge for the ticket's three states, shared by the queue row and
// the conversation header — the palette is the prototype's .st family
// (green replied, amber pending, gray closed); an unknown status draws
// nothing rather than masquerading.
export function StatusBadge({ status }: Readonly<{ status: string | undefined }>) {
  const badge =
    status === STATUS_REPLIED
      ? { cls: "bg-status-ok-bg text-status-ok", label: m.feedback_status_replied() }
      : status === STATUS_CLOSED
        ? { cls: "bg-status-off-bg text-status-off", label: m.feedback_status_closed() }
        : status === STATUS_PENDING
          ? {
              cls: "bg-status-pending-bg text-status-pending",
              label: m.feedback_status_pending(),
            }
          : null;
  if (badge === null) {
    return null;
  }
  return (
    <span
      className={`shrink-0 rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold ${badge.cls}`}
    >
      {badge.label}
    </span>
  );
}

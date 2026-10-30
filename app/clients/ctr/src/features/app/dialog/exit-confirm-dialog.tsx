import { m } from "#/paraglide/messages";
import { BottomConfirmSheet } from "./bottom-confirm-sheet";

interface ExitConfirmDialogProps {
  /** Controlled: the caller owns open/close; onConfirm does not auto-close. */
  open: boolean;
  onConfirm: () => void;
  /** Fires on cancel button, Escape, and backdrop press (all dismiss paths). */
  onCancel: () => void;
  title?: string;
  description?: string;
}

/**
 * The confirm-before-loss sheet for flow exits (leaving a running test,
 * abandoning an unfinished form). The prototype's own button arrangement
 * carries the safety: "keep testing" is the primary, focused answer and
 * leaving is the ghost — a stray Enter keeps the user in the flow, and
 * red is never spent on an exit the user explicitly chose.
 */
export function ExitConfirmDialog({
  open,
  onConfirm,
  onCancel,
  title = m.common_exit_title(),
  description = m.common_exit_desc(),
}: Readonly<ExitConfirmDialogProps>) {
  // Mount-conditional: RAC 1.21's exit-animation state machine can deadlock
  // on close (an exiting ref never resolves, so isOpen=false never unmounts
  // the overlay) — the caller's open flag owns the mount cycle directly.
  if (!open) {
    return null;
  }
  return (
    <BottomConfirmSheet
      actions={[
        {
          label: m.common_exit_resume(),
          kind: "primary",
          onPress: onCancel,
          autoFocus: true,
        },
        { label: m.common_exit_leave(), kind: "ghost", onPress: onConfirm },
      ]}
      body={description}
      onDismiss={onCancel}
      title={title}
    />
  );
}

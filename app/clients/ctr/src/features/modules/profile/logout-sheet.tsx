import { Button, Dialog, ModalOverlay } from "react-aria-components";
import { m } from "#/paraglide/messages";

interface LogoutSheetProps {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

/**
 * The prototype's bottom-sheet confirm for sign-out (not a centered dialog):
 * cancel is the autoFocus target so a stray Enter keeps the session, and
 * confirm never auto-closes — the caller owns closing after logout settles.
 *
 * Mount-conditional on purpose: RAC 1.21's exit-animation state machine can
 * deadlock on close (an exiting ref never resolves, so isOpen=false never
 * unmounts the overlay) — the caller's open flag owns the mount cycle
 * directly, close means unmount and there is no exit phase to hang.
 */
export function LogoutSheet({ open, onCancel, onConfirm }: Readonly<LogoutSheetProps>) {
  if (!open) {
    return null;
  }
  return (
    <ModalOverlay
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/70 backdrop-blur-sm"
      isDismissable
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onCancel();
        }
      }}
      // The backdrop press is owned natively too: filterDOMProps does pass
      // pointer handlers, and target-is-overlay means the press landed on
      // the dim area, not the sheet. Idempotent with onOpenChange.
      onPointerDown={(e) => {
        if (e.target === e.currentTarget) {
          onCancel();
        }
      }}
    >
      <Dialog
        className="safe-mb mx-3 w-full max-w-md rounded-2xl border border-line bg-surface-card p-4 pb-6 outline-none"
        role="alertdialog"
      >
        <div className="flex flex-col items-center gap-1 pb-4">
          <div className="font-bold text-ink text-sm">{m.profile_logout_title()}</div>
          <p className="text-center text-[11.5px] text-ink-tertiary leading-relaxed">
            {m.profile_logout_body()}
          </p>
        </div>
        <div className="flex gap-3">
          <Button
            autoFocus
            className="flex-1 rounded-sm border border-line px-4 py-2.5 text-ink-secondary text-sm"
            onPress={onCancel}
          >
            {m.profile_logout_cancel()}
          </Button>
          <Button
            className="flex-1 rounded-sm border border-danger-border bg-danger/20 px-4 py-2.5 font-bold text-danger-text text-sm"
            onPress={onConfirm}
          >
            {m.profile_logout_confirm()}
          </Button>
        </div>
      </Dialog>
    </ModalOverlay>
  );
}

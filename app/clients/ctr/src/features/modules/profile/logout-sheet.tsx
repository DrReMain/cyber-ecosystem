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
        className="safe-mb border-line bg-surface-card mx-3 w-full max-w-md rounded-2xl border p-4 pb-6 outline-none"
        role="alertdialog"
      >
        <div className="flex flex-col items-center gap-1 pb-4">
          <div className="text-ink text-sm font-bold">{m.profile_logout_title()}</div>
          <p className="text-ink-tertiary text-center text-[11.5px] leading-relaxed">
            {m.profile_logout_body()}
          </p>
        </div>
        <div className="flex gap-3">
          <Button
            autoFocus
            className="border-line text-ink-secondary flex-1 rounded-sm border px-4 py-2.5 text-sm"
            onPress={onCancel}
          >
            {m.profile_logout_cancel()}
          </Button>
          <Button
            className="border-danger-border bg-danger/20 text-danger-text flex-1 rounded-sm border px-4 py-2.5 text-sm font-bold"
            onPress={onConfirm}
          >
            {m.profile_logout_confirm()}
          </Button>
        </div>
      </Dialog>
    </ModalOverlay>
  );
}

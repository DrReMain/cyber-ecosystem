import { Button, Dialog, ModalOverlay } from "react-aria-components";
import { m } from "#/paraglide/messages";

interface ExitConfirmDialogProps {
  /** Controlled: the caller owns open/close; onConfirm does not auto-close. */
  open: boolean;
  onConfirm: () => void;
  /** Fires on cancel button, Escape, and backdrop press (all dismiss paths). */
  onCancel: () => void;
  title?: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
}

/**
 * The confirm-before-loss dialog for flow exits (leaving a running test,
 * abandoning an unfinished form). Destructive semantics: danger accents,
 * cancel is the autoFocus target so an accidental Enter keeps the user in
 * the flow. react-aria supplies the machinery — focus trap and restore,
 * role="alertdialog" wiring, Escape/backdrop dismissal — no RAC navigation
 * APIs are used, so no RouterProvider context is required.
 */
export function ExitConfirmDialog({
  open,
  onConfirm,
  onCancel,
  title = m.common_exit_title(),
  description = m.common_exit_desc(),
  confirmLabel = m.common_exit_confirm(),
  cancelLabel = m.common_exit_cancel(),
}: Readonly<ExitConfirmDialogProps>) {
  // Mount-conditional: RAC 1.21's exit-animation state machine can deadlock
  // on close (an exiting ref never resolves, so isOpen=false never unmounts
  // the overlay) — the caller's open flag owns the mount cycle directly.
  if (!open) {
    return null;
  }
  return (
    <ModalOverlay
      className="motion-safe:data-[entering]:ease-out-expo fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-6 backdrop-blur-md motion-safe:data-[entering]:duration-200"
      isDismissable
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) onCancel();
      }}
      onPointerDown={(e) => {
        if (e.target === e.currentTarget) {
          onCancel();
        }
      }}
    >
      <Dialog
        className="rounded-card border-danger/40 bg-surface-card shadow-card motion-safe:data-[entering]:ease-out-expo w-full max-w-md border p-8 motion-safe:data-[entering]:scale-95 motion-safe:data-[entering]:opacity-0 motion-safe:data-[entering]:duration-200"
        role="alertdialog"
      >
        <h2 className="text-ink text-center text-xl font-bold tracking-[0.02em] rtl:tracking-normal">
          {title}
        </h2>
        <p className="text-ink-secondary mt-3 text-center text-sm leading-relaxed">{description}</p>
        <div className="mt-8 flex gap-4">
          <Button
            autoFocus
            className="rounded-control bg-surface-input text-ink ease-out-expo hover:bg-surface-active flex-1 cursor-pointer px-6 py-3.5 text-sm font-semibold transition-transform duration-200 active:scale-95"
            onPress={onCancel}
          >
            {cancelLabel}
          </Button>
          <Button
            className="rounded-control bg-danger shadow-cta-primary ease-out-expo hover:bg-danger/90 flex-1 cursor-pointer px-6 py-3.5 text-sm font-semibold text-white transition-transform duration-200 active:scale-95"
            onPress={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </Dialog>
    </ModalOverlay>
  );
}

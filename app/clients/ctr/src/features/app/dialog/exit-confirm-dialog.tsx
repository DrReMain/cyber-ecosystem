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
  return (
    <ModalOverlay
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-6 backdrop-blur-md motion-safe:data-[entering]:duration-200 motion-safe:data-[entering]:ease-out-expo"
      isDismissable
      isOpen={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) onCancel();
      }}
    >
      <Dialog
        className="w-full max-w-md rounded-card border border-danger/40 bg-surface-card p-8 shadow-card motion-safe:data-[entering]:scale-95 motion-safe:data-[entering]:opacity-0 motion-safe:data-[entering]:duration-200 motion-safe:data-[entering]:ease-out-expo"
        role="alertdialog"
      >
        <h2 className="text-center font-bold text-ink text-xl tracking-[0.02em] rtl:tracking-normal">
          {title}
        </h2>
        <p className="mt-3 text-center text-ink-secondary text-sm leading-relaxed">{description}</p>
        <div className="mt-8 flex gap-4">
          <Button
            autoFocus
            className="flex-1 cursor-pointer rounded-control bg-surface-input px-6 py-3.5 font-semibold text-ink text-sm transition-transform duration-200 ease-out-expo hover:bg-surface-active active:scale-95"
            onPress={onCancel}
          >
            {cancelLabel}
          </Button>
          <Button
            className="flex-1 cursor-pointer rounded-control bg-danger px-6 py-3.5 font-semibold text-sm text-white shadow-cta-primary transition-transform duration-200 ease-out-expo hover:bg-danger/90 active:scale-95"
            onPress={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </Dialog>
    </ModalOverlay>
  );
}

import type { ReactNode } from "react";
import { Button, Dialog, ModalOverlay } from "react-aria-components";

// The prototype's one global sheet (CTS.sheet): bottom-docked, top-corner
// radii only, a centred title/body, and a vertical stack of full-width
// buttons — primary ("keep going" answers), hot (deliberate irreversible
// yes), ghost (the way back). Every confirm float in the app renders
// through this shape; a page may not grow its own dialog geometry — a
// float that needs a visual above its copy gets it through the sheet's
// own centred slot, never a page-local float.

export type SheetActionKind = "primary" | "hot" | "ghost";

export interface SheetAction {
  label: string;
  kind: SheetActionKind;
  onPress: () => void;
  /** The focus target when the sheet opens — a stray Enter lands here. */
  autoFocus?: boolean;
}

interface BottomConfirmSheetProps {
  title: string;
  body?: string;
  /** A centred visual between the title and the body — content only; the
   * sheet owns the geometry. */
  visual?: ReactNode;
  /** Render order is stack order, top to bottom — the caller arranges. */
  actions: SheetAction[];
  /** Escape and backdrop press — every dismiss path lands here. */
  onDismiss: () => void;
}

const KIND_CLASS: Record<SheetActionKind, string> = {
  primary: "bg-cta-primary shadow-cta-primary font-bold text-white",
  hot: "bg-cta-practice shadow-cta-practice font-bold text-white",
  ghost: "border-[1.5px] border-ambient-slate bg-surface-input text-ink-secondary font-medium",
};

export function BottomConfirmSheet({
  title,
  body,
  visual,
  actions,
  onDismiss,
}: Readonly<BottomConfirmSheetProps>) {
  return (
    <ModalOverlay
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/70 backdrop-blur-sm"
      isDismissable
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onDismiss();
        }
      }}
      // The backdrop press is owned natively too: filterDOMProps does pass
      // pointer handlers, and target-is-overlay means the press landed on
      // the dim area, not the sheet. Idempotent with onOpenChange.
      onPointerDown={(e) => {
        if (e.target === e.currentTarget) {
          onDismiss();
        }
      }}
    >
      <Dialog
        aria-label={title}
        className="border-line-soft bg-surface-card w-full max-w-97.5 rounded-t-[18px] border px-4 pt-4.5 pb-[calc(18px+env(safe-area-inset-bottom))] outline-none"
        role="alertdialog"
      >
        <div className="text-ink text-center text-[15px] font-extrabold">{title}</div>
        {visual !== undefined ? <div className="mt-3.5 flex justify-center">{visual}</div> : null}
        {body !== undefined ? (
          <p className="text-ink-muted mt-1.5 mb-3.5 text-center text-[12.5px] leading-[1.7]">
            {body}
          </p>
        ) : null}
        <div className="flex flex-col gap-2.25">
          {actions.map((action) => (
            <Button
              autoFocus={action.autoFocus}
              className={`w-full cursor-pointer rounded-[12px] py-3.5 text-[15px] active:opacity-80 ${KIND_CLASS[action.kind]}`}
              key={action.label}
              onPress={action.onPress}
            >
              {action.label}
            </Button>
          ))}
        </div>
      </Dialog>
    </ModalOverlay>
  );
}

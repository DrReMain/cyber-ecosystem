import { Button, Dialog, ModalOverlay } from "react-aria-components";

// The prototype's one global sheet (CTS.sheet): bottom-docked, top-corner
// radii only, a centred title/body, and a vertical stack of full-width
// buttons — primary ("keep going" answers), hot (deliberate irreversible
// yes), ghost (the way back). Every confirm float in the app renders
// through this shape; a page may not grow its own dialog geometry.

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
  /** Render order is stack order, top to bottom — the caller arranges. */
  actions: SheetAction[];
  /** Escape and backdrop press — every dismiss path lands here. */
  onDismiss: () => void;
}

const KIND_CLASS: Record<SheetActionKind, string> = {
  primary: "bg-cta-primary shadow-cta-primary font-bold text-white",
  hot: "bg-cta-practice shadow-cta-practice font-bold text-white",
  ghost: "border-[1.5px] border-[#4B5563] bg-surface-input text-ink-secondary font-medium",
};

export function BottomConfirmSheet({
  title,
  body,
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
        className="border-line-soft bg-surface-card w-full max-w-[390px] rounded-t-[18px] border px-4 pt-[18px] pb-[calc(18px+env(safe-area-inset-bottom))] outline-none"
        role="alertdialog"
      >
        <div className="text-ink text-center text-[15px] font-extrabold">{title}</div>
        {body !== undefined ? (
          <p className="text-ink-muted mt-[6px] mb-3.5 text-center text-[12.5px] leading-[1.7]">
            {body}
          </p>
        ) : null}
        <div className="flex flex-col gap-[9px]">
          {actions.map((action) => (
            <Button
              autoFocus={action.autoFocus}
              className={`w-full cursor-pointer rounded-[12px] py-[14px] text-[15px] active:opacity-80 ${KIND_CLASS[action.kind]}`}
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

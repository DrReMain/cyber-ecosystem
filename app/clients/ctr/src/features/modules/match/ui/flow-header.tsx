import { ChevronLeft } from "lucide-react";
import type { ReactNode } from "react";
import { m } from "#/paraglide/messages";

interface FlowHeaderProps {
  title: string;
  onBack: () => void;
  /** Slot after the title — page badges (the done page's N/N tag). */
  trailing?: ReactNode;
}

/** The flow-page header: back fab + title, over the ambient canvas. */
export function FlowHeader({ title, onBack, trailing }: Readonly<FlowHeaderProps>) {
  return (
    <div className="flex items-center gap-3">
      <button
        aria-label={m.common_back()}
        className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
        onClick={onBack}
        type="button"
      >
        <ChevronLeft aria-hidden size={20} />
      </button>
      <h1 className="text-ink text-[19px] font-bold">{title}</h1>
      {trailing}
    </div>
  );
}

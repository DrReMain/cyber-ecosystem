import { ChevronLeft } from "lucide-react";
import { m } from "#/paraglide/messages";

interface FlowHeaderProps {
  title: string;
  onBack: () => void;
}

/** The flow-page header: back fab + title, over the ambient canvas. */
export function FlowHeader({ title, onBack }: Readonly<FlowHeaderProps>) {
  return (
    <div className="flex items-center gap-3">
      <button
        aria-label={m.common_back()}
        className="border-line text-ink-secondary flex h-11 w-11 shrink-0 items-center justify-center rounded-full border active:opacity-60"
        onClick={onBack}
        type="button"
      >
        <ChevronLeft aria-hidden size={18} />
      </button>
      <h1 className="text-ink text-lg font-extrabold">{title}</h1>
    </div>
  );
}

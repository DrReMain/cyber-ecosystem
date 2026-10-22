import { Suggestion } from "@ant-design/x";
import clsx from "clsx";
import { Check, ChevronDown, Loader2, Zap } from "lucide-react";
import { m } from "#/paraglide/messages";

interface ModelChipProps {
  models: string[];
  model: string | undefined;
  loading: boolean;
  onModelChange: (model: string) => void;
}

// The picker rides x's Suggestion — an upward-anchored list with keyboard
// navigation (arrows/enter/esc) and the icon/label/extra row layout built in —
// instead of a hand-rolled popover.
export function ModelChip({ models, model, loading, onModelChange }: Readonly<ModelChipProps>) {
  const items = models.map((id) => ({
    extra: id === model ? <Check className="text-primary size-3.5" /> : undefined,
    label: <span className="max-w-48 truncate">{id}</span>,
    value: id,
  }));

  return (
    <Suggestion items={items} onSelect={(value) => onModelChange(value)}>
      {({ onTrigger, onKeyDown, open }) => (
        <button
          className={clsx(
            "rounded-control border-line-soft text-ink-secondary hover:border-line hover:text-ink inline-flex h-7 max-w-56 items-center gap-1.5 border px-2.5 text-[12px] transition-colors",
            open && "border-line bg-fill-quaternary text-ink",
          )}
          onClick={() => (open ? onTrigger(false) : onTrigger())}
          onKeyDown={onKeyDown}
          type="button"
        >
          {loading && model === undefined ? (
            <Loader2 className="size-3.5 animate-spin" />
          ) : (
            <Zap className="text-ink-quaternary size-3.5" />
          )}
          <span className="truncate">
            {model ?? (loading ? m.agents_chat_model_loading() : m.agents_chat_model_empty())}
          </span>
          <ChevronDown
            className={clsx(
              "text-ink-quaternary size-3 shrink-0 transition-transform",
              open && "rotate-180",
            )}
          />
        </button>
      )}
    </Suggestion>
  );
}

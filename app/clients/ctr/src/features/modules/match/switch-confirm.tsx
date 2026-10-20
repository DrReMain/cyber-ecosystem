import { useNavigate, useSearch } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import type { Genre } from "./arranger";
import { FlowHeader } from "./ui/flow-header";
import { genreLabel } from "./vocabulary";

/**
 * The switch confirmation (prototype match-switch-confirm): current →
 * target cards, the three-rule card, confirm/cancel. The progress warning
 * face exists in the prototype but stays unrendered until the match session
 * carries real progress — this session's done count is always zero now.
 */
export function SwitchConfirmPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/match/switch" }) as { from?: Genre };
  const cur = search.from === "moba" ? "moba" : "fps";
  const tgt: Genre = cur === "fps" ? "moba" : "fps";

  function onConfirm() {
    toast.success(m.match_switch_toast({ type: genreLabel(tgt) }));
    void navigate({ to: "/match", search: { genre: tgt } });
  }

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader
          onBack={() => void navigate({ to: "/match", search: { genre: cur } })}
          title={m.match_switch_title()}
        />

        <div className="rounded-lg border border-line bg-surface-card p-4">
          <div className="flex items-center gap-2.5">
            <div className="flex-1 rounded-tile bg-surface-deep p-3 text-center">
              <div className="text-[10.5px] text-ink-tertiary">{m.match_switch_cur()}</div>
              <div className="mt-1 font-extrabold text-ink text-sm">{genreLabel(cur)}</div>
            </div>
            <ArrowRight aria-hidden className="shrink-0 text-ink-tertiary" size={16} />
            <div className="flex-1 rounded-tile border border-line-accent-purple/60 bg-surface-deep p-3 text-center">
              <div className="text-[10.5px] text-ink-tertiary">{m.match_switch_tgt()}</div>
              <div className="mt-1 font-extrabold text-ink text-sm">{genreLabel(tgt)}</div>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-line bg-surface-deep p-3.5 text-[12px] text-ink-tertiary leading-relaxed">
          <div className="font-bold text-ink-secondary">{m.match_switch_rules()}</div>
          <div className="mt-1.5">· {m.match_switch_rule1()}</div>
          <div>· {m.match_switch_rule2()}</div>
          <div>· {m.match_switch_rule3()}</div>
        </div>

        <button
          className="w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
          onClick={onConfirm}
          type="button"
        >
          {m.match_switch_ok()}
        </button>
        <button
          className="w-full rounded-sm border border-line bg-surface-content px-4 py-3 font-semibold text-ink active:opacity-60"
          onClick={() => void navigate({ to: "/match", search: { genre: cur } })}
          type="button"
        >
          {m.match_switch_cancel()}
        </button>
      </div>
    </AmbientLayout>
  );
}

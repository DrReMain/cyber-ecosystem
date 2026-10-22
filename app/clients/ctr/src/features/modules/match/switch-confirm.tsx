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

        <div className="border-line bg-surface-card rounded-card border p-4">
          <div className="flex items-center gap-2.5">
            <div className="rounded-tile bg-surface-deep flex-1 p-3 text-center">
              <div className="text-ink-tertiary text-[10.5px]">{m.match_switch_cur()}</div>
              <div className="text-ink mt-1 text-sm font-extrabold">{genreLabel(cur)}</div>
            </div>
            <ArrowRight aria-hidden className="text-ink-tertiary shrink-0" size={16} />
            <div className="rounded-tile border-line-accent-purple/60 bg-surface-deep flex-1 border p-3 text-center">
              <div className="text-ink-tertiary text-[10.5px]">{m.match_switch_tgt()}</div>
              <div className="text-ink mt-1 text-sm font-extrabold">{genreLabel(tgt)}</div>
            </div>
          </div>
        </div>

        <div className="border-line bg-surface-deep text-ink-tertiary rounded-control border p-3.5 text-[12px] leading-relaxed">
          <div className="text-ink-secondary font-bold">{m.match_switch_rules()}</div>
          <div className="mt-1.5">· {m.match_switch_rule1()}</div>
          <div>· {m.match_switch_rule2()}</div>
          <div>· {m.match_switch_rule3()}</div>
        </div>

        <button
          className="bg-cta-primary rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
          onClick={onConfirm}
          type="button"
        >
          {m.match_switch_ok()}
        </button>
        <button
          className="border-line bg-surface-content text-ink rounded-control w-full border px-4 py-[14px] text-[15px] font-bold active:opacity-60"
          onClick={() => void navigate({ to: "/match", search: { genre: cur } })}
          type="button"
        >
          {m.match_switch_cancel()}
        </button>
      </div>
    </AmbientLayout>
  );
}

import { useQuery as useConnectQuery, useMutation } from "@connectrpc/connect-query";
import {
  getMatchSession,
  switchMatchGenre,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { invalidateMatchSession } from "./session";
import { FlowHeader } from "./ui/flow-header";
import { genreLabel, narrowGenre } from "./vocabulary";

/* The confirm page's own single orb (prototype match-switch-confirm): one
 * purple sphere bleeding off the top-right corner — not the login trio. */
const switchOrbs: readonly AmbientOrb[] = [
  { className: "-top-15 -right-15 size-55", tone: "purple" },
];

/**
 * The switch confirmation (prototype match-switch-confirm): current →
 * target cards read from the live session, the three-rule card, the
 * progress warning when this session already has stored attempts, and the
 * real switch call — the server closes the active session (even a fresh
 * one) and draws a replacement under a forced-new seed.
 */
export function SwitchConfirmPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const sessionQuery = useConnectQuery(getMatchSession, {});
  const session = sessionQuery.data?.session ?? null;
  const doneCount = (session?.done ?? []).length;
  const cur = narrowGenre(session?.genre) ?? "fps";
  const tgt = cur === "fps" ? ("moba" as const) : ("fps" as const);
  const doSwitch = useMutation(switchMatchGenre, {
    onSuccess: () => {
      invalidateMatchSession(queryClient);
      toast.success(m.match_switch_toast({ type: genreLabel(tgt) }));
      void navigate({ to: "/match" });
    },
  });

  return (
    <AmbientLayout orbs={switchOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/match" })} title={m.match_switch_title()} />

        <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
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
          {/* Walkthrough N1 (registered): rule 2 deliberately diverges from
              the prototype — the prototype promises 「已测过的游戏不再重复编
              排」, but the server's switch law closes the session and redraws
              under a forced-new seed, so the same test CAN be drawn again.
              The copy stays honest to the behavior; never "fix" it back to
              the prototype's promise of an absent guarantee. */}
          <div className="mt-1.5">· {m.match_switch_rule1()}</div>
          <div>· {m.match_switch_rule2()}</div>
          <div>· {m.match_switch_rule3()}</div>
        </div>

        {doneCount > 0 && (
          <div className="border-status-pending/45 bg-surface-deep text-status-pending rounded-control border p-3.5 text-[12px] leading-relaxed">
            {m.match_switch_warn({ n: doneCount, type: genreLabel(tgt) })}
          </div>
        )}

        <button
          className="bg-cta-primary rounded-control w-full px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80 disabled:opacity-40"
          disabled={doSwitch.isPending}
          onClick={() => doSwitch.mutate({ genre: tgt })}
          type="button"
        >
          {m.match_switch_ok()}
        </button>
        <button
          className="border-line bg-surface-content text-ink rounded-control w-full border px-4 py-3.5 text-[15px] font-bold active:opacity-60"
          onClick={() => void navigate({ to: "/match" })}
          type="button"
        >
          {m.match_switch_cancel()}
        </button>
      </div>
    </AmbientLayout>
  );
}

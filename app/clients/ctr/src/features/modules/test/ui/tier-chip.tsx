import { m } from "#/paraglide/messages";

// The grade chip's four-color ladder (prototype .tier) and its grey off
// state — the server's curve-pending verdicts render as 未评级.
const TIERS: Record<string, { readonly cls: string; readonly cn: () => string }> = {
  S: { cls: "text-grade-s bg-[rgba(253,224,71,0.15)]", cn: m.test_grade_s },
  A: { cls: "text-grade-a bg-status-ok-bg", cn: m.test_grade_a },
  B: { cls: "text-grade-b bg-[rgba(59,130,246,0.15)]", cn: m.test_grade_b },
  C: { cls: "text-grade-c bg-[rgba(249,115,22,0.15)]", cn: m.test_grade_c },
};

export function TierChip({ grade }: Readonly<{ grade: string | undefined }>) {
  const tier = grade !== undefined ? TIERS[grade] : undefined;
  if (tier === undefined) {
    return (
      <span className="text-status-off rounded-full bg-[rgba(107,114,128,0.15)] px-2.5 py-0.75 text-[11px] font-extrabold">
        {m.test_grade_off()}
      </span>
    );
  }
  return (
    <span className={`rounded-full px-2.5 py-0.75 text-[11px] font-extrabold ${tier.cls}`}>
      {grade} · {tier.cn()}
    </span>
  );
}

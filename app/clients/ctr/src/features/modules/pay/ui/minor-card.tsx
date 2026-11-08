import { Button } from "react-aria-components";

/** The minor's wall (prototype pay-confirm, the old draft's escape-hatch
 * fix): the refusal card the compliance prompt may override, then the two
 * ways out — back to the free tier, back to home. No pay button renders
 * at all under this state; the exits are the page. */
export function MinorCard({
  copy,
  freeLabel,
  homeLabel,
  onFree,
  onHome,
}: Readonly<{
  copy: string;
  freeLabel: string;
  homeLabel: string;
  onFree: () => void;
  onHome: () => void;
}>) {
  return (
    <div className="flex flex-col gap-3">
      <div className="rounded-card bg-surface-card shadow-card border border-[rgba(239,68,68,0.5)] px-3.5 py-3.5 text-[12px] leading-[1.85] text-[#FCA5A5] backdrop-blur-md">
        {copy}
      </div>
      <Button
        className="bg-cta-primary shadow-cta-primary w-full rounded-[12px] py-3.5 text-[15px] font-bold text-white active:opacity-80"
        onPress={onFree}
      >
        {freeLabel}
      </Button>
      <Button
        className="bg-surface-input text-ink-secondary border-ambient-slate w-full rounded-[12px] border-[1.5px] py-3.5 text-[15px] font-medium active:opacity-80"
        onPress={onHome}
      >
        {homeLabel}
      </Button>
    </div>
  );
}

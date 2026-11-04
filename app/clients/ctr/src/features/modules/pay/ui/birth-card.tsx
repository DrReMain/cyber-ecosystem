import { Button } from "react-aria-components";

/** The missing-birth wall: the server refuses BIRTH_MISSING and routes
 * here — the card states the need (the compliance prompt may override
 * the copy) and hands the member to the backfill page. */
export function BirthCard({
  actionLabel,
  copy,
  onBackfill,
}: Readonly<{
  actionLabel: string;
  copy: string;
  onBackfill: () => void;
}>) {
  return (
    <div className="flex flex-col gap-3">
      <div className="border-line bg-surface-card shadow-card rounded-card text-ink-muted border px-4 py-[18px] text-[12px] leading-[1.85] backdrop-blur-md">
        {copy}
      </div>
      <Button
        className="bg-cta-primary shadow-cta-primary w-full rounded-[12px] py-[14px] text-[15px] font-bold text-white active:opacity-80"
        onPress={onBackfill}
      >
        {actionLabel}
      </Button>
    </div>
  );
}

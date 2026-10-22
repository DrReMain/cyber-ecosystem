import { m } from "#/paraglide/messages";

interface SoonCardProps {
  title: string;
  description: string;
}

/**
 * The dashed placeholder card for modules that have not landed yet (home's
 * assessments-and-reports block, profile's more-modules block): one chrome,
 * per-surface copy — the coming-soon badge is shared vocabulary.
 */
export function SoonCard({ title, description }: Readonly<SoonCardProps>) {
  return (
    <div className="border-ink-tertiary/60 bg-surface-deep/50 rounded-lg border-dashed p-4">
      <div className="flex items-center justify-between">
        <span className="text-ink-tertiary text-sm font-bold">{title}</span>
        <span className="bg-status-pending-bg text-status-pending rounded-full px-2 py-0.5 text-[10px]">
          {m.common_soon_badge()}
        </span>
      </div>
      <p className="text-ink-tertiary mt-1.5 text-center text-[11.5px] leading-relaxed">
        {description}
      </p>
    </div>
  );
}

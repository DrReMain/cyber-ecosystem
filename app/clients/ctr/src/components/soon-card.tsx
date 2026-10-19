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
    <div className="rounded-lg border-ink-tertiary/60 border-dashed bg-surface-deep/50 p-4">
      <div className="flex items-center justify-between">
        <span className="font-bold text-ink-tertiary text-sm">{title}</span>
        <span className="rounded-full bg-status-pending-bg px-2 py-0.5 text-[10px] text-status-pending">
          {m.common_soon_badge()}
        </span>
      </div>
      <p className="mt-1.5 text-center text-[11.5px] text-ink-tertiary leading-relaxed">
        {description}
      </p>
    </div>
  );
}

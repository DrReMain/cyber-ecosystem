/** The standing banner above the wall (态②/③): an active window
 * states its expiry date on the membership tint, the membership tag
 * beside it; an expired one states the past date on the neutral gray.
 * The caller hands the fully localized sentence (the date rides inside
 * the message). */
export function StateBanner({
  expired,
  tag,
  text,
}: Readonly<{ expired?: boolean; tag?: string; text: string }>) {
  if (expired) {
    return (
      <div className="rounded-card bg-surface-deep border border-[rgba(107,114,128,0.4)] px-4 py-3.5">
        <span className="text-ink-muted text-[12.5px]">{text}</span>
      </div>
    );
  }
  return (
    <div className="rounded-card border border-[rgba(168,85,247,0.4)] bg-[linear-gradient(135deg,rgba(168,85,247,0.08),rgba(17,24,39,0.92))] px-4 py-3.5">
      <span className="flex items-center gap-2">
        {tag !== undefined ? (
          <span className="text-status-soon rounded-full border border-[rgba(168,85,247,0.5)] bg-[rgba(168,85,247,0.22)] px-2.5 py-0.5 text-[10.5px] font-semibold">
            {tag}
          </span>
        ) : null}
        <span className="text-[12.5px] text-[#E5E7EB]">{text}</span>
      </span>
    </div>
  );
}

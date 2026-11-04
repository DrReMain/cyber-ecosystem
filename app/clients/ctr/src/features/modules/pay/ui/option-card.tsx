/** One purchase option in the choose-how card (态①/③): name, the
 * membership's "better value" badge, one-line pitch, and the catalog
 * price — the membership row highlighted as the recommended pick. */
export function OptionCard({
  bestBadge,
  desc,
  highlighted,
  name,
  price,
}: Readonly<{
  bestBadge?: string;
  desc: string;
  highlighted?: boolean;
  name: string;
  price: string;
}>) {
  return (
    <div
      className={`rounded-[14px] border p-3.5 ${
        highlighted
          ? "border-[rgba(168,85,247,0.7)] bg-[linear-gradient(135deg,rgba(168,85,247,0.16),rgba(17,24,39,0.92))]"
          : "border-[rgba(168,85,247,0.45)] bg-[linear-gradient(135deg,rgba(168,85,247,0.08),rgba(17,24,39,0.92))]"
      }`}
    >
      <div className="flex items-center justify-between">
        <span className="text-ink text-[15px] font-extrabold">
          {name}
          {bestBadge !== undefined ? (
            <span className="ml-1.5 rounded-full border border-[rgba(168,85,247,0.4)] bg-[rgba(168,85,247,0.2)] px-2 py-px text-[9.5px] font-semibold text-[#C4B5FD]">
              {bestBadge}
            </span>
          ) : null}
        </span>
        <span className="text-[20px] font-black text-[#FDE047] tabular-nums">{price}</span>
      </div>
      <p className="text-ink-muted mt-1 text-left text-[10.5px] leading-[1.7]">{desc}</p>
    </div>
  );
}

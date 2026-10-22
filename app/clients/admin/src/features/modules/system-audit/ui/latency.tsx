const tone = (ms: number) => (ms >= 5000 ? "text-red-500" : ms >= 1000 ? "text-amber-500" : "");

export function Latency({ ms }: Readonly<{ ms: number }>) {
  return (
    <span className="inline-flex items-baseline gap-0.5">
      <span className={`font-mono text-[13px] font-medium ${tone(ms)}`}>{ms}</span>
      <span className="text-ink-tertiary text-[11px]">ms</span>
    </span>
  );
}

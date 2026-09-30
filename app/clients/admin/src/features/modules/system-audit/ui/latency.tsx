const tone = (ms: number) => (ms >= 5000 ? "text-red-500" : ms >= 1000 ? "text-amber-500" : "");

export function Latency({ ms }: Readonly<{ ms: number }>) {
  return (
    <span className="inline-flex items-baseline gap-0.5">
      <span className={`font-medium font-mono text-[13px] ${tone(ms)}`}>{ms}</span>
      <span className="text-[11px] text-black/45 dark:text-white/45">ms</span>
    </span>
  );
}

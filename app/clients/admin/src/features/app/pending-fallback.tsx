import { m } from "#/paraglide/messages";

export function PendingFallback() {
  return (
    <div className="relative flex size-full min-h-24 grow items-center justify-center">
      <style>{`
@keyframes nc-sweep {
  from { transform: translateX(0); }
  to { transform: translateX(350%); }
}
@keyframes nc-sweep-rev {
  from { transform: translateX(0); }
  to { transform: translateX(-350%); }
}`}</style>
      <div className="relative w-[min(200px,60%)]" role="status">
        <span className="sr-only">{m.common_loading()}</span>
        <div
          aria-hidden
          className="text-center font-mono text-[11px] text-current/50 uppercase tracking-[0.18em]"
          dir="ltr"
        >
          LOADING
        </div>
        <span
          aria-hidden
          className="relative mt-2 block h-0.5 overflow-hidden rounded bg-current/12"
        >
          <span className="absolute inset-s-[-40%] inset-y-0 w-2/5 bg-linear-to-r from-transparent via-current/70 to-transparent motion-safe:animate-[nc-sweep_2.6s_ease-in-out_infinite] rtl:motion-safe:animate-[nc-sweep-rev_2.6s_ease-in-out_infinite]" />
        </span>
      </div>
    </div>
  );
}

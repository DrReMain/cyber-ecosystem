import { view } from "#/domains/error";
import { m } from "#/paraglide/messages";

export function ErrorFallback({
  error,
  onRetry,
}: Readonly<{ error: unknown; onRetry?: () => void }>) {
  const resolved = view(error);
  const retry = onRetry ? (
    <button
      className="cursor-pointer text-[12.5px] underline decoration-current/30 underline-offset-3 transition-colors hover:decoration-current"
      onClick={onRetry}
      type="button"
    >
      {m.common_retry()}
    </button>
  ) : null;

  if (resolved.silent) {
    return (
      <div className="flex size-full min-h-24 grow flex-col items-center justify-center gap-2 p-4 text-center">
        <span className="text-[12.5px] text-current/55" role="status">
          {resolved.title}
        </span>
        {retry}
      </div>
    );
  }

  return (
    <div className="flex size-full min-h-24 grow flex-col items-center justify-center gap-2 p-4 text-center">
      <span
        className="rounded-full border border-current/15 px-2.5 py-0.75 font-mono text-[9.5px] text-current/40 uppercase tracking-[0.18em]"
        dir="ltr"
      >
        {`ERR · ${resolved.kind.toUpperCase()}`}
      </span>
      <span className="text-[12.5px] text-current/55">{resolved.title}</span>
      {retry}
    </div>
  );
}

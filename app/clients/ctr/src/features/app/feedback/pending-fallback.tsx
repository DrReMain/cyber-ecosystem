import { m } from "#/paraglide/messages";
import { LoadingStrip } from "../states/loading-strip";

export function PendingFallback() {
  return (
    <div className="relative flex size-full min-h-app grow items-center justify-center">
      <div className="w-[min(200px,60%)]" role="status">
        <span className="sr-only">{m.common_loading()}</span>
        <LoadingStrip tone="inherit" />
      </div>
    </div>
  );
}

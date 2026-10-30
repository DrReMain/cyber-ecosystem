import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";

// Shared by the bind callbacks: the transient "binding…" pane with the
// honest exit — every verdict that lands here navigates onward itself.
export function PendingLanding() {
  return (
    <main className="safe-px safe-pt min-h-app flex items-center justify-center gap-3">
      <p className="text-ink-tertiary">{m.security_bind_wechat_pending()}</p>
      <Link className="text-ink-secondary text-sm underline" to="/profile/security">
        {m.common_back()}
      </Link>
    </main>
  );
}

// Route-loading placeholder — shown while an async route component loads.
// Aligned with the other fallbacks: minimal, no antd, no i18n (yet).
// TODO(error-wiring): i18n copy ("Loading…" is hardcoded).
export function PendingFallback() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-2">
      <p className="animate-pulse text-sm opacity-70">Loading…</p>
    </div>
  );
}

// 404 presentation — aligned with ErrorFallback: minimal, plain html +
// Tailwind, no antd, no i18n (yet).
// TODO(error-wiring): i18n copy ("This page could not be found." is hardcoded).
export function NotFoundFallback() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-2">
      <h1 className="font-bold text-4xl">404</h1>
      <p className="opacity-70">This page could not be found.</p>
      <a className="border px-2 py-1" href="/">
        Back Home
      </a>
    </div>
  );
}

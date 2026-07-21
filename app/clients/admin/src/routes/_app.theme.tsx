import { createFileRoute } from "@tanstack/react-router";
import { ThemeToggle, useTheme } from "#/domains/theme";

export const Route = createFileRoute("/_app/theme")({ component: ThemeDemo });

const MODES = ["light", "dark", "system"] as const;

function ThemeDemo() {
  const { mode, preference, compact, setMode, setCompact } = useTheme();

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <ThemeToggle />
      </section>

      <section className="space-y-2">
        <div className="flex gap-2">
          {MODES.map((m) => (
            <button
              className={`rounded border px-3 py-1 text-sm ${mode === m ? "border-current font-semibold" : ""}`}
              key={m}
              onClick={() => setMode(m)}
              type="button"
            >
              {m}
            </button>
          ))}
        </div>
      </section>

      <section className="space-y-2">
        <button
          className="rounded border px-3 py-1 text-sm"
          onClick={() => setCompact(!compact)}
          type="button"
        >
          {compact ? "compact on" : "compact off"}
        </button>
      </section>

      <section className="space-y-1 text-sm opacity-70">
        <p>mode: {mode}</p>
        <p>preference: {preference}</p>
        <p>compact: {String(compact)}</p>
      </section>
    </div>
  );
}

import { m } from "#/paraglide/messages";

export function HomePage() {
  return (
    <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center">
      <h1 className="font-bold text-2xl text-ink">{m.home_title()}</h1>
    </main>
  );
}

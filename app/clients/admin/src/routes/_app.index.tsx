import { createFileRoute } from "@tanstack/react-router";
import { LocaleSwitcher } from "#/domains/i18n";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/")({ component: Home });

function Home() {
  return (
    <div className="space-y-4">
      <h1 className="font-bold text-3xl">{m.welcome({ name: "World" })}</h1>
      <LocaleSwitcher />
    </div>
  );
}

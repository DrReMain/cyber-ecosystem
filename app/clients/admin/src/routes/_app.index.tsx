import { createFileRoute } from "@tanstack/react-router";
import { LocaleSwitcher } from "#/domains/i18n";

export const Route = createFileRoute("/_app/")({ component: Home });

function Home() {
  return (
    <div className="space-y-4">
      <h1 className="font-bold text-3xl">Home</h1>
      <LocaleSwitcher />
    </div>
  );
}

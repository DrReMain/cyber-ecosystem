import { m } from "#/paraglide/messages";

export function DashboardNotFound() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-4">
      <div className="font-black text-9xl">404</div>
      <div className="text-xl">{m.not_found_title()}</div>
    </div>
  );
}

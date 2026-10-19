import { m } from "#/paraglide/messages";

export function AuthDivider() {
  return (
    <div className="my-6 flex items-center gap-3.5 font-mono text-[10px] text-black/35 tracking-[0.3em] rtl:tracking-normal dark:text-white/35">
      <span className="h-px flex-1 bg-black/6 dark:bg-white/6" />
      {m.login_default_divider_or()}
      <span className="h-px flex-1 bg-black/6 dark:bg-white/6" />
    </div>
  );
}

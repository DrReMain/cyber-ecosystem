import { m } from "#/paraglide/messages";
import { useLogout } from "./use-logout";

export function ErrorLogoutButton() {
  const logout = useLogout();

  return (
    <button
      className="inline-flex cursor-pointer items-center gap-2 rounded-xl border border-white/10 bg-white/4 px-7 py-3.25 font-semibold text-[#e8ecf4] text-sm tracking-[0.04em] backdrop-blur-[6px] transition-all hover:border-white/40 hover:bg-white/8 rtl:tracking-normal"
      disabled={logout.isPending}
      onClick={() => {
        if (!logout.isPending) void logout.mutateAsync();
      }}
      type="button"
    >
      {m.layout_dashboard_logout()}
    </button>
  );
}

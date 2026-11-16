import { useQuery } from "@tanstack/react-query";
import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { getLoginConfigFn } from "../login-config.fn";
import { AuthDivider } from "./auth-divider";

// Renders only what the deployment wired: a closed deployment (no enabled
// identity providers) renders nothing — the account form is the whole page,
// never a dead button. The per-provider flows themselves are a Phase 2
// wiring; their entry points appear the moment a provider is enabled.
export function SsoLogin() {
  const { data } = useQuery({
    queryKey: ["login-config"],
    queryFn: () => getLoginConfigFn(),
    throwOnError: false,
    staleTime: Infinity,
  });
  if (data === undefined || data.providers.length === 0) {
    return null;
  }
  return (
    <>
      <AuthDivider />
      <Button
        className="h-11 w-full cursor-pointer rounded-[10px] border border-black/10 bg-black/3 text-[13.5px] font-medium transition-all duration-200 hover:border-black/40 hover:bg-black/8 dark:border-white/10 dark:bg-white/4 dark:hover:border-white/40 dark:hover:bg-white/8"
        type="button"
      >
        {m.login_default_sso()}
      </Button>
    </>
  );
}

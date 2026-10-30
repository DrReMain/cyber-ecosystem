import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { clearWechatProfile, parkWechatProfile } from "#/features/app/wechat-profile";
import { bindWechatCredentialFn } from "#/features/modules/security/security.fn";
import { m } from "#/paraglide/messages";

// The wizard pill's landing: the bind closes and the person behind the
// openid arrives in the same exchange, so one hop carries both — park the
// profile and return to the nickname step, which prefills from it. An
// occupied openid has no wizard answer (the merge offer lives on the
// security face, not in onboarding), so every refusal lands back on the
// step with manual input as the honest path.
export function SetupWechatCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const search = useSearch({ from: "/profile/setup/wechat-callback" });
  const started = useRef(false);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code, state } = search;
    if (code === undefined || state === undefined) {
      // Wechat never redirects here on a refused grant (it keeps the user
      // in its own page), so a codeless arrival is a stray bookmark or a
      // mistyped URL — the honest answer is the wizard step, not a dead
      // "processing" pane.
      toast.error(m.profile_setup_wechat_bind_failed());
      void navigate({ to: "/profile/setup/nickname", viewTransition: false });
      return;
    }
    void bindWechatCredentialFn({ data: { code, state } }).then(async (res) => {
      if (res.status === "ok") {
        if (res.wechatProfile != null) {
          parkWechatProfile(res.wechatProfile);
        }
        toast.success(m.profile_setup_wechat_bound());
        await queryClient.invalidateQueries();
        await navigate({ to: "/profile/setup/nickname", viewTransition: false });
        return;
      }
      toast.error(m.profile_setup_wechat_bind_failed());
      // A refused bind also retires whatever an earlier OAuth moment
      // parked — the server just disowned that identity for this seat.
      clearWechatProfile();
      await navigate({ to: "/profile/setup/nickname", viewTransition: false });
    });
  }, [navigate, queryClient, search]);

  return (
    <main className="safe-px safe-pt min-h-app flex items-center justify-center gap-3">
      <p className="text-ink-tertiary">{m.profile_setup_wechat_pending()}</p>
      <Link className="text-ink-secondary text-sm underline" to="/profile/setup/nickname">
        {m.common_back()}
      </Link>
    </main>
  );
}

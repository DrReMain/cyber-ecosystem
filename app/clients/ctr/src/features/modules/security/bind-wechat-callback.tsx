import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { m } from "#/paraglide/messages";
import { PendingLanding } from "./pending-landing";
import { beginBindWechatFn, bindWechatCredentialFn } from "./security.fn";

// The bind leg's landing: the happy exchange closes here; an occupied
// wechat (6005) immediately opens a fresh OAuth leg toward the preview
// callback — the openid needed for the offer verdict only materializes at
// an exchange, so the preview cannot reuse this leg's burned code.
export function BindWechatCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const search = useSearch({ from: "/profile/security/bind-wechat/callback" });
  const started = useRef(false);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code, state } = search;
    if (code === undefined || state === undefined) {
      // A codeless arrival is a stray bookmark (wechat never redirects on
      // a refused grant) — bounce to the bind page's failed face rather
      // than parking on a "binding…" pane forever.
      void navigate({
        search: { failed: true },
        to: "/profile/security/bind-wechat",
        viewTransition: false,
      });
      return;
    }
    void bindWechatCredentialFn({ data: { code, state } }).then(async (res) => {
      if (res.status === "ok") {
        toast.success(m.security_bind_wechat_done());
        await queryClient.invalidateQueries();
        await navigate({ to: "/profile/security", viewTransition: false });
        return;
      }
      if (systemReason(res.reason) === System.IDENTITY_ALREADY_BOUND) {
        const begin = await beginBindWechatFn({
          data: {
            redirectUri: `${window.location.origin}/profile/security/bind-wechat/preview-callback`,
          },
        });
        if (begin.status === "ok") {
          window.location.href = begin.authorizeUrl;
          return;
        }
      }
      await navigate({
        search: { failed: true },
        to: "/profile/security/bind-wechat",
        viewTransition: false,
      });
    });
  }, [navigate, queryClient, search]);

  return <PendingLanding />;
}

import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { setMergeFlow } from "./merge-flow";
import { PendingLanding } from "./pending-landing";
import { previewMergeWechatFn } from "./security.fn";

// The preview leg's landing: the offer verdict routes onward — a mergeable
// or blocked conflict hops to the confirm page (BLOCKED's pane is the
// manual path), a SELF verdict is already bound, everything else renders
// the occupied card on the bind page.
export function BindWechatPreviewCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const search = useSearch({ from: "/profile/security/bind-wechat/preview-callback" });
  const started = useRef(false);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code, state } = search;
    if (code === undefined || state === undefined) {
      void navigate({
        search: { failed: true },
        to: "/profile/security/bind-wechat",
        viewTransition: false,
      });
      return;
    }
    void previewMergeWechatFn({ data: { code, state } }).then(async (res) => {
      if (res.status === "rejected") {
        await navigate({
          search: { failed: true },
          to: "/profile/security/bind-wechat",
          viewTransition: false,
        });
        return;
      }
      if (
        res.offer.status === "KEEP_CALLER" ||
        res.offer.status === "KEEP_OTHER" ||
        res.offer.status === "BLOCKED"
      ) {
        setMergeFlow({ offer: res.offer });
        await navigate({ to: "/profile/security/merge", viewTransition: false });
        return;
      }
      if (res.offer.status === "SELF") {
        toast.info(m.security_bind_phone_self());
        await queryClient.invalidateQueries();
        await navigate({ to: "/profile/security", viewTransition: false });
        return;
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

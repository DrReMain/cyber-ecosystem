import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { m } from "#/paraglide/messages";
import { mergeFailureCopy } from "./merge-failure";
import { mergeWechatFn } from "./security.fn";

// The direct merge's execution landing: a fresh OAuth leg minted on the
// surviving session proves the wechat, the merge runs here, the success
// page closes.
export function MergeBindCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const search = useSearch({ from: "/profile/security/merge/bind-callback" });
  const [reason, setReason] = useState<string | undefined>(undefined);
  const started = useRef(false);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code, state } = search;
    if (code === undefined || state === undefined) {
      setReason("");
      return;
    }
    void mergeWechatFn({ data: { code, state } }).then(async (res) => {
      if (res.status === "rejected") {
        setReason(res.reason);
        return;
      }
      await queryClient.invalidateQueries();
      await navigate({ to: "/profile/security/merge/success", viewTransition: false });
    });
  }, [navigate, queryClient, search]);

  if (reason !== undefined) {
    return (
      <main className="safe-px safe-pt min-h-app flex flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{mergeFailureCopy(reason)}</p>
        <Link className="text-ink-secondary text-sm underline" to="/profile/security">
          {m.common_back()}
        </Link>
      </main>
    );
  }
  return (
    <main className="safe-px safe-pt min-h-app flex items-center justify-center">
      <p className="text-ink-tertiary">{m.security_merge_pending()}</p>
    </main>
  );
}

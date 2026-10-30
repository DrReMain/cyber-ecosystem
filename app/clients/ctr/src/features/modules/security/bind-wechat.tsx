import { useNavigate, useSearch } from "@tanstack/react-router";
import { MessageCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { beginWechatLoginFn } from "#/features/modules/login/wechat.fn";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { beginBindWechatFn } from "./security.fn";

// Prototype 32: the wechat bind page is a pitch card plus one green launch
// button — the OAuth leg and its verdicts live on the callback routes.
// Outside the wechat browser the button stays honest about why it dims.
export function BindWechatPage() {
  const navigate = useNavigate();
  const { failed, from } = useSearch({ from: "/profile/security/bind-wechat/" });
  const [pending, setPending] = useState(false);
  const [wechatEnv, setWechatEnv] = useState(false);

  useEffect(() => {
    setWechatEnv(/MicroMessenger/i.test(navigator.userAgent));
  }, []);

  async function onBind() {
    if (!wechatEnv) {
      toast.error(m.login_error_wechat_env());
      return;
    }
    setPending(true);
    try {
      if (failed) {
        // The occupied card's exit is the conflict copy's own suggestion:
        // go log in AS that wechat, on the stock login leg (its fn throws
        // on failure, no verdict wrapper).
        const res = await beginWechatLoginFn({
          data: { redirectUri: `${window.location.origin}/login/wechat/callback` },
        });
        window.location.href = res.authorizeUrl;
        return;
      }
      const res = await beginBindWechatFn({
        data: { redirectUri: `${window.location.origin}/profile/security/bind-wechat/callback` },
      });
      if (res.status === "rejected") {
        toast.error(m.bind_wechat_begin_failed());
        setPending(false);
        return;
      }
      window.location.href = res.authorizeUrl;
    } catch {
      toast.error(m.bind_wechat_begin_failed());
      setPending(false);
    }
  }

  return (
    <AmbientLayout>
      <FlowHeader
        onBack={() => void navigate({ to: "/profile/security" })}
        title={m.bind_wechat_title()}
      />
      <div className="flex w-full flex-col gap-4 pt-5">
        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-5 text-center">
          <span className="bg-cta-success mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-xl">
            <MessageCircle aria-hidden className="text-white" size={26} />
          </span>
          <div className="text-ink text-base font-extrabold">{m.bind_wechat_title()}</div>
          <div className="text-ink-secondary mx-auto mt-2 max-w-[300px] text-[12.5px] leading-[1.8]">
            {m.bind_wechat_pitch()}
          </div>
        </div>

        {failed ? (
          <div className="rounded-card border border-[rgba(220,38,38,0.55)] bg-[rgba(127,29,29,0.22)] px-3.5 py-3">
            <p className="text-danger-text text-[12px] leading-[1.8]">
              <b>{m.bind_wechat_conflict_head()}</b>
              {m.bind_wechat_conflict_body()}
            </p>
          </div>
        ) : null}

        <div className="text-ink-tertiary text-[11.5px] leading-[1.8]">
          {m.bind_wechat_entry_line({
            entry:
              from === "login"
                ? m.bind_entry_login()
                : from === "guide"
                  ? m.bind_entry_guide()
                  : m.bind_entry_me(),
          })}
        </div>
        <button
          className={`bg-cta-success shadow-cta-success rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50 ${
            wechatEnv ? "" : "opacity-55"
          }`}
          disabled={pending}
          onClick={() => void onBind()}
          type="button"
        >
          {pending
            ? m.login_pending()
            : failed
              ? m.bind_wechat_conflict_cta()
              : m.bind_wechat_submit()}
        </button>
        {failed ? (
          <div className="text-ink-tertiary text-center text-[11px] leading-[1.7]">
            {m.bind_wechat_conflict_note()}
          </div>
        ) : null}
        <button
          className="border-line text-ink-secondary bg-surface-input rounded-control border-[1.5px] px-4 py-3.5 text-[15px] font-medium active:opacity-60"
          onClick={() => {
            toast.info(m.bind_wechat_skipped());
            void navigate({
              to: from === "me" ? "/profile/security" : "/",
              viewTransition: false,
            });
          }}
          type="button"
        >
          {m.bind_wechat_skip()}
        </button>
        <div className="text-ink-tertiary text-center text-[11px] leading-[1.7]">
          {m.bind_wechat_skip_note()}
        </div>
      </div>
    </AmbientLayout>
  );
}

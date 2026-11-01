import { useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { beginWechatLoginFn } from "../wechat.fn";

// The WeChat leg: snsapi authorize needs the WeChat browser, so outside it
// the green button dims and the note says what tapping it does (prototype
// envNote). The authorize hop never puts a query on the redirect_uri —
// WeChat's validation rejects embedded queries, so the post-login target
// is always derived afresh.
export function WechatButton({
  wechatEnv,
  agreeOK,
  onError,
}: Readonly<{
  wechatEnv: boolean;
  agreeOK: () => boolean;
  onError: () => void;
}>) {
  const [pending, setPending] = useState(false);

  async function onWechat() {
    if (!wechatEnv) {
      toast.error(m.login_error_wechat_env());
      return;
    }
    if (!agreeOK()) {
      return;
    }
    setPending(true);
    try {
      const redirectUri = `${window.location.origin}/login/wechat/callback`;
      const res = await beginWechatLoginFn({ data: { redirectUri } });
      window.location.href = res.authorizeUrl;
    } catch {
      setPending(false);
      onError();
    }
  }

  return (
    <>
      <Button
        className={`bg-cta-success shadow-cta-success rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50 ${
          wechatEnv ? "" : "opacity-55"
        }`}
        isDisabled={pending}
        onPress={onWechat}
      >
        {pending ? m.login_pending() : m.login_wechat_button()}
      </Button>
      {!wechatEnv && (
        <div className="text-ink-tertiary text-center text-[11px] leading-[1.6]">
          {m.login_wechat_env_note()}
        </div>
      )}
    </>
  );
}

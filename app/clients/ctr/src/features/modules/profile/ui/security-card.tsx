import { Lock, MessageCircle, Smartphone } from "lucide-react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "rounded-lg border border-line bg-surface-card p-4";

/**
 * Account-security rows whose faces are contract-gated (SMS provider, H2):
 * every row stays visible with an honest not-yet-open state.
 */
export function SecurityCard() {
  const rows = [
    { Icon: MessageCircle, key: "wechat", label: m.profile_row_wechat() },
    { Icon: Smartphone, key: "phone", label: m.profile_row_phone() },
    { Icon: Lock, key: "password", label: m.profile_row_password() },
  ];
  return (
    <div className={CARD_CLASS}>
      <div className="mb-1 font-extrabold text-ink text-sm">{m.profile_card_account()}</div>
      {rows.map((row, index) => (
        <div
          className={`flex items-center justify-between py-3 ${
            index < rows.length - 1 ? "border-line-soft border-b" : ""
          }`}
          key={row.key}
        >
          <span className="flex items-center gap-2 font-bold text-ink text-sm">
            <row.Icon aria-hidden size={16} />
            {row.label}
          </span>
          <button
            className="tap-zone rounded-full bg-status-soon-bg px-2.5 py-1 text-[10px] text-status-soon active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            {m.profile_status_not_open()}
          </button>
        </div>
      ))}
    </div>
  );
}

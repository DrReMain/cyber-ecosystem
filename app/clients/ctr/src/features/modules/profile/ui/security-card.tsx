import { Lock, MessageCircle, Smartphone } from "lucide-react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "rounded-lg border border-line bg-surface-card p-4";

/**
 * Account-security rows whose faces are contract-gated: every row stays
 * visible with an honest not-yet-open state.
 */
export function SecurityCard() {
  const rows = [
    { Icon: MessageCircle, key: "wechat", label: m.profile_row_wechat() },
    { Icon: Smartphone, key: "phone", label: m.profile_row_phone() },
    { Icon: Lock, key: "password", label: m.profile_row_password() },
  ];
  return (
    <div className={CARD_CLASS}>
      <div className="text-ink mb-1 text-sm font-extrabold">{m.profile_card_account()}</div>
      {rows.map((row, index) => (
        <div
          className={`flex items-center justify-between py-3 ${
            index < rows.length - 1 ? "border-line-soft border-b" : ""
          }`}
          key={row.key}
        >
          <span className="text-ink flex items-center gap-2 text-sm font-bold">
            <row.Icon aria-hidden size={16} />
            {row.label}
          </span>
          <button
            className="tap-zone bg-status-soon-bg text-status-soon rounded-full px-2.5 py-1 text-[10px] active:opacity-60"
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

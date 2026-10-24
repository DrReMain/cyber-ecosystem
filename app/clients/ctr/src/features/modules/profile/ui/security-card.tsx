import { Lock, MessageCircle, Smartphone } from "lucide-react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]";

/**
 * Account-security rows whose faces are contract-gated: every row stays
 * visible with an honest not-yet-open state.
 */
export function SecurityCard() {
  const rows = [
    {
      Icon: MessageCircle,
      chip: "bg-[rgba(34,197,94,0.16)]",
      key: "wechat",
      label: m.profile_row_wechat(),
    },
    {
      Icon: Smartphone,
      chip: "bg-[rgba(37,99,235,0.16)]",
      key: "phone",
      label: m.profile_row_phone(),
    },
    {
      Icon: Lock,
      chip: "bg-[rgba(107,114,128,0.18)]",
      key: "password",
      label: m.profile_row_password(),
    },
  ];
  return (
    <div className={CARD_CLASS}>
      <div className="text-ink mb-[11px] text-[14.5px] font-extrabold">
        {m.profile_card_account()}
      </div>
      {rows.map((row, index) => (
        <div
          className={`flex items-center justify-between py-[9px] ${
            index < rows.length - 1 ? "border-b border-[rgba(55,65,81,0.4)]" : ""
          }`}
          key={row.key}
        >
          <span className="text-ink flex items-center gap-[9px] text-[13px] font-bold">
            <span
              className={`rounded-control flex h-[30px] w-[30px] shrink-0 items-center justify-center ${row.chip}`}
            >
              <row.Icon aria-hidden size={14} />
            </span>
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

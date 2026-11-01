import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { systemReason } from "#/domains/error";
import { m } from "#/paraglide/messages";

// One error line for the whole form, as the prototype draws it: field
// verdicts first, then the server's reason story. The seat-required
// verdict is the one branch with a next action — the switch-account hop.
export function FormErrorLine({
  copy,
  mode,
  serverReason,
  onSwitchAccount,
}: Readonly<{
  copy: string | null;
  mode: "code" | "password";
  serverReason: string | undefined;
  onSwitchAccount: () => void;
}>) {
  if (copy === null) {
    return null;
  }
  return (
    <div className="flex items-start justify-between gap-2" role="alert">
      <p className="text-danger-text text-[11.5px] leading-snug">{copy}</p>
      {mode === "password" && systemReason(serverReason) === System.LOGIN_SEAT_REQUIRED && (
        <button
          className="tap-zone text-ink-secondary shrink-0 font-bold active:opacity-60"
          onClick={onSwitchAccount}
          type="button"
        >
          {m.login_error_switch_account()}
        </button>
      )}
    </div>
  );
}

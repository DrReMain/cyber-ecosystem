import { Input, TextField } from "react-aria-components";
import { m } from "#/paraglide/messages";
import type { SmsLogin } from "../sms";

export const inputClass =
  "w-full rounded-control border border-line bg-surface-input px-3.5 py-3.25 text-[15px] text-ink outline-none placeholder:text-ink-tertiary focus:border-focus";
export const badInputClass = inputClass.replace("border-line", "border-danger-text");

// The prototype's outline-blue send rail: fixed width, full row height,
// dimming while the countdown owns it.
const sendBtnClass =
  "border-focus text-link bg-transparent rounded-control w-29 shrink-0 border px-3 text-[13px] font-bold active:opacity-60 disabled:opacity-60";

// The code tab's field group: phone with the prototype's live verdict, and
// the code row — input plus the send rail on one line. Pure presentational:
// the state machine lives in the useSmsLogin hook the page wires in, and the
// send button fires the page's gated onSend (the agree gate wraps it).
export function SmsFields({ onSend, sms }: { onSend: () => void; sms: SmsLogin }) {
  return (
    <>
      {/* aria-label, not a visually-hidden Label: sr-only text still leaks
          into innerText, and the prototype's fields are label-less. */}
      <TextField aria-label={m.login_phone_label()} autoComplete="tel" isRequired name="phone">
        <Input
          className={sms.phoneErr ? badInputClass : inputClass}
          inputMode="numeric"
          maxLength={11}
          onBlur={sms.blurPhone}
          onChange={(event) => sms.setPhone(event.target.value)}
          placeholder={m.login_phone_placeholder()}
          value={sms.phone}
        />
      </TextField>
      <div className="flex items-stretch gap-3">
        <TextField
          aria-label={m.login_code_label()}
          autoComplete="one-time-code"
          className="flex-1"
          isRequired
          name="code"
        >
          <Input
            className={sms.codeErr ? badInputClass : inputClass}
            inputMode="numeric"
            maxLength={6}
            onChange={(event) => sms.setCode(event.target.value)}
            placeholder={m.login_code_placeholder()}
            value={sms.code}
          />
        </TextField>
        <button
          className={sendBtnClass}
          disabled={sms.countdown > 0 || sms.sending}
          onClick={onSend}
          type="button"
        >
          {sms.countdown > 0
            ? m.login_code_resend_in({ seconds: sms.countdown })
            : m.login_code_send()}
        </button>
      </div>
    </>
  );
}

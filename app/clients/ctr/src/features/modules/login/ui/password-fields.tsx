import { Input, TextField } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { inputClass } from "./sms-fields";

// The password leg's two fields. aria-label, not a visually-hidden Label:
// sr-only text still leaks into innerText, and the prototype's fields are
// label-less.
export function PasswordFields() {
  return (
    <>
      <TextField
        aria-label={m.login_account_label()}
        autoComplete="username"
        isRequired
        name="account"
        type="text"
      >
        <Input className={inputClass} placeholder={m.login_account_placeholder()} />
      </TextField>
      <TextField
        aria-label={m.login_password_label()}
        autoComplete="current-password"
        isRequired
        name="password"
        type="password"
      >
        <Input className={inputClass} maxLength={20} placeholder={m.login_password_placeholder()} />
      </TextField>
    </>
  );
}

import { Link } from "@tanstack/react-router";
import { CheckboxButton, CheckboxField } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { AgreeBox } from "./agree-box";

export interface AgreeState {
  privacy: boolean;
  user: boolean;
}

// The prototype's consent gate, read top-to-bottom as it draws: user row,
// select-all (on the user row's line, right edge), privacy row, then the
// minor note indented under the checkbox column. The checked state is the
// claim; each login leg refuses to fire while a box is still empty.
export function AgreeGate({
  agree,
  missing,
  onUser,
  onPrivacy,
  onAll,
}: Readonly<{
  agree: AgreeState;
  missing: AgreeState;
  onUser: (selected: boolean) => void;
  onPrivacy: (selected: boolean) => void;
  onAll: (selected: boolean) => void;
}>) {
  const rowClass = (field: "user" | "privacy"): string =>
    `flex items-start gap-2.5 py-1.5 ${missing[field] ? "text-danger-text" : "text-ink-muted"}`;

  const agreeRow = (field: "user" | "privacy", label: string, to: string) => (
    <CheckboxField
      isSelected={agree[field]}
      onChange={(selected) => (field === "user" ? onUser : onPrivacy)(selected)}
    >
      <CheckboxButton>
        <div className={rowClass(field)}>
          <AgreeBox checked={agree[field]} />
          <span className="text-xs leading-[1.6]">
            {m.login_agree_prefix()}
            {` `}
            <Link
              className="text-link active:opacity-60"
              onClick={(event) => event.stopPropagation()}
              to={to}
            >
              {label}
            </Link>
          </span>
        </div>
      </CheckboxButton>
    </CheckboxField>
  );

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-start justify-between gap-2.5">
        {agreeRow("user", m.login_agree_user_name(), "/agreement/user")}
        <CheckboxField
          isIndeterminate={agree.user !== agree.privacy}
          isSelected={agree.user && agree.privacy}
          onChange={onAll}
        >
          <CheckboxButton>
            <div className="text-ink-secondary flex items-center gap-2 font-bold whitespace-nowrap">
              <AgreeBox
                checked={agree.user && agree.privacy}
                indeterminate={agree.user !== agree.privacy}
              />
              <span className="text-sm">{m.login_agree_all()}</span>
            </div>
          </CheckboxButton>
        </CheckboxField>
      </div>
      {agreeRow("privacy", m.login_agree_privacy_name(), "/agreement/privacy")}
      <span className="text-ink-tertiary ml-6.25 text-[10.5px] leading-snug">
        {m.login_agree_minor_note()}
      </span>
    </div>
  );
}

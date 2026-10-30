import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";

// The login card's static footer: forgot-password (the SMS reset leg), the
// agreement links, and the company lines. Zero state of its own.
export function LoginFooter() {
  return (
    <>
      <div className="text-ink-tertiary mt-2 text-center text-[10.5px] leading-[1.95]">
        <Link className="tap-zone text-link py-1 active:opacity-60" to="/login/forgot">
          {m.login_forgot()}
        </Link>
        <span className="mx-1">{" · "}</span>
        <Link className="tap-zone text-link py-1 active:opacity-60" to="/agreement/user">
          {m.login_footer_user()}
        </Link>
        <span className="mx-1">{" · "}</span>
        <Link className="tap-zone text-link py-1 active:opacity-60" to="/agreement/privacy">
          {m.login_footer_privacy()}
        </Link>
        <span className="mx-1">{" · "}</span>
        <Link className="tap-zone text-link py-1 active:opacity-60" to="/">
          {m.login_footer_home()}
        </Link>
      </div>
      <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.95]">
        {m.login_footer_company()}
        {` · `}
        {m.login_footer_icp()}
        <br />
        {m.login_footer_email()}
      </div>
    </>
  );
}

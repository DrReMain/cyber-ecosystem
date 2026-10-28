import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";

/** The filing footer (prototype ⑦, login-footer shape): company, ICP
 * number, support line, and the two agreements — the report's own keys
 * per the usage-prefix rule, the number itself the ruling's string. */
export function FilingFooter() {
  return (
    <div className="border-t border-[rgba(55,65,81,0.4)] pt-2.5 text-center">
      <div className="text-ink-tertiary text-[9px] leading-[1.8]">
        {m.report_filing_company()} · {m.report_filing_icp()}
        <br />
        {m.report_filing_email()} ·{" "}
        <Link className="tap-zone text-link active:opacity-60" to="/agreement/user">
          {m.report_filing_user()}
        </Link>{" "}
        ·{" "}
        <Link className="tap-zone text-link active:opacity-60" to="/agreement/privacy">
          {m.report_filing_privacy()}
        </Link>
      </div>
    </div>
  );
}

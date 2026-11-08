import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";

/** The pay filing footer (one of the four filing seats): the tap-to-pay
 * consent line over the two agreements, then the company, ICP number,
 * and support line — the pay page's own keys per the usage-prefix rule,
 * the legal identifiers verbatim (locale-stable by nature, the same
 * values the report seat carries). */
export function PayFilingFooter() {
  return (
    <div className="text-status-off text-center text-[9px] leading-[1.8]">
      {m.pay_filing_lead()}{" "}
      <Link className="tap-zone text-status-lock active:opacity-60" to="/agreement/user">
        {m.pay_filing_user()}
      </Link>{" "}
      {m.pay_filing_and()}{" "}
      <Link className="tap-zone text-status-lock active:opacity-60" to="/agreement/privacy">
        {m.pay_filing_privacy()}
      </Link>
      <br />
      <span className="text-status-lock">
        {m.pay_filing_company()} · {m.pay_filing_icp()} · {m.pay_filing_email()}
      </span>
    </div>
  );
}

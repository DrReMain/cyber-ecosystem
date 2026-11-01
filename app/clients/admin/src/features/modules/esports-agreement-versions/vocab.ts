import { AgreementDoc } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_pb";
import { m } from "#/paraglide/messages";

// The agreement domain's shared vocabulary — this module owns it; the
// consents page imports from here rather than forking the doc/method
// word tables (the same single-source law as the report-type options).
export const docOptions = [
  { label: () => m.esports_agreement_doc_user(), value: "AGREEMENT_DOC_USER" },
  { label: () => m.esports_agreement_doc_privacy(), value: "AGREEMENT_DOC_PRIVACY" },
];

// The URL and tables carry wire names; requests carry the generated
// enum — this module owns the one translation.
export const docFromWire = (wire?: string): AgreementDoc | undefined => {
  switch (wire) {
    case "AGREEMENT_DOC_USER":
      return AgreementDoc.USER;
    case "AGREEMENT_DOC_PRIVACY":
      return AgreementDoc.PRIVACY;
    default:
      return undefined;
  }
};

export const loginMethodLabel = (wire: string | undefined): string => {
  switch (wire) {
    case "AGREEMENT_LOGIN_METHOD_PASSWORD":
      return m.esports_agreement_login_password();
    case "AGREEMENT_LOGIN_METHOD_WECHAT":
      return m.esports_agreement_login_wechat();
    default:
      return m.esports_agreement_login_sms_code();
  }
};

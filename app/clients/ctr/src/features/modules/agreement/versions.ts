export type AgreementDoc = "user" | "privacy";

export const AGREEMENT_DOC_META: Record<AgreementDoc, { effectiveDate: string; no: string }> = {
  privacy: { effectiveDate: "2026-08-19", no: "v1.0.0" },
  user: { effectiveDate: "2026-08-19", no: "v1.0.0" },
};

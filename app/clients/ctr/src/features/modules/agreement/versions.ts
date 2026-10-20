// Placeholder agreement-version metadata, transcribed from the prototype's
// _state.js AGREEMENTS. The source of truth is the server's
// GET /agreements/current (openspec agreement-consent — the registered
// contract gap, progress.md E6b): a user must consent to the exact version
// they saw, so the page never bakes these into copy keys. H2 carries the
// display slot only; the version param arrives with the consent contract.

export type AgreementDoc = "user" | "privacy";

export const AGREEMENT_DOC_META: Record<AgreementDoc, { effectiveDate: string; no: string }> = {
  privacy: { effectiveDate: "2026-08-19", no: "v1.0.0" },
  user: { effectiveDate: "2026-08-19", no: "v1.0.0" },
};

import { FeedbackStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_pb";
import { m } from "#/paraglide/messages";

// The feedback domain's shared vocabulary — this module owns it; the
// detail page imports from here rather than forking the word tables
// (the agreement doc-options precedent).
export const statusTags: Record<string, { color: string; label: () => string }> = {
  FEEDBACK_STATUS_PENDING: { color: "warning", label: () => m.esports_feedback_status_pending() },
  FEEDBACK_STATUS_REPLIED: { color: "success", label: () => m.esports_feedback_status_replied() },
  FEEDBACK_STATUS_CLOSED: { color: "default", label: () => m.esports_feedback_status_closed() },
};

const categoryLabels: Record<string, () => string> = {
  FEEDBACK_CATEGORY_FEATURE_BUG: () => m.esports_feedback_cat_feature_bug(),
  FEEDBACK_CATEGORY_ASSESSMENT: () => m.esports_feedback_cat_assessment(),
  FEEDBACK_CATEGORY_REPORT: () => m.esports_feedback_cat_report(),
  FEEDBACK_CATEGORY_PAYMENT: () => m.esports_feedback_cat_payment(),
  FEEDBACK_CATEGORY_ACCOUNT: () => m.esports_feedback_cat_account(),
  FEEDBACK_CATEGORY_OTHER: () => m.esports_feedback_cat_other(),
};

export const categoryLabel = (wire?: string): string =>
  categoryLabels[String(wire ?? "")]?.() ?? "-";

// The URL segment and the tabs carry short names; requests carry the
// generated enum — this module owns the one translation.
export const TAB_TO_STATUS: Record<string, FeedbackStatus | undefined> = {
  all: undefined,
  pending: FeedbackStatus.PENDING,
  replied: FeedbackStatus.REPLIED,
  closed: FeedbackStatus.CLOSED,
};

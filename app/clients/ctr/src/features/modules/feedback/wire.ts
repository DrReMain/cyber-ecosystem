import type {
  Feedback,
  FeedbackMessage,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_pb";
import { m } from "#/paraglide/messages";

// Runtime string enums — the transport delivers enums as value-name
// strings while the generated types keep numeric enums (same law as the
// pay module's order wire).
export const CATEGORY_FEATURE_BUG = "FEEDBACK_CATEGORY_FEATURE_BUG";
export const CATEGORY_ASSESSMENT = "FEEDBACK_CATEGORY_ASSESSMENT";
export const CATEGORY_REPORT = "FEEDBACK_CATEGORY_REPORT";
export const CATEGORY_PAYMENT = "FEEDBACK_CATEGORY_PAYMENT";
export const CATEGORY_ACCOUNT = "FEEDBACK_CATEGORY_ACCOUNT";
export const CATEGORY_OTHER = "FEEDBACK_CATEGORY_OTHER";

export const STATUS_PENDING = "FEEDBACK_STATUS_PENDING";
export const STATUS_REPLIED = "FEEDBACK_STATUS_REPLIED";
export const STATUS_CLOSED = "FEEDBACK_STATUS_CLOSED";

export const SENDER_USER = "FEEDBACK_MESSAGE_SENDER_USER";
export const SENDER_ADMIN = "FEEDBACK_MESSAGE_SENDER_ADMIN";

// The ticket root as the wire delivers it — the list row and the detail
// header both read this shape.
export interface WireFeedback {
  id?: string;
  category?: string;
  content?: string;
  contact?: string;
  status?: string;
  createdAt?: string;
  updatedAt?: string;
}

export function wireFeedbackOf(row: Feedback): WireFeedback {
  return {
    id: row.id ?? undefined,
    category: String(row.category),
    content: row.content ?? undefined,
    contact: row.contact ?? undefined,
    status: String(row.status),
    createdAt: typeof row.createdAt === "string" ? row.createdAt : undefined,
    updatedAt: typeof row.updatedAt === "string" ? row.updatedAt : undefined,
  };
}

// One conversation turn, oldest first — sender decides which side of the
// thread the bubble labels.
export interface WireFeedbackMessage {
  id?: string;
  sender?: string;
  content?: string;
  createdAt?: string;
}

export function wireFeedbackMessageOf(row: FeedbackMessage): WireFeedbackMessage {
  return {
    id: row.id ?? undefined,
    sender: String(row.sender),
    content: row.content ?? undefined,
    createdAt: typeof row.createdAt === "string" ? row.createdAt : undefined,
  };
}

// The category's word, keyed by its runtime value name — the one map the
// submit chips and the conversation header both read, so a new contract
// category gains its word in exactly one place. Unknown reads the honest
// empty string, never a wrong category's word.
export function categoryWordOf(category: string | undefined): string {
  switch (category) {
    case CATEGORY_FEATURE_BUG:
      return m.feedback_cat_feature_bug();
    case CATEGORY_ASSESSMENT:
      return m.feedback_cat_assessment();
    case CATEGORY_REPORT:
      return m.feedback_cat_report();
    case CATEGORY_PAYMENT:
      return m.feedback_cat_payment();
    case CATEGORY_ACCOUNT:
      return m.feedback_cat_account();
    case CATEGORY_OTHER:
      return m.feedback_cat_other();
    default:
      return "";
  }
}

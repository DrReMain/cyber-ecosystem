import type { Message } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_pb";

// Runtime string enums — the transport delivers enums as value-name
// strings while the generated types keep numeric enums; these constants
// restate what actually arrives (the pay module's order wire set the
// pattern). Announcement stays unnamed: every consumer's default arm is
// the announcement face, the fallback for unrecognized kinds.
export const TYPE_REPORT = "MESSAGE_TYPE_REPORT";
export const TYPE_MEMBERSHIP = "MESSAGE_TYPE_MEMBERSHIP";
export const TYPE_FEEDBACK_REPLY = "MESSAGE_TYPE_FEEDBACK_REPLY";

// One letterbox row as the wire actually delivers it: wrapper fields
// collapsed to plain strings, the enum as its value name, the timestamp
// as a JSON string.
export interface WireMessage {
  id?: string;
  type?: string;
  title?: string;
  content?: string;
  refId?: string;
  isRead?: boolean;
  createdAt?: string;
}

export function wireMessageOf(row: Message): WireMessage {
  return {
    id: row.id ?? undefined,
    type: String(row.type),
    title: row.title ?? undefined,
    content: row.content ?? undefined,
    refId: row.refId ?? undefined,
    isRead: row.isRead ?? undefined,
    createdAt: typeof row.createdAt === "string" ? row.createdAt : undefined,
  };
}

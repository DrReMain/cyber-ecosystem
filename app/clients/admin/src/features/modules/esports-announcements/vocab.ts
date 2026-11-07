import {
  AnnouncementKind,
  AnnouncementScope,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin_pb";
import { m } from "#/paraglide/messages";

// The announcement domain's vocabulary — wire names in, copy out; this
// module is the single translation table for all three enums.
export const kindTags: Record<string, { color: string; label: () => string }> = {
  ANNOUNCEMENT_KIND_MESSAGE: {
    color: "processing",
    label: () => m.esports_announcements_kind_message(),
  },
  ANNOUNCEMENT_KIND_BANNER: {
    color: "warning",
    label: () => m.esports_announcements_kind_banner(),
  },
};

export const scopeLabels: Record<string, () => string> = {
  ANNOUNCEMENT_SCOPE_ALL: () => m.esports_announcements_scope_all(),
  ANNOUNCEMENT_SCOPE_MEMBERSHIP: () => m.esports_announcements_scope_membership(),
};

export const statusTags: Record<string, { color: string; label: () => string }> = {
  ANNOUNCEMENT_STATUS_DRAFT: {
    color: "default",
    label: () => m.esports_announcements_status_draft(),
  },
  ANNOUNCEMENT_STATUS_PUBLISHED: {
    color: "success",
    label: () => m.esports_announcements_status_published(),
  },
  ANNOUNCEMENT_STATUS_OFFLINE: {
    color: "default",
    label: () => m.esports_announcements_status_offline(),
  },
};

export const kindOptions = [
  { label: () => m.esports_announcements_kind_message(), value: AnnouncementKind.MESSAGE },
  { label: () => m.esports_announcements_kind_banner(), value: AnnouncementKind.BANNER },
];

export const scopeOptions = [
  { label: () => m.esports_announcements_scope_all(), value: AnnouncementScope.ALL },
  {
    label: () => m.esports_announcements_scope_membership(),
    value: AnnouncementScope.MEMBERSHIP,
  },
];

export const kindLabel = (wire?: string): string => kindTags[String(wire ?? "")]?.label() ?? "-";
export const scopeLabel = (wire?: string): string => scopeLabels[String(wire ?? "")]?.() ?? "-";

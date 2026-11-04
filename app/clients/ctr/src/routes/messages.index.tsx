import { createFileRoute } from "@tanstack/react-router";
import { MessagesPage } from "#/features/modules/messages/page";

// The message center (prototype message-center): the tab-owning letterbox
// page — 消息 and 我的反馈 in one URL, the tab as a search param like the
// prototype's ?t=. Head and guard come from the layout parent; the page
// reads its own search.
export const Route = createFileRoute("/messages/")({
  validateSearch: (search: Record<string, unknown>): { t: "msg" | "fb" } => ({
    t: search.t === "fb" ? "fb" : "msg",
  }),
  component: MessagesPage,
});

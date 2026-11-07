import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { FeedbackTicket } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import {
  closeFeedback,
  listFeedbacks,
  replyFeedback,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin-FeedbackAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Card, Tabs, Typography } from "antd";
import { useState } from "react";
import { toast } from "sonner";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parseTicketsSearch, ticketsSearchSchema } from "./search";
import { ReplyModal } from "./ui/reply-modal";
import { TicketsTable } from "./ui/tickets-table";
import { TAB_TO_STATUS } from "./vocab";

const AREA = "/dashboard/feedback/tickets/";

export function FeedbackTicketsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseTicketsSearch);
  const navigate = useNavigate();
  const [replying, setReplying] = useState<FeedbackTicket | null>(null);

  const patchSearch = (patch: Record<string, unknown>) => {
    void navigate({ to: "/dashboard/feedback/tickets", search: { ...search, ...patch } });
  };

  const listQuery = useQuery(listFeedbacks, {
    status: TAB_TO_STATUS[search.status ?? "all"],
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });

  const rowName = (row?: FeedbackTicket): string => row?.nickname ?? row?.userId ?? "-";

  // The answer is one move with two effects — the ticket flips and the
  // member's letterbox gains the letter in the same transaction, so the
  // toast can promise the C-end sync the backend already guaranteed.
  const replyMut = useMutation(replyFeedback, {
    onSuccess: (_resp, vars) => {
      const name = replying ? rowName(replying) : (vars.feedbackId ?? "");
      setReplying(null);
      void toast.success(m.esports_feedback_tickets_reply_ok({ user: name }));
      void listQuery.refetch();
    },
  });
  const closeMut = useMutation(closeFeedback, {
    onSuccess: (_resp, vars) => {
      const row = (listQuery.data?.tickets ?? []).find((t) => t.id === vars.feedbackId);
      void toast.success(
        m.esports_feedback_tickets_close_ok({ user: row ? rowName(row) : (vars.feedbackId ?? "") }),
      );
      void listQuery.refetch();
    },
  });

  const pagination = useServerPagination(store, ticketsSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.esports_feedback_tickets_total({ n: t }),
  });

  return (
    <div className="flex flex-col gap-3">
      <div>
        <Typography.Title className="!mb-0" level={4}>
          {m.esports_feedback_tickets_title()}
        </Typography.Title>
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_feedback_tickets_desc()}
        </Typography.Text>
      </div>

      <Tabs
        activeKey={search.status ?? "all"}
        items={[
          { key: "all", label: m.esports_feedback_tickets_filter_all() },
          { key: "pending", label: m.esports_feedback_tickets_filter_pending() },
          { key: "replied", label: m.esports_feedback_tickets_filter_replied() },
          { key: "closed", label: m.esports_feedback_tickets_filter_closed() },
        ]}
        onChange={(key) =>
          patchSearch({
            status: key === "all" ? undefined : (key as "pending" | "replied" | "closed"),
            pageNo: undefined,
          })
        }
      />

      <Card>
        <TicketsTable
          loading={listQuery.isLoading}
          onClose={(row) => closeMut.mutate({ feedbackId: row.id ?? "" })}
          onRefresh={() => void listQuery.refetch()}
          onReply={setReplying}
          pagination={pagination}
          rows={listQuery.data?.tickets ?? []}
        />
      </Card>

      <ReplyModal
        onClose={() => setReplying(null)}
        onSubmit={(content) =>
          replying && void replyMut.mutate({ feedbackId: replying.id ?? "", content })
        }
        open={replying !== null}
        submitting={replyMut.isPending}
        ticket={replying}
      />
    </div>
  );
}

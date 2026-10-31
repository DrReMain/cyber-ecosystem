import { useQuery } from "@connectrpc/connect-query";
import { getPaymentOrder } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin-OrderAdminService_connectquery";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import { Alert, Button, Card, Descriptions, Tag, Typography } from "antd";
import { ArrowLeft, RefreshCw } from "lucide-react";
import { dateOf, fullDate, shortOf } from "#/features/modules/esports-members/ui/dates";
import { stateTags } from "#/features/modules/esports-orders/ui/orders-table";
import { fenToYuan } from "#/features/modules/esports-skus/ui/fen";
import { m } from "#/paraglide/messages";

export function EsportsOrderDetailPage() {
  const { id } = useParams({ from: "/dashboard/pricing/orders/$id" });
  const navigate = useNavigate();
  const detailQuery = useQuery(getPaymentOrder, { orderNo: id });
  const refetch = () => detailQuery.refetch();

  const d = detailQuery.data;
  const o = d?.order;
  const state = stateTags[String(o?.order?.state ?? "")];
  const unlock = d?.unlock;

  const unlockLine = () => {
    if (!o) return "";
    if (String(o.order?.state) === "ORDER_STATE_CLOSED") return m.esports_order_unlock_closed();
    if (!unlock) return m.esports_order_unlock_pending();
    if (unlock.recordId) return m.esports_order_unlock_trial({ id: unlock.recordId });
    return m.esports_order_unlock_period({
      from: shortOf(dateOf(unlock.grantedAt)),
      to: shortOf(dateOf(unlock.expiresAt)),
      // protojson drops zero scalars; zero unlocks is a real state.
      n: unlock.unlockCount ?? 0,
    });
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <Button
          icon={<ArrowLeft size={14} />}
          onClick={() =>
            void navigate({ to: "/dashboard/pricing/orders", search: { pageNo: 1, pageSize: 20 } })
          }
          size="small"
          type="text"
        >
          {m.esports_order_detail_back()}
        </Button>
        <Button
          icon={<RefreshCw size={14} />}
          loading={detailQuery.isFetching}
          onClick={refetch}
          size="small"
          type="text"
        >
          {m.common_toolbar_refresh()}
        </Button>
      </div>

      <Card loading={detailQuery.isLoading}>
        <div className="mb-4 flex items-center gap-3">
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_order_detail_title()}
          </Typography.Title>
          <span className="font-mono text-sm">{id}</span>
          {state ? (
            <Tag className="m-0" color={state.color}>
              {state.label()}
            </Tag>
          ) : null}
        </div>
        <Descriptions bordered column={2} size="small">
          <Descriptions.Item label={m.esports_order_user()}>
            {o?.userId ? (
              <Link className="text-primary" params={{ id: o.userId }} to="/dashboard/people/users/$id">
                {[o?.nickname, o?.phoneMasked].filter(Boolean).join(" · ") || "-"}
              </Link>
            ) : (
              [o?.nickname, o?.phoneMasked].filter(Boolean).join(" · ") || "-"
            )}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_amount()}>
            <span className="font-semibold">{fenToYuan(o?.order?.amountFen)}</span>
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_sku()}>
            {o?.skuName || o?.order?.skuCode || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_channel()}>
            {o?.order?.channel || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_created()}>
            {fullDate(o?.order?.createdAt)}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_paid()}>
            {o?.order?.paidAt ? fullDate(o.order.paidAt) : "-"}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_txn()}>
            {o?.order?.transactionId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_order_unlock()}>{unlockLine()}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Alert message={m.esports_order_refund_note()} showIcon type="info" />
    </div>
  );
}

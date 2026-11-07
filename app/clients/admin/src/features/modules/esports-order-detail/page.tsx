import { useMutation, useQuery } from "@connectrpc/connect-query";
import type {
  AdminOrder,
  OrderRefundView,
  OrderUnlockView,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin_pb";
import {
  getPaymentOrder,
  refundPaymentOrder,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin-OrderAdminService_connectquery";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import { Alert, Button, Card, Descriptions, Input, Modal, Tag, Typography } from "antd";
import { ArrowLeft, RefreshCw, RotateCcw } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { dateOf, fullDate, shortOf } from "#/features/modules/esports-members/ui/dates";
import { refundTags, stateTags } from "#/features/modules/esports-orders/ui/orders-table";
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
          <Typography.Title className="mb-0!" level={4}>
            {m.esports_order_detail_title()}
          </Typography.Title>
          <span className="font-mono text-sm">{id}</span>
          {state ? (
            <Tag className="m-0" color={state.color}>
              {state.label()}
            </Tag>
          ) : null}
          {String(o?.order?.state) === "ORDER_STATE_PAID" && !o?.refund ? (
            <RefundAction amount={fenToYuan(o?.order?.amountFen)} onDone={refetch} orderNo={id} />
          ) : null}
        </div>
        <OrderFacts o={o} unlock={d?.unlock} />
        {o?.refund ? (
          <div className="mt-4">
            <RefundFacts refund={o.refund} />
          </div>
        ) : null}
      </Card>

      <Alert message={m.esports_order_detail_refund_note()} showIcon type="info" />
    </div>
  );
}

// The order's base fact table. Every row is a direct Descriptions.Item
// child — antd's Descriptions scans its immediate children, so rows never
// hide behind wrapper components.
function OrderFacts({ o, unlock }: Readonly<{ o?: AdminOrder; unlock?: OrderUnlockView }>) {
  return (
    <Descriptions bordered column={2} size="small">
      <Descriptions.Item label={m.esports_order_detail_user()}>
        {o?.userId ? (
          <Link className="text-primary" params={{ id: o.userId }} to="/dashboard/people/users/$id">
            {[o?.nickname, o?.phoneMasked].filter(Boolean).join(" · ") || "-"}
          </Link>
        ) : (
          [o?.nickname, o?.phoneMasked].filter(Boolean).join(" · ") || "-"
        )}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_amount()}>
        <span className="font-semibold">{fenToYuan(o?.order?.amountFen)}</span>
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_sku()}>
        {o?.skuName || o?.order?.skuCode || "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_channel()}>
        {o?.order?.channel || "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_created()}>
        {fullDate(o?.order?.createdAt)}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_paid()}>
        {o?.order?.paidAt ? fullDate(o.order.paidAt) : "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_txn()}>
        {o?.order?.transactionId || "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_unlock()}>
        {unlockLineOf(o, unlock)}
      </Descriptions.Item>
    </Descriptions>
  );
}

// The derived refund fact as its own table: state tag, the money that went
// back with the channel-side success time, the refund identifiers, and the
// operator's recorded reason.
function RefundFacts({ refund }: Readonly<{ refund: OrderRefundView }>) {
  const tag = refundTags[String(refund.state ?? "")];
  return (
    <Descriptions bordered column={2} size="small">
      <Descriptions.Item label={m.esports_order_detail_refund()} span={2}>
        <div className="flex flex-col gap-0.5">
          <span>
            {tag ? (
              <Tag className="m-0" color={tag.color}>
                {tag.label()}
              </Tag>
            ) : (
              String(refund.state ?? "")
            )}
            <span className="ms-2 font-semibold">{fenToYuan(refund.amountFen)}</span>
            {refund.refundedAt ? (
              <span className="ms-2 text-xs">{fullDate(refund.refundedAt)}</span>
            ) : null}
          </span>
          {refund.refundNo ? (
            <span className="font-mono text-xs">
              {refund.refundNo}
              {refund.channelRefundId ? ` · ${refund.channelRefundId}` : ""}
            </span>
          ) : null}
        </div>
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_order_detail_refund_reason()} span={2}>
        {refund.reason || "-"}
      </Descriptions.Item>
    </Descriptions>
  );
}

function unlockLineOf(o: AdminOrder | undefined, unlock: OrderUnlockView | undefined): string {
  if (!o) return "";
  if (String(o.order?.state) === "ORDER_STATE_CLOSED")
    return m.esports_order_detail_unlock_closed();
  if (!unlock) return m.esports_order_detail_unlock_pending();
  if (unlock.recordId) return m.esports_order_detail_unlock_trial({ id: unlock.recordId });
  return m.esports_order_detail_unlock_period({
    from: shortOf(dateOf(unlock.grantedAt)),
    to: shortOf(dateOf(unlock.expiresAt)),
    // protojson drops zero scalars; zero unlocks is a real state.
    n: unlock.unlockCount ?? 0,
  });
}

// The offline-exception executor: one full refund per paid order, gated to
// reachability here (paid and never refunded) — the backend arbitrates the
// same gate anyway. The reason is the audit paper trail, so an empty one
// keeps the confirm disabled.
function RefundAction({
  orderNo,
  amount,
  onDone,
}: Readonly<{ orderNo: string; amount: string; onDone: () => void }>) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("");
  const refundMut = useMutation(refundPaymentOrder, {
    onSuccess: (resp) => {
      void toast.success(
        m.esports_order_detail_refund_done({ no: resp.order?.order?.orderNo ?? "" }),
      );
      setOpen(false);
      setReason("");
      onDone();
    },
  });
  return (
    <>
      <Button
        danger
        icon={<RotateCcw size={14} />}
        onClick={() => setOpen(true)}
        size="small"
        type="text"
      >
        {m.esports_order_detail_refund_act()}
      </Button>
      <Modal
        cancelText={m.esports_order_detail_refund_cancel()}
        okButtonProps={{
          danger: true,
          disabled: !reason.trim(),
          loading: refundMut.isPending,
        }}
        okText={m.esports_order_detail_refund_ok()}
        onCancel={() => setOpen(false)}
        onOk={() => refundMut.mutate({ orderNo, reason: reason.trim() })}
        open={open}
        title={m.esports_order_detail_refund_title()}
      >
        <p className="text-sm">{m.esports_order_detail_refund_desc({ amount })}</p>
        <div className="mt-3 mb-1 text-sm">{m.esports_order_detail_refund_label()}</div>
        <Input.TextArea
          onChange={(e) => setReason(e.target.value)}
          placeholder={m.esports_order_detail_refund_ph()}
          rows={3}
          value={reason}
        />
      </Modal>
    </>
  );
}

import { useMutation, useQuery } from "@connectrpc/connect-query";
import type {
  MembershipPeriod,
  MembershipSummary,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import {
  getMembership,
  grantMembership,
  terminateMembership,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import { Button, Card, Statistic, Tag, Typography } from "antd";
import { ArrowLeft, Ban, ExternalLink, RefreshCw } from "lucide-react";
import { useState } from "react";
import { bandLabel } from "#/features/modules/esports-users/bands";
import { m } from "#/paraglide/messages";
import { fullDate } from "./ui/dates";
import { GrantModal, MEMBER_MONTH_SKU } from "./ui/grant-modal";
import { isLive as isLiveRow, statusTags } from "./ui/members-table";
import { TerminateModal } from "./ui/terminate-modal";

const LIST_SEARCH = { pageNo: 1, pageSize: 20 };

export function EsportsMemberDetailPage() {
  const { id } = useParams({ from: "/dashboard/people/members/$id" });
  const navigate = useNavigate();
  const detailQuery = useQuery(getMembership, { userId: id });
  const [grantOpen, setGrantOpen] = useState(false);
  const [terminateRow, setTerminateRow] = useState<MembershipSummary | null>(null);

  const refetch = () => detailQuery.refetch();
  const grantMut = useMutation(grantMembership, {
    onSuccess: () => {
      setGrantOpen(false);
      void refetch();
    },
  });
  const terminateMut = useMutation(terminateMembership, {
    onSuccess: () => {
      setTerminateRow(null);
      void refetch();
    },
  });

  const d = detailQuery.data;
  const summary = d?.membership;
  const headline = [
    summary?.phoneMasked,
    d?.wechatBound ? m.esports_members_wechat_bound() : "",
    d?.ageBand ? bandLabel(d.ageBand) : "",
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="flex flex-col gap-4 p-4">
      <DetailHeader
        grantOpen={setGrantOpen}
        headline={headline}
        id={id}
        navigate={navigate}
        summary={summary}
        terminate={setTerminateRow}
      />

      <StatCards
        daysLeft={summary?.daysLeft}
        expiresAt={summary?.expiresAt}
        periodCount={summary?.periodCount}
        unlockCount={summary?.unlockCount}
      />

      <CurrentMembershipCard
        expiresAt={summary?.expiresAt}
        grantedAt={summary?.grantedAt}
        source={summary?.source}
      />

      <HistoryCard periods={d?.periods ?? []} />

      <GrantModal
        detail
        onCancel={() => setGrantOpen(false)}
        onSubmit={(userId) => grantMut.mutate({ userId, skuCode: MEMBER_MONTH_SKU })}
        open={grantOpen}
        pending={grantMut.isPending}
        row={summary ?? null}
      />
      <TerminateModal
        detail
        onCancel={() => setTerminateRow(null)}
        onSubmit={(userId) => terminateMut.mutate({ userId })}
        pending={terminateMut.isPending}
        row={terminateRow}
      />
    </div>
  );
}

interface HeaderProps {
  id: string;
  summary?: MembershipSummary;
  headline: string;
  grantOpen: (open: boolean) => void;
  terminate: (row: MembershipSummary) => void;
  navigate: ReturnType<typeof useNavigate>;
}

function DetailHeader({ id, summary, headline, grantOpen, terminate, navigate }: HeaderProps) {
  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Button
            icon={<ArrowLeft size={14} />}
            onClick={() => void navigate({ to: "/dashboard/people/members", search: LIST_SEARCH })}
            type="text"
          >
            {m.esports_members_detail_back()}
          </Button>
          <Typography.Title className="m-0" level={4}>
            {summary?.nickname || id}
          </Typography.Title>
          {summary ? <StatusTag status={String(summary.status ?? "")} /> : null}
        </div>
        <div className="flex items-center gap-2">
          <Link params={{ id }} to="/dashboard/people/users/$id">
            <Button icon={<ExternalLink size={14} />}>{m.esports_members_detail_nav()}</Button>
          </Link>
          {summary && isLiveRow(summary) ? (
            <>
              <Button icon={<RefreshCw size={14} />} onClick={() => grantOpen(true)} type="primary">
                {m.esports_members_act_renew()}
              </Button>
              <Button danger icon={<Ban size={14} />} onClick={() => terminate(summary)}>
                {m.esports_members_act_terminate()}
              </Button>
            </>
          ) : null}
        </div>
      </div>
      {headline ? <Typography.Text type="secondary">{headline}</Typography.Text> : null}
    </>
  );
}

function StatCards(props: {
  daysLeft?: number;
  expiresAt?: MembershipSummary["expiresAt"];
  periodCount?: number;
  unlockCount?: number;
}) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
      <Card>
        <Statistic
          suffix={props.daysLeft != null ? m.esports_members_card_days_unit() : undefined}
          title={m.esports_members_card_days()}
          value={props.daysLeft ?? "-"}
        />
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_members_card_days_sub({ date: fullDate(props.expiresAt) })}
        </Typography.Text>
      </Card>
      <Card>
        <Statistic title={m.esports_members_card_periods()} value={props.periodCount ?? "-"} />
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_members_card_periods_sub({ n: props.periodCount ?? 0 })}
        </Typography.Text>
      </Card>
      <Card>
        <Statistic title={m.esports_members_card_unlocks()} value={props.unlockCount ?? "-"} />
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_members_card_unlocks_sub()}
        </Typography.Text>
      </Card>
    </div>
  );
}

function CurrentMembershipCard(props: {
  grantedAt?: MembershipSummary["grantedAt"];
  expiresAt?: MembershipSummary["expiresAt"];
  source?: MembershipSummary["source"];
}) {
  return (
    <Card title={m.esports_members_current_title()}>
      <div className="flex flex-col gap-2">
        <Kv label={m.esports_members_current_sku()} value={m.esports_members_sku_label()} />
        <Kv
          label={m.esports_members_current_window()}
          value={`${fullDate(props.grantedAt)} → ${fullDate(props.expiresAt)}`}
        />
        <Kv
          label={m.esports_members_current_source()}
          value={sourceText(String(props.source ?? ""))}
        />
      </div>
    </Card>
  );
}

function HistoryCard({ periods }: { periods: MembershipPeriod[] }) {
  return (
    <Card
      extra={<Typography.Text type="secondary">{m.esports_members_history_note()}</Typography.Text>}
      title={m.esports_members_history_title()}
    >
      <div className="flex flex-col gap-3">
        {periods.map((p, i) => (
          <div
            className="flex flex-wrap items-baseline gap-2"
            // The timeline is append-only and read-only; same-day grants
            // would collide on date+source, so the row's ledger position is
            // the stable identity.
            // biome-ignore lint/suspicious/noArrayIndexKey: the ledger position is the stable identity here
            key={i}
          >
            <Typography.Text className="w-24 shrink-0">{fullDate(p.grantedAt)}</Typography.Text>
            <span>{m.esports_members_history_grant()}</span>
            <span>{m.esports_members_sku_label()}</span>
            <Typography.Text type="secondary">
              （{sourceText(String(p.source ?? ""))}）
            </Typography.Text>
            <StatusTag status={String(p.status ?? "")} />
            <span>→ {fullDate(p.expiresAt)}</span>
          </div>
        ))}
      </div>
    </Card>
  );
}

function Kv({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex gap-2">
      <Typography.Text className="w-20 shrink-0" type="secondary">
        {label}
      </Typography.Text>
      <span>{value}</span>
    </div>
  );
}

function StatusTag({ status }: { status: string }) {
  const tag = statusTags[status];
  return tag ? (
    <Tag className="m-0" color={tag.color}>
      {tag.label()}
    </Tag>
  ) : (
    <Tag className="m-0">{status}</Tag>
  );
}

function sourceText(source: string): string {
  if (source === "MEMBERSHIP_SOURCE_MANUAL") return m.esports_members_source_manual();
  if (source === "MEMBERSHIP_SOURCE_LEGACY_CODE") return m.esports_members_source_legacy();
  return m.esports_members_source_order();
}

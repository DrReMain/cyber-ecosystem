import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getMembership,
  grantMembership,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { Link } from "@tanstack/react-router";
import { Button, Tag, Typography } from "antd";
import { ExternalLink, RefreshCw } from "lucide-react";
import { useState } from "react";
import { view } from "#/domains/error";
import { fullDate } from "#/features/modules/esports-members/ui/dates";
import { GrantModal, MEMBER_MONTH_SKU } from "#/features/modules/esports-members/ui/grant-modal";
import {
  isLive,
  sourceText,
  statusTags,
} from "#/features/modules/esports-members/ui/members-table";
import { m } from "#/paraglide/messages";

// The membership view inside the user detail. A member without any
// membership row answers not-found — here that is the ordinary "never
// held" shape, not an error; the grant entry rides the page header.
export function MembershipTab({ userId }: Readonly<{ userId: string }>) {
  // Not-found is the ordinary never-held shape — handled below, not thrown
  // to the page boundary.
  const mshipQuery = useQuery(getMembership, { userId }, { throwOnError: false });
  const [renewOpen, setRenewOpen] = useState(false);
  const grantMut = useMutation(grantMembership, {
    onSuccess: () => {
      setRenewOpen(false);
      void mshipQuery.refetch();
    },
  });
  const summary = mshipQuery.data?.membership;

  if (mshipQuery.isError) {
    // The domain's own envelope answers the shape — a 404 is the ordinary
    // never-held answer, anything else is a real error line.
    if (view(mshipQuery.error).http === 404) {
      return (
        <div className="p-10 text-center">
          <Typography.Text type="secondary">{m.esports_user_detail_mem_none()}</Typography.Text>
        </div>
      );
    }
    return (
      <div className="p-10 text-center">
        <Typography.Text type="danger">{m.esports_user_detail_error()}</Typography.Text>
      </div>
    );
  }

  if (!summary) {
    return <div className="flex justify-center p-10" />;
  }

  const tag = statusTags[String(summary.status ?? "")];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          {tag ? (
            <Tag className="m-0" color={tag.color}>
              {tag.label()}
            </Tag>
          ) : (
            <Tag className="m-0">{String(summary.status ?? "")}</Tag>
          )}
          <span className="tabular-nums">
            {fullDate(summary.grantedAt)} → {fullDate(summary.expiresAt)}
          </span>
          {summary.daysLeft != null ? (
            <Typography.Text type="secondary">
              {m.esports_user_detail_mem_days_left({ n: summary.daysLeft })}
            </Typography.Text>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <Link params={{ id: userId }} to="/dashboard/people/members/$id">
            <Button icon={<ExternalLink size={14} />}>
              {m.esports_user_detail_mem_to_member()}
            </Button>
          </Link>
          {isLive(summary) ? (
            <Button
              icon={<RefreshCw size={14} />}
              onClick={() => setRenewOpen(true)}
              type="primary"
            >
              {m.esports_user_detail_act_renew()}
            </Button>
          ) : null}
        </div>
      </div>
      <div className="flex flex-col gap-1">
        <Kv
          label={m.esports_user_detail_mem_source()}
          value={sourceText(String(summary.source ?? ""))}
        />
        <Kv
          label={m.esports_user_detail_mem_periods()}
          value={m.esports_user_detail_mem_periods_value({ n: summary.periodCount ?? 0 })}
        />
        <Kv
          label={m.esports_user_detail_mem_unlocks()}
          value={m.esports_user_detail_mem_unlocks_value({ n: summary.unlockCount ?? 0 })}
        />
      </div>
      <div className="flex flex-col gap-2">
        <Typography.Text className="font-semibold" type="secondary">
          {m.esports_user_detail_mem_history()}
        </Typography.Text>
        {(mshipQuery.data?.periods ?? []).map((p, i) => (
          <div
            className="flex flex-wrap items-baseline gap-2"
            key={p.grantedAt ? `${p.grantedAt}-${i}` : i}
          >
            <Typography.Text className="tabular-nums">{fullDate(p.grantedAt)}</Typography.Text>
            <span>{sourceText(String(p.source ?? ""))}</span>
            <Typography.Text type="secondary">→ {fullDate(p.expiresAt)}</Typography.Text>
            {(() => {
              const t = statusTags[String(p.status ?? "")];
              return t ? (
                <Tag className="m-0" color={t.color}>
                  {t.label()}
                </Tag>
              ) : null;
            })()}
          </div>
        ))}
      </div>
      <Typography.Text className="text-xs" type="secondary">
        {m.esports_user_detail_mem_note()}
      </Typography.Text>
      <GrantModal
        detail
        onCancel={() => setRenewOpen(false)}
        onSubmit={(uid) => grantMut.mutate({ userId: uid, skuCode: MEMBER_MONTH_SKU })}
        open={renewOpen}
        pending={grantMut.isPending}
        row={summary}
      />
    </div>
  );
}

function Kv({ label, value }: Readonly<{ label: string; value: string }>) {
  return (
    <div className="flex gap-2">
      <Typography.Text className="w-24 shrink-0" type="secondary">
        {label}
      </Typography.Text>
      <span>{value}</span>
    </div>
  );
}

import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getMembership,
  grantMembership,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import type { Member } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { getMember } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member-MemberService_connectquery";
import { updateUserStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { useColumns } from "@cyber-ecosystem/shared-antd/table";
import { useNavigate, useParams } from "@tanstack/react-router";
import {
  Button,
  Card,
  Descriptions,
  Modal,
  Popconfirm,
  Space,
  Spin,
  Tabs,
  Tag,
  Typography,
} from "antd";
import { ArrowLeft, BadgeCheck, Power, PowerOff } from "lucide-react";
import { useState } from "react";
import { GrantModal, MEMBER_MONTH_SKU } from "#/features/modules/esports-members/ui/grant-modal";
import { parseUsersSearch } from "#/features/modules/esports-users";
import { bandLabel } from "#/features/modules/esports-users/bands";
import { UserAvatar } from "#/features/modules/system-users/ui/user-avatar";
import { m } from "#/paraglide/messages";
import { AssessmentsTab } from "./ui/assessments-tab";
import { MembershipTab } from "./ui/membership-tab";

const LIST_SEARCH = parseUsersSearch({});

export function EsportsUserDetailPage() {
  const { id: userId } = useParams({ from: "/dashboard/people/users/$id" });
  const navigate = useNavigate();
  const memberQuery = useQuery(getMember, { userId });
  const toggleMut = useMutation(updateUserStatus, { onSuccess: () => memberQuery.refetch() });
  const member = memberQuery.data?.member;
  // Shared with the membership tab and the unlock card — one request, one
  // cache entry; the grant entry in the header needs the live summary.
  // Not-found is the ordinary never-held shape, not a page error — the
  // session-watcher's own local opt-out.
  const mshipQuery = useQuery(getMembership, { userId }, { throwOnError: false });
  const [grantOpen, setGrantOpen] = useState(false);
  const grantMut = useMutation(grantMembership, {
    onSuccess: () => {
      setGrantOpen(false);
      void mshipQuery.refetch();
    },
  });
  const summary = mshipQuery.data?.membership;

  if (memberQuery.isPending) {
    return (
      <div className="flex justify-center p-16">
        <Spin />
      </div>
    );
  }

  const back = (
    <Button
      className="self-start"
      color="default"
      icon={<ArrowLeft size={14} />}
      onClick={() => void navigate({ to: "/dashboard/people/users", search: LIST_SEARCH })}
      variant="text"
    >
      {m.esports_user_detail_back()}
    </Button>
  );

  if (member === undefined) {
    return (
      <div className="flex flex-col gap-4 p-4">
        {back}
        <Card>
          <Typography.Text type="danger">{m.esports_user_detail_error()}</Typography.Text>
        </Card>
      </div>
    );
  }

  const notOpen = (key: string, label: string) => ({
    key,
    label,
    children: (
      <div className="flex justify-center p-10">
        <Typography.Text type="secondary">{m.esports_user_detail_not_open()}</Typography.Text>
      </div>
    ),
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      {back}
      <MemberHeader
        member={member}
        onGrant={() => setGrantOpen(true)}
        onToggle={(next) => toggleMut.mutate({ id: member.userId ?? "", enabled: next })}
        pending={toggleMut.isPending}
      />
      <Card>
        <Tabs
          defaultActiveKey="overview"
          items={[
            {
              key: "overview",
              label: m.esports_user_detail_tab_overview(),
              children: <OverviewTab member={member} />,
            },
            notOpen("trace", m.esports_user_detail_tab_trace()),
            {
              key: "assessments",
              label: m.esports_user_detail_tab_assessments(),
              children: <AssessmentsTab userId={member.userId ?? ""} />,
            },
            {
              key: "membership",
              label: m.esports_user_detail_tab_membership(),
              children: <MembershipTab userId={member.userId ?? ""} />,
            },
            notOpen("messages", m.esports_user_detail_tab_messages()),
            {
              key: "account",
              label: m.esports_user_detail_tab_account(),
              children: <AccountTab member={member} />,
            },
          ]}
        />
      </Card>
      {/* The grant entry picks by what there is to show: no summary yet —
          never-held, still loading, or a failed read — takes the plain
          confirm (granting needs no read premise); a summary on hand takes
          the renewal's double-confirm shape. */}
      {summary == null ? (
        <Modal
          cancelText={m.common_cancel()}
          okButtonProps={{ loading: grantMut.isPending }}
          okText={m.esports_user_detail_grant_ok()}
          onCancel={() => setGrantOpen(false)}
          onOk={() => grantMut.mutate({ userId: member.userId ?? "", skuCode: MEMBER_MONTH_SKU })}
          open={grantOpen}
          title={m.esports_user_detail_act_grant()}
        >
          <p>
            {m.esports_user_detail_grant_body({
              name: member.nickname ?? member.account ?? member.userId ?? "",
            })}
          </p>
        </Modal>
      ) : (
        <GrantModal
          detail
          onCancel={() => setGrantOpen(false)}
          onSubmit={(uid) => grantMut.mutate({ userId: uid, skuCode: MEMBER_MONTH_SKU })}
          open={grantOpen}
          pending={grantMut.isPending}
          row={summary}
        />
      )}
    </div>
  );
}

function StatusTag({ enabled }: Readonly<{ enabled: boolean }>) {
  return enabled ? (
    <Tag className="m-0" color="success">
      {m.esports_users_status_on()}
    </Tag>
  ) : (
    <Tag className="m-0" color="error">
      {m.esports_users_status_off()}
    </Tag>
  );
}

function MemberHeader({
  member,
  pending,
  onGrant,
  onToggle,
}: Readonly<{
  member: Member;
  pending: boolean;
  onGrant: () => void;
  onToggle: (next: boolean) => void;
}>) {
  return (
    <Card>
      <Space size="middle">
        <UserAvatar account={member.account ?? member.userId ?? ""} size={48} url={undefined} />
        <div className="flex flex-col gap-1">
          <Space size={8}>
            <span className="text-base font-semibold">
              {member.nickname ?? m.esports_users_nickname_unset()}
            </span>
            {member.ageBand !== undefined && <Tag className="m-0">{bandLabel(member.ageBand)}</Tag>}
            <StatusTag enabled={member.enabled ?? false} />
          </Space>
          <Typography.Text className="font-mono" type="secondary">
            {member.account}
          </Typography.Text>
        </div>
      </Space>
      <div className="mt-4 flex flex-wrap gap-2">
        <Button color="purple" icon={<BadgeCheck size={14} />} onClick={onGrant} variant="filled">
          {m.esports_user_detail_act_grant()}
        </Button>
        <Popconfirm
          cancelButtonProps={{ variant: "filled", color: "default" }}
          description={member.enabled ? m.esports_users_ban_confirm_description() : undefined}
          okButtonProps={{ variant: "filled", color: "primary" }}
          onConfirm={() => onToggle(!member.enabled)}
          title={
            member.enabled
              ? m.esports_users_ban_confirm_title()
              : m.esports_users_unban_confirm_title()
          }
        >
          <Button
            color={member.enabled ? "danger" : "primary"}
            disabled={pending}
            icon={member.enabled ? <PowerOff size={14} /> : <Power size={14} />}
            variant="filled"
          >
            {member.enabled ? m.esports_users_act_ban() : m.esports_users_act_unban()}
          </Button>
        </Popconfirm>
      </div>
    </Card>
  );
}

function OverviewTab({ member }: Readonly<{ member: Member }>) {
  const { formatTime } = useColumns();
  const birth =
    member.birthYear !== undefined && member.birthMonth !== undefined
      ? `${member.birthYear} · ${member.birthMonth}`
      : m.esports_user_detail_birth_unset();
  return (
    <Descriptions bordered column={2} size="small">
      <Descriptions.Item label={m.esports_user_detail_kv_account()} span={2}>
        <span className="font-mono">{member.account}</span>
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_nickname()}>
        {member.nickname ?? "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_age_band()}>
        {member.ageBand !== undefined ? bandLabel(member.ageBand) : "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_birth()}>
        {member.birthYear !== undefined ? (
          birth
        ) : (
          <Space size={4}>
            <span>{birth}</span>
            <Tag className="m-0" color="warning">
              {m.esports_user_detail_birth_uncollected()}
            </Tag>
          </Space>
        )}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_registered_at()}>
        {formatTime(member.registeredAt, { time: "second" }) || "-"}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_games()} span={2}>
        {(member.games ?? []).length > 0 ? (
          <Space size={[4, 8]} wrap>
            {(member.games ?? []).map((game, index) => (
              <Tag className="m-0" key={`${game.gameId ?? game.customName ?? index}`}>
                {game.customName ?? game.gameId}
              </Tag>
            ))}
          </Space>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        )}
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_user_id()} span={2}>
        <span className="font-mono">{member.userId}</span>
      </Descriptions.Item>
    </Descriptions>
  );
}

function AccountTab({ member }: Readonly<{ member: Member }>) {
  return (
    <Descriptions bordered column={2} size="small">
      <Descriptions.Item label={m.esports_user_detail_kv_status()}>
        <StatusTag enabled={member.enabled ?? false} />
      </Descriptions.Item>
      <Descriptions.Item label={m.esports_user_detail_kv_population()}>member</Descriptions.Item>
    </Descriptions>
  );
}

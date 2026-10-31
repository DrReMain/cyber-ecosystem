import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Member } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { getMember } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member-MemberService_connectquery";
import { updateUserStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { useColumns } from "@cyber-ecosystem/shared-antd/table";
import { useNavigate, useParams } from "@tanstack/react-router";
import { Button, Card, Descriptions, Popconfirm, Space, Spin, Tabs, Tag, Typography } from "antd";
import { ArrowLeft, Power, PowerOff } from "lucide-react";
import { parseUsersSearch } from "#/features/modules/esports-users";
import { bandLabel } from "#/features/modules/esports-users/bands";
import { UserAvatar } from "#/features/modules/system-users/ui/user-avatar";
import { m } from "#/paraglide/messages";

const LIST_SEARCH = parseUsersSearch({});

export function EsportsUserDetailPage() {
  const { id: userId } = useParams({ from: "/dashboard/people/users/$id" });
  const navigate = useNavigate();
  const memberQuery = useQuery(getMember, { userId });
  const toggleMut = useMutation(updateUserStatus, { onSuccess: () => memberQuery.refetch() });
  const member = memberQuery.data?.member;

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
            notOpen("assessments", m.esports_user_detail_tab_assessments()),
            notOpen("membership", m.esports_user_detail_tab_membership()),
            notOpen("messages", m.esports_user_detail_tab_messages()),
            {
              key: "account",
              label: m.esports_user_detail_tab_account(),
              children: <AccountTab member={member} />,
            },
          ]}
        />
      </Card>
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
  onToggle,
}: Readonly<{ member: Member; pending: boolean; onToggle: (next: boolean) => void }>) {
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
      <div className="mt-4">
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

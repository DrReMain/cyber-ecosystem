import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { useColumns } from "@cyber-ecosystem/shared-antd/table";
import { Button, Empty, Input, Popconfirm, Tag } from "antd";
import { Search } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

export interface MemberView {
  principalType: string;
  principalId: string;
  createdAt?: Timestamp | string;
  displayName: string;
}

interface RoleMembersProps {
  members: MemberView[];
  pendingId: string | null;
  onUnbind: (member: MemberView) => void;
}

export function RoleMembers({ members, pendingId, onUnbind }: Readonly<RoleMembersProps>) {
  const { formatTime } = useColumns();
  const [query, setQuery] = useState("");
  const needle = query.trim().toLowerCase();
  const visible = members.filter(
    (member) =>
      member.principalId.toLowerCase().includes(needle) ||
      member.displayName.toLowerCase().includes(needle),
  );

  return (
    <div className="flex flex-col gap-3">
      <Input
        allowClear
        onChange={(e) => setQuery(e.target.value)}
        placeholder={m.system_roles_member_search_placeholder()}
        prefix={<Search className="text-ink-quaternary" size={14} />}
        value={query}
      />
      {visible.length === 0 ? (
        <Empty
          description={
            members.length === 0
              ? m.system_roles_members_empty()
              : m.system_roles_members_no_match()
          }
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      ) : (
        <div className="flex flex-col gap-2">
          {visible.map((member) => (
            <div
              className="flex items-center gap-3 rounded-lg border border-line-soft px-3 py-2"
              key={`${member.principalType}:${member.principalId}`}
            >
              {member.principalType === "user" ? (
                <Tag>{m.system_roles_member_user()}</Tag>
              ) : (
                <Tag color="geekblue">{member.principalType}</Tag>
              )}
              <span className="text-[13px]">{member.displayName}</span>
              <span className="font-mono text-[12px] text-ink-tertiary">{member.principalId}</span>
              <span className="ms-auto font-mono text-[12px] text-ink-tertiary">
                {formatTime(member.createdAt)}
              </span>
              <Popconfirm
                cancelButtonProps={{ color: "default", variant: "filled" }}
                okButtonProps={{ color: "danger", variant: "filled" }}
                onConfirm={() => onUnbind(member)}
                placement={getTextDirection() === "rtl" ? "right" : "left"}
                title={m.system_roles_unbind_confirm()}
              >
                <Button
                  color="danger"
                  disabled={pendingId === member.principalId}
                  size="small"
                  variant="text"
                >
                  {m.system_roles_member_act_unbind()}
                </Button>
              </Popconfirm>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

import { Button, Empty, Input, Popconfirm, Tag } from "antd";
import { Search } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

export interface DeptMemberView {
  userId: string;
  email: string;
  enabled: boolean;
}

interface DeptMembersProps {
  deptId: string;
  members: DeptMemberView[];
  pendingId: string | null;
  onRemove: (member: DeptMemberView) => void;
}

export function DeptMembers({ members, pendingId, onRemove }: Readonly<DeptMembersProps>) {
  const [query, setQuery] = useState("");
  const needle = query.trim().toLowerCase();
  const visible = members.filter(
    (member) =>
      member.email.toLowerCase().includes(needle) || member.userId.toLowerCase().includes(needle),
  );

  return (
    <div className="flex flex-col gap-3">
      <Input
        allowClear
        onChange={(e) => setQuery(e.target.value)}
        placeholder={m.system_depts_member_search_placeholder()}
        prefix={<Search className="text-ink-quaternary" size={14} />}
        value={query}
      />
      {visible.length === 0 ? (
        <Empty
          description={
            members.length === 0
              ? m.system_depts_members_empty()
              : m.system_depts_members_no_match()
          }
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      ) : (
        <div className="flex flex-col gap-2">
          {visible.map((member) => (
            <div
              className="flex items-center justify-between gap-2 rounded-lg border border-line-soft px-3 py-2"
              key={member.userId}
            >
              <span className="text-[13px]">{member.email}</span>

              <div className="flex items-center gap-2">
                {!member.enabled && <Tag>{m.system_depts_member_disabled()}</Tag>}
                <Popconfirm
                  cancelButtonProps={{ color: "default", variant: "filled" }}
                  okButtonProps={{ color: "danger", variant: "filled" }}
                  onConfirm={() => onRemove(member)}
                  placement={getTextDirection() === "rtl" ? "right" : "left"}
                  title={m.system_depts_member_remove_confirm()}
                >
                  <Button
                    color="danger"
                    disabled={pendingId === member.userId}
                    size="small"
                    variant="text"
                  >
                    {m.system_depts_member_act_remove()}
                  </Button>
                </Popconfirm>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

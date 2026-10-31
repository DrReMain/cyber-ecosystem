import type { Timestamp } from "@bufbuild/protobuf/wkt";
import type { MembershipSummary } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { Input, Modal, Typography } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";
import { dateOf, shortDate, shortOf } from "./dates";

export const MEMBER_MONTH_SKU = "member_month";

function renewTarget(ts: Timestamp | null | undefined): Date | null {
  const d = dateOf(ts);
  return d ? new Date(d.getTime() + 31 * 86_400_000) : null;
}

interface GrantModalProps {
  open: boolean;
  // Null = the toolbar's blank grant (operator types the id); a row = the
  // renew entry, which already knows its member.
  row: MembershipSummary | null;
  // The detail page's renew variant names the extension target and carries
  // the double-confirm wording; the list row stays compact.
  detail?: boolean;
  pending: boolean;
  onCancel: () => void;
  onSubmit: (userId: string) => void;
}

export function GrantModal({
  open,
  row,
  detail = false,
  pending,
  onCancel,
  onSubmit,
}: Readonly<GrantModalProps>) {
  const [userId, setUserId] = useState("");
  useEffect(() => {
    if (open) setUserId(row?.userId ?? "");
  }, [open, row]);
  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ loading: pending, disabled: !row && userId.trim().length !== 20 }}
      okText={row ? m.esports_members_renew_ok() : m.esports_members_grant_ok()}
      onCancel={onCancel}
      onOk={() => onSubmit((row?.userId ?? userId).trim())}
      open={open}
      title={
        row
          ? m.esports_members_renew_title({ name: row.nickname || row.userId || "" })
          : m.esports_members_grant_title()
      }
    >
      {row ? (
        <p className="mb-2">
          {detail
            ? m.esports_members_renew_body_detail({
                from: shortOf(dateOf(row.expiresAt)),
                to: shortOf(renewTarget(row.expiresAt)),
              })
            : m.esports_members_renew_body({ date: shortDate(row.expiresAt) })}
        </p>
      ) : (
        <div className="mb-2">
          <p className="mb-2">{m.esports_members_grant_body()}</p>
          <Input onChange={(e) => setUserId(e.target.value)} placeholder="user_id" value={userId} />
        </div>
      )}
      <Typography.Text type="secondary">
        {row
          ? detail
            ? m.esports_members_renew_note_detail()
            : m.esports_members_renew_note()
          : m.esports_members_grant_note()}
      </Typography.Text>
    </Modal>
  );
}

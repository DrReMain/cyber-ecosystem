import type { MembershipSummary } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { Modal, Typography } from "antd";
import { m } from "#/paraglide/messages";

interface TerminateModalProps {
  row: MembershipSummary | null;
  // The detail page variant spells the report-unlock fate out fully; the
  // list row keeps the compact wording.
  detail?: boolean;
  pending: boolean;
  onCancel: () => void;
  onSubmit: (userId: string) => void;
}

export function TerminateModal({
  row,
  detail = false,
  pending,
  onCancel,
  onSubmit,
}: Readonly<TerminateModalProps>) {
  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ danger: true, loading: pending, disabled: !row }}
      okText={m.esports_members_terminate_ok()}
      onCancel={onCancel}
      onOk={() => row && onSubmit(row.userId ?? "")}
      open={!!row}
      title={m.esports_members_terminate_title({ name: row?.nickname || row?.userId || "" })}
    >
      <p className="mb-2">
        {detail ? m.esports_members_terminate_body_detail() : m.esports_members_terminate_body()}
      </p>
      <Typography.Text type="secondary">{m.esports_members_terminate_note()}</Typography.Text>
    </Modal>
  );
}

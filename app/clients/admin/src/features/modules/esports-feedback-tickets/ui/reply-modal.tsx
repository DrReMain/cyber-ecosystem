import type { FeedbackTicket } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import { Form, Input, Modal, Typography } from "antd";
import { m } from "#/paraglide/messages";

interface ReplyModalProps {
  onClose: () => void;
  onSubmit: (content: string) => void;
  open: boolean;
  submitting?: boolean;
  ticket: FeedbackTicket | null;
}

export function ReplyModal({
  onClose,
  onSubmit,
  open,
  submitting,
  ticket,
}: Readonly<ReplyModalProps>) {
  const [form] = Form.useForm<{ content?: string }>();

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ loading: submitting }}
      okText={m.esports_feedback_tickets_reply_submit()}
      onCancel={() => {
        form.resetFields();
        onClose();
      }}
      onOk={() => {
        void form.validateFields().then((v) => onSubmit((v.content ?? "").trim()));
      }}
      open={open && ticket !== null}
      title={m.esports_feedback_tickets_reply_title()}
    >
      {/* The ask being answered, so the reply is never written blind. */}
      <Typography.Paragraph className="text-xs" type="secondary">
        {m.esports_feedback_tickets_reply_question({
          user: ticket?.nickname ?? ticket?.userId ?? "-",
          content: ticket?.content ?? "",
        })}
      </Typography.Paragraph>
      <Form form={form} layout="vertical">
        {/* The wire ceiling is 2000; the counter makes the budget visible
            the way the prototype's open textarea did not. */}
        <Form.Item name="content" rules={[{ required: true }, { max: 2000 }, { whitespace: true }]}>
          <Input.TextArea
            autoSize={{ maxRows: 10, minRows: 4 }}
            maxLength={2000}
            placeholder={m.esports_feedback_tickets_reply_placeholder()}
            showCount
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

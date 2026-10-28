import { create } from "@bufbuild/protobuf";
import type { GenerateActivationCodesResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { GenerateActivationCodesRequestSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import type { ReportType } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/entitlement_pb";
import { Alert, Form, InputNumber, Modal, Select, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { reportTypeFromWire } from "../search";

interface GenerateModalProps {
  open: boolean;
  result: GenerateActivationCodesResponse | null;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (req: ReturnType<typeof buildRequest>) => void;
}

// Generation mints each code for a single redeem — an operator who needs N
// hands-out mints N; the count here is codes, never redeems per code.
const buildRequest = (reportType: ReportType | undefined, count?: number, expireDays?: number) =>
  create(GenerateActivationCodesRequestSchema, { reportType, count, expireDays });
export function GenerateModal({
  open,
  result,
  submitting,
  onClose,
  onSubmit,
}: Readonly<GenerateModalProps>) {
  const [form] = Form.useForm<{
    reportType?: string;
    count?: number;
    expireDays?: number;
  }>();

  return (
    <Modal
      cancelText={m.common_cancel()}
      destroyOnHidden
      okButtonProps={result ? undefined : { loading: submitting }}
      okText={result ? m.common_confirm() : m.esports_activation_codes_generate_submit()}
      onCancel={onClose}
      onOk={() => {
        if (result) {
          onClose();
          return;
        }
        void form.validateFields().then((v) => {
          onSubmit(buildRequest(reportTypeFromWire(v.reportType), v.count, v.expireDays));
        });
      }}
      open={open}
      title={m.esports_activation_codes_generate()}
    >
      {result ? (
        <div className="flex flex-col gap-3">
          <Alert
            message={m.esports_activation_codes_generate_success({ n: result.codes.length })}
            type="success"
          />
          <div className="flex max-h-72 flex-col gap-1 overflow-auto">
            {result.codes.map((code) => (
              <Typography.Text className="font-mono" copyable key={code}>
                {code}
              </Typography.Text>
            ))}
          </div>
          <Typography.Text type="secondary">
            {m.esports_activation_codes_generate_hint()}
          </Typography.Text>
        </div>
      ) : (
        <Form form={form} layout="vertical">
          <Form.Item
            label={m.esports_activation_codes_generate_type()}
            name="reportType"
            rules={[{ required: true }]}
          >
            <Select
              options={[
                { label: m.esports_report_type_cognitive(), value: "REPORT_TYPE_COGNITIVE" },
                { label: m.esports_report_type_career(), value: "REPORT_TYPE_CAREER" },
                { label: m.esports_report_type_school(), value: "REPORT_TYPE_SCHOOL" },
              ]}
            />
          </Form.Item>
          <Form.Item
            initialValue={1}
            label={m.esports_activation_codes_generate_count()}
            name="count"
            rules={[{ required: true }]}
          >
            <InputNumber className="w-full" max={1000} min={1} precision={0} />
          </Form.Item>
          <Form.Item
            initialValue={30}
            label={m.esports_activation_codes_generate_days()}
            name="expireDays"
            rules={[{ required: true }]}
          >
            <InputNumber className="w-full" max={3650} min={1} precision={0} />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}

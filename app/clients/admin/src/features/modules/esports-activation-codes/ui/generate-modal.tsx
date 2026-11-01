import { create } from "@bufbuild/protobuf";
import type { GenerateActivationCodesResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { GenerateActivationCodesRequestSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import type { ReportType } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/entitlement_pb";
import { Link } from "@tanstack/react-router";
import { Alert, Form, InputNumber, Modal, Select, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { reportTypeFromWire } from "../search";
import { reportTypeOptions } from "../vocab";

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
          <Link
            className="text-primary font-mono font-semibold"
            params={{ batchNo: result.batchNo }}
            search={{ pageNo: 1, pageSize: 50 }}
            to="/dashboard/channels/activation-codes/$batchNo"
          >
            {m.esports_activation_codes_generate_batch({ no: result.batchNo })}
          </Link>
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
              options={reportTypeOptions.map((o) => ({ label: o.label(), value: o.value }))}
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

import { create } from "@bufbuild/protobuf";
import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { CreateAgreementVersionRequestSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import { DatePicker, Form, Input, Modal, Select } from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { m } from "#/paraglide/messages";
import { docFromWire, docOptions } from "../vocab";

interface NewVersionModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (req: ReturnType<typeof buildRequest>) => void;
  submitting: boolean;
}

// Publishing is atomic by contract: the created version becomes current
// the moment it saves — there is no draft state to offer.
const buildRequest = (
  doc: ReturnType<typeof docFromWire>,
  no: string,
  effective: Dayjs | undefined,
  body: string,
) =>
  create(CreateAgreementVersionRequestSchema, {
    doc,
    no,
    effectiveAt: effective ? timestampFromMs(effective.valueOf()) : undefined,
    body,
  });

export function NewVersionModal({
  open,
  onClose,
  onSubmit,
  submitting,
}: Readonly<NewVersionModalProps>) {
  const [form] = Form.useForm<{ doc?: string; no?: string; effective?: Dayjs; body?: string }>();

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ loading: submitting }}
      okText={m.esports_agreement_versions_new_submit()}
      onCancel={onClose}
      onOk={() => {
        void form.validateFields().then((v) => {
          onSubmit(buildRequest(docFromWire(v.doc), v.no ?? "", v.effective, v.body ?? ""));
        });
      }}
      open={open}
      title={m.esports_agreement_versions_new()}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label={m.esports_agreement_versions_new_doc()}
          name="doc"
          rules={[{ required: true }]}
        >
          <Select options={docOptions.map((o) => ({ label: o.label(), value: o.value }))} />
        </Form.Item>
        <Form.Item
          label={m.esports_agreement_versions_new_no()}
          name="no"
          rules={[{ required: true }]}
        >
          <Input maxLength={20} placeholder={m.esports_agreement_versions_new_no_placeholder()} />
        </Form.Item>
        <Form.Item
          initialValue={dayjs()}
          label={m.esports_agreement_versions_new_effective()}
          name="effective"
          rules={[{ required: true }]}
        >
          <DatePicker className="w-full" />
        </Form.Item>
        <Form.Item
          label={m.esports_agreement_versions_new_body()}
          name="body"
          rules={[{ required: true }]}
        >
          <Input.TextArea
            placeholder={m.esports_agreement_versions_new_body_placeholder()}
            rows={10}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

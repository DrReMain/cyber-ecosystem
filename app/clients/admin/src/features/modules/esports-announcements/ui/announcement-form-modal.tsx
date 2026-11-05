import { Form, Input, Modal, Select, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { kindLabel, kindOptions, scopeLabel, scopeOptions } from "../vocab";

export interface AnnouncementDraftSeed {
  content: string;
  /**
   * The birth pair rides two carriers: an edit seed carries the
   * transport wire-name strings (the frozen pair, rendered through
   * the string-keyed vocab tables), while a create submission
   * carries the form's generated numeric constants. Consumers
   * narrow with typeof before any numeric send.
   */
  id?: string;
  kind?: number | string;
  scope?: number | string;
  title: string;
}

interface AnnouncementFormModalProps {
  seed: AnnouncementDraftSeed | null;
  onClose: () => void;
  onSubmit: (seed: AnnouncementDraftSeed) => void;
  submitting?: boolean;
}

// One form, two entries: create carries kind and scope (the contract
// offers no move for them later); edit reopens title and content only,
// with the frozen pair stated as a fact line. Hidden-on-close remounts
// the form so the seed's prefill applies afresh every open.
export function AnnouncementFormModal({
  seed,
  onClose,
  onSubmit,
  submitting,
}: Readonly<AnnouncementFormModalProps>) {
  const [form] = Form.useForm<{
    content?: string;
    kind?: number;
    scope?: number;
    title?: string;
  }>();
  const editing = seed?.kind !== undefined;

  return (
    <Modal
      cancelText={m.common_cancel()}
      destroyOnHidden
      okButtonProps={{ loading: submitting }}
      okText={
        editing
          ? m.esports_announcements_form_submit_edit()
          : m.esports_announcements_form_submit_new()
      }
      onCancel={onClose}
      onOk={() => {
        void form.validateFields().then((v) => {
          onSubmit({
            content: (v.content ?? "").trim(),
            // Edit mode never registers the kind/scope items, so the
            // birth pair and the row id ride the seed; create mode
            // reads kind/scope from the validated form.
            id: editing ? seed?.id : undefined,
            kind: editing ? seed?.kind : v.kind,
            scope: editing ? seed?.scope : v.scope,
            title: (v.title ?? "").trim(),
          });
        });
      }}
      open={seed !== null}
      title={
        editing
          ? m.esports_announcements_form_title_edit()
          : m.esports_announcements_form_title_new()
      }
    >
      <Form
        form={form}
        initialValues={{ content: seed?.content, title: seed?.title }}
        layout="vertical"
      >
        <Form.Item
          label={m.esports_announcements_form_field_title()}
          name="title"
          rules={[{ max: 64, required: true }]}
        >
          <Input maxLength={64} placeholder={m.esports_announcements_form_ph_title()} showCount />
        </Form.Item>

        {editing ? (
          // The birth pair, stated as a fact — the contract has no
          // update arm for kind or scope, so the edit face cannot offer
          // one. The wire carries enum names as strings while the options
          // carry the generated enum constants — the string-keyed label
          // tables bridge the two, an options.find on === never matches.
          <Typography.Paragraph className="text-xs" type="secondary">
            {m.esports_announcements_form_meta_edit({
              kind: kindLabel(String(seed?.kind ?? "")),
              scope: scopeLabel(String(seed?.scope ?? "")),
            })}
          </Typography.Paragraph>
        ) : (
          <>
            <Form.Item
              label={m.esports_announcements_form_field_kind()}
              name="kind"
              rules={[{ required: true }]}
            >
              <Select options={kindOptions.map((o) => ({ label: o.label(), value: o.value }))} />
            </Form.Item>
            <Form.Item
              label={m.esports_announcements_form_field_scope()}
              name="scope"
              rules={[{ required: true }]}
            >
              <Select options={scopeOptions.map((o) => ({ label: o.label(), value: o.value }))} />
            </Form.Item>
            {/* Outside the form item on purpose: a named Form.Item takes
                exactly one control child — a second child breaks the
                value injection and the field silently never binds. */}
            <Typography.Text className="-mt-2 block text-xs" type="secondary">
              {m.esports_announcements_form_scope_note()}
            </Typography.Text>
          </>
        )}

        <Form.Item
          label={m.esports_announcements_form_field_content()}
          name="content"
          rules={[{ max: 5000, required: true }]}
        >
          <Input.TextArea
            autoSize={{ maxRows: 12, minRows: 6 }}
            maxLength={5000}
            placeholder={m.esports_announcements_form_ph_content()}
            showCount
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

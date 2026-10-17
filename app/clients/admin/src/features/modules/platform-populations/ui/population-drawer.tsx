import { Button, Drawer, Form, Input, Select, Space } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

// Same charset law as the tenant join key; the server enforces it as the
// backstop, the form rule is the first gate.
const CODE_PATTERN = /^[a-z][a-z0-9-]{0,18}[a-z0-9]$/;

export interface PopulationDraft {
  code: string;
  name: string;
  archetype: string;
}

const ARCHETYPES = ["engine", "patron"] as const;

interface PopulationDrawerProps {
  open: boolean;
  pending: boolean;
  onClose: () => void;
  onSave: (draft: PopulationDraft) => void;
}

export function PopulationDrawer({
  open,
  pending,
  onClose,
  onSave,
}: Readonly<PopulationDrawerProps>) {
  const [form] = Form.useForm<PopulationDraft>();

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave(values);
    } catch {
      // invalid fields: the per-field feedback is the terminal state here
    }
  };

  return (
    <Drawer
      footer={
        <Space className="w-full justify-end">
          <Button color="default" onClick={onClose} variant="filled">
            {m.common_cancel()}
          </Button>
          <Button
            color="primary"
            loading={pending}
            onClick={() => void handleSave()}
            variant="filled"
          >
            {m.common_save()}
          </Button>
        </Space>
      }
      onClose={onClose}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      size={480}
      title={m.platform_populations_drawer_create_title()}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label={
            <>
              {m.platform_populations_form_code()}
              <span className="ms-1.5 font-normal text-[11px] text-ink-tertiary">
                {m.platform_populations_form_code_hint()}
              </span>
            </>
          }
          name="code"
          rules={[
            { required: true, message: m.platform_populations_form_code_required() },
            { pattern: CODE_PATTERN, message: m.platform_populations_form_code_pattern() },
          ]}
        >
          <Input maxLength={20} placeholder={m.platform_populations_ph_code()} />
        </Form.Item>
        <Form.Item
          label={m.platform_populations_form_name()}
          name="name"
          rules={[{ required: true, message: m.platform_populations_form_name_required() }]}
        >
          <Input maxLength={128} placeholder={m.platform_populations_ph_name()} />
        </Form.Item>
        <Form.Item
          label={m.platform_populations_form_archetype()}
          name="archetype"
          rules={[{ required: true, message: m.platform_populations_form_archetype_required() }]}
        >
          <Select
            options={ARCHETYPES.map((value) => ({
              value,
              label:
                value === "engine"
                  ? m.platform_populations_arch_engine()
                  : m.platform_populations_arch_patron(),
            }))}
            placeholder={m.platform_populations_ph_archetype()}
          />
        </Form.Item>
      </Form>
    </Drawer>
  );
}

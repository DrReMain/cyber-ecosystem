import { useMutation } from "@connectrpc/connect-query";
import { createTenant } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant-TenantService_connectquery";
import { Button, Drawer, Form, Input } from "antd";
import { m } from "#/paraglide/messages";

interface CreateTenantDrawerProps {
  onClose: () => void;
  onCreated: () => void;
  open: boolean;
}

// The runtime tenant-opening form: code and name only — administrators
// join through the tenant's own add-admin face, so opening carries no
// credentials and no tenant can lock itself out. A taken code answers
// through the shared error toast (the mutation cache renders it).
export function CreateTenantDrawer({
  open,
  onClose,
  onCreated,
}: Readonly<CreateTenantDrawerProps>) {
  const [form] = Form.useForm();
  const createMut = useMutation(createTenant, {
    onSuccess: () => {
      form.resetFields();
      onCreated();
      onClose();
    },
  });

  const handleSave = () => {
    void form.validateFields().then((values) => {
      createMut.mutate({
        code: values.code as string,
        name: values.name as string,
      });
    });
  };

  return (
    <Drawer
      footer={
        <div className="flex justify-end gap-2">
          <Button onClick={onClose}>{m.common_cancel()}</Button>
          <Button loading={createMut.isPending} onClick={handleSave} type="primary">
            {m.platform_tenants_create_submit()}
          </Button>
        </div>
      }
      onClose={onClose}
      open={open}
      title={m.platform_tenants_create_title()}
      width={480}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          extra={m.platform_tenants_create_code_hint()}
          label={m.platform_tenants_create_code()}
          name="code"
          rules={[
            { message: m.platform_tenants_create_code(), required: true },
            {
              message: m.platform_tenants_create_code_hint(),
              pattern: /^[a-z0-9][a-z0-9-]*$/,
            },
          ]}
        >
          <Input maxLength={64} placeholder="acme" />
        </Form.Item>
        <Form.Item
          label={m.platform_tenants_create_name()}
          name="name"
          rules={[{ message: m.platform_tenants_create_name(), required: true }]}
        >
          <Input maxLength={128} placeholder="ACME Corporation" />
        </Form.Item>
      </Form>
    </Drawer>
  );
}

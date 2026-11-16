import { useMutation } from "@connectrpc/connect-query";
import type { Tenant } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant_pb";
import { addTenantAdmin } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant-TenantService_connectquery";
import { Button, Drawer, Form, Input, Typography } from "antd";
import { m } from "#/paraglide/messages";

interface AddAdminDrawerProps {
  onClose: () => void;
  onAdded: () => void;
  open: boolean;
  tenant: Tenant | null;
}

// The runtime tenant's break-glass face: an existing account joins as
// super administrator (seated and bound, password untouched); an account
// that does not exist yet is created with the offered password. This is
// how a tenant whose administrators all lost their passwords gets back in
// — the platform admin adds a fresh one from here, the same role the boot
// seed plays for the default tenant.
export function AddAdminDrawer({ open, onClose, onAdded, tenant }: Readonly<AddAdminDrawerProps>) {
  const [form] = Form.useForm();
  const addMut = useMutation(addTenantAdmin, {
    onSuccess: () => {
      form.resetFields();
      onAdded();
      onClose();
    },
  });

  const handleSave = () => {
    if (tenant === null) return;
    void form.validateFields().then((values) => {
      addMut.mutate({
        tenantCode: tenant.code ?? "",
        account: values.account as string,
        password: values.password as string,
      });
    });
  };

  return (
    <Drawer
      footer={
        <div className="flex justify-end gap-2">
          <Button onClick={onClose}>{m.common_cancel()}</Button>
          <Button loading={addMut.isPending} onClick={handleSave} type="primary">
            {m.platform_tenants_add_admin_submit()}
          </Button>
        </div>
      }
      onClose={onClose}
      open={open}
      title={m.platform_tenants_add_admin_title()}
      width={480}
    >
      {tenant !== null && (
        <Typography.Paragraph className="font-mono" type="secondary">
          {tenant.name} · {tenant.code}
        </Typography.Paragraph>
      )}
      <Form form={form} layout="vertical">
        <Form.Item
          label={m.platform_tenants_add_admin_account()}
          name="account"
          rules={[{ message: m.platform_tenants_add_admin_account(), required: true }]}
        >
          <Input maxLength={64} placeholder="globex-root" />
        </Form.Item>
        <Form.Item
          label={m.platform_tenants_add_admin_password()}
          name="password"
          rules={[{ message: m.platform_tenants_add_admin_password(), required: true }]}
        >
          <Input.Password maxLength={64} placeholder="••••••••" />
        </Form.Item>
      </Form>
    </Drawer>
  );
}

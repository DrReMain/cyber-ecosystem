import { useMutation } from "@connectrpc/connect-query";
import { changePassword } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { Button, Card, Form, Input } from "antd";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

interface PasswordForm {
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}

export function PasswordPage() {
  const [form] = Form.useForm<PasswordForm>();
  const changeMut = useMutation(changePassword, {
    onSuccess: () => {
      toast.success(m.profile_password_changed());
      form.resetFields();
    },
  });

  return (
    <Card title={m.profile_password_title()}>
      <Form
        className="max-w-md"
        form={form}
        layout="vertical"
        onFinish={(values) =>
          changeMut.mutate({ oldPassword: values.oldPassword, newPassword: values.newPassword })
        }
      >
        <Form.Item
          label={m.profile_password_old()}
          name="oldPassword"
          rules={[{ required: true, message: m.profile_password_old_required() }]}
        >
          <Input.Password maxLength={128} placeholder="••••••••" />
        </Form.Item>
        <Form.Item
          dependencies={["oldPassword"]}
          label={m.profile_password_new()}
          name="newPassword"
          rules={[
            { required: true, message: m.profile_password_new_required() },
            {
              validator: (_, value: string) =>
                value && value === form.getFieldValue("oldPassword")
                  ? Promise.reject(new Error(m.profile_password_same_as_old()))
                  : Promise.resolve(),
            },
          ]}
        >
          <Input.Password maxLength={128} placeholder="••••••••" />
        </Form.Item>
        <Form.Item
          dependencies={["newPassword"]}
          label={m.profile_password_confirm()}
          name="confirmPassword"
          rules={[
            { required: true, message: m.profile_password_confirm_required() },
            ({ getFieldValue }) => ({
              validator: (_, value: string) =>
                !value || value === getFieldValue("newPassword")
                  ? Promise.resolve()
                  : Promise.reject(new Error(m.profile_password_confirm_mismatch())),
            }),
          ]}
        >
          <Input.Password maxLength={128} placeholder="••••••••" />
        </Form.Item>
        <Button color="primary" htmlType="submit" loading={changeMut.isPending} variant="filled">
          {m.profile_password_submit()}
        </Button>
      </Form>
    </Card>
  );
}

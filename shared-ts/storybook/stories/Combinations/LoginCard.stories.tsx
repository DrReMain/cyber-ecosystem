import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, Card, Checkbox, Form, Input, theme } from "antd";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Combinations/Login Card",
};

export default meta;

type Story = StoryObj;

function LoginCardPage() {
  const { token } = theme.useToken();
  return (
    <div style={{ maxWidth: 480 }}>
      <Section
        description="The skin's first-impression surface: brand accent, filled inputs, primary action."
        title="Sign-in"
      >
        <div
          style={{
            background: token.colorBgLayout,
            borderRadius: token.borderRadiusLG,
            padding: 40,
          }}
        >
          <Card style={{ maxWidth: 380, margin: "0 auto" }} title="CYBER ECOSYSTEM">
            <Form layout="vertical" onFinish={() => {}}>
              <Form.Item label="Email" name="email" rules={[{ required: true }]}>
                <Input autoComplete="off" placeholder="operator@grid.cn" />
              </Form.Item>
              <Form.Item label="Password" name="password" rules={[{ required: true }]}>
                <Input.Password placeholder="••••••••" />
              </Form.Item>
              <Form.Item name="remember" noStyle valuePropName="checked">
                <Checkbox>Keep me signed in</Checkbox>
              </Form.Item>
              <Button block htmlType="submit" style={{ marginTop: 16 }} type="primary">
                Sign in
              </Button>
            </Form>
          </Card>
        </div>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <LoginCardPage />,
};

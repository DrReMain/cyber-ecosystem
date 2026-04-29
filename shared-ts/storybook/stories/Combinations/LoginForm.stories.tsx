import { LockOutlined, MailOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Alert,
  Button,
  Checkbox,
  Divider,
  Flex,
  Form,
  Input,
  Typography,
} from "antd"
import { expect, fn, userEvent, within } from "storybook/test"

const meta: Meta<typeof Form> = {
  title: "Antd/Combinations/LoginForm",
  component: Form,
  parameters: { layout: "centered" },
}

export default meta
type Story = StoryObj<typeof Form>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div style={{ width: 400, padding: 32 }}>
      <Typography.Title level={3} style={{ marginBottom: 4 }}>
        Sign In
      </Typography.Title>
      <Typography.Text
        type="secondary"
        style={{ display: "block", marginBottom: 24 }}
      >
        Welcome back! Enter your credentials to continue.
      </Typography.Text>

      <Alert
        type="error"
        title="Invalid email or password"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Form layout="vertical" initialValues={{ remember: true }}>
        <Form.Item
          label="Email"
          name="email"
          rules={[{ required: true, type: "email" }]}
        >
          <Input prefix={<MailOutlined />} placeholder="you@example.com" />
        </Form.Item>
        <Form.Item
          label="Password"
          name="password"
          rules={[{ required: true, min: 8 }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter password"
          />
        </Form.Item>
        <Flex justify="space-between" align="center">
          <Form.Item name="remember" valuePropName="checked" noStyle>
            <Checkbox>Remember me</Checkbox>
          </Form.Item>
          <Typography.Link>Forgot password?</Typography.Link>
        </Flex>
        <Form.Item style={{ marginTop: 16, marginBottom: 8 }}>
          <Button type="primary" variant="solid" block>
            Sign In
          </Button>
        </Form.Item>
      </Form>

      <Divider>or</Divider>

      <Button variant="outlined" block>
        Continue with SSO
      </Button>

      <Typography.Text
        type="secondary"
        style={{ display: "block", marginTop: 16, textAlign: "center" }}
      >
        Don't have an account? <Typography.Link>Sign up</Typography.Link>
      </Typography.Text>
    </div>
  ),
}

export const Playground: Story = {
  render: () => (
    <div style={{ width: 400, padding: 32 }}>
      <Typography.Title level={3} style={{ marginBottom: 4 }}>
        Sign In
      </Typography.Title>
      <Typography.Text
        type="secondary"
        style={{ display: "block", marginBottom: 24 }}
      >
        Welcome back! Enter your credentials to continue.
      </Typography.Text>
      <Form layout="vertical" initialValues={{ remember: true }}>
        <Form.Item
          label="Email"
          name="email"
          rules={[{ required: true, type: "email" }]}
        >
          <Input prefix={<MailOutlined />} placeholder="you@example.com" />
        </Form.Item>
        <Form.Item
          label="Password"
          name="password"
          rules={[{ required: true, min: 8 }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter password"
          />
        </Form.Item>
        <Flex justify="space-between" align="center">
          <Form.Item name="remember" valuePropName="checked" noStyle>
            <Checkbox>Remember me</Checkbox>
          </Form.Item>
          <Typography.Link>Forgot password?</Typography.Link>
        </Flex>
        <Form.Item style={{ marginTop: 16, marginBottom: 8 }}>
          <Button type="primary" variant="solid" block>
            Sign In
          </Button>
        </Form.Item>
      </Form>
      <Divider>or</Divider>
      <Button variant="outlined" block>
        Continue with SSO
      </Button>
    </div>
  ),
}

export const ValidationErrors: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div style={{ width: 400, padding: 32 }}>
      <Typography.Title level={3} style={{ marginBottom: 4 }}>
        Sign In
      </Typography.Title>
      <Typography.Text
        type="secondary"
        style={{ display: "block", marginBottom: 24 }}
      >
        Welcome back! Enter your credentials to continue.
      </Typography.Text>

      <Alert
        type="error"
        title="Invalid email or password"
        description="Please check your credentials and try again."
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Form layout="vertical">
        <Form.Item
          label="Email"
          validateStatus="error"
          help="Please enter a valid email address"
        >
          <Input prefix={<MailOutlined />} value="invalid-email" />
        </Form.Item>
        <Form.Item
          label="Password"
          validateStatus="error"
          help="Password must be at least 8 characters"
        >
          <Input.Password prefix={<LockOutlined />} value="short" />
        </Form.Item>
        <Flex justify="space-between" align="center">
          <Form.Item name="remember" valuePropName="checked" noStyle>
            <Checkbox>Remember me</Checkbox>
          </Form.Item>
          <Typography.Link>Forgot password?</Typography.Link>
        </Flex>
        <Form.Item style={{ marginTop: 16, marginBottom: 8 }}>
          <Button type="primary" variant="solid" block>
            Sign In
          </Button>
        </Form.Item>
      </Form>

      <Divider>or</Divider>

      <Button variant="outlined" block>
        Continue with SSO
      </Button>

      <Typography.Text
        type="secondary"
        style={{ display: "block", marginTop: 16, textAlign: "center" }}
      >
        Don't have an account? <Typography.Link>Sign up</Typography.Link>
      </Typography.Text>
    </div>
  ),
}

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    onFinish: fn(),
  },
  render: (args) => (
    <div style={{ width: 400, padding: 32 }}>
      <Typography.Title level={3} style={{ marginBottom: 4 }}>
        Sign In
      </Typography.Title>
      <Typography.Text
        type="secondary"
        style={{ display: "block", marginBottom: 24 }}
      >
        Welcome back! Enter your credentials to continue.
      </Typography.Text>
      <Form name="interactive-login" layout="vertical" onFinish={args.onFinish}>
        <Form.Item
          label="Email"
          name="email"
          rules={[{ required: true, type: "email" }]}
        >
          <Input prefix={<MailOutlined />} placeholder="you@example.com" />
        </Form.Item>
        <Form.Item
          label="Password"
          name="password"
          rules={[{ required: true, min: 8 }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter password"
          />
        </Form.Item>
        <Form.Item style={{ marginTop: 16, marginBottom: 8 }}>
          <Button type="primary" variant="solid" block htmlType="submit">
            Sign In
          </Button>
        </Form.Item>
      </Form>
    </div>
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const emailInput = canvas.getByPlaceholderText("you@example.com")
    await userEvent.type(emailInput, "test@example.com")
    const passwordInput = canvas.getByPlaceholderText("Enter password")
    await userEvent.type(passwordInput, "password123")
    const submitBtn = canvas.getByRole("button", { name: "Sign In" })
    await submitBtn.click()
    await expect(args.onFinish).toHaveBeenCalled()
  },
}

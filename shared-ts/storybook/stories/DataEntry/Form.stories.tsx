import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Button,
  Checkbox,
  Flex,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
} from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { callbackArgs, disabledArg, sizeArg, variantArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Form> = {
  title: "Antd/Data Entry/Form",
  component: Form,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    size: sizeArg(["small", "medium", "large"]),
    variant: variantArg(),
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof Form>

const layouts = ["horizontal", "vertical", "inline"] as const
const sizes = ["small", "medium", "large"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Layouts">
        <Flex vertical gap={24}>
          {layouts.map((layout) => (
            <Flex key={layout} vertical gap={4}>
              <Label>{layout}</Label>
              <Form
                name={`layout-${layout}`}
                layout={layout}
                style={{ maxWidth: 500 }}
                initialValues={{ name: "John", remember: true }}
              >
                <Form.Item label="Name" name="name">
                  <Input placeholder="Your name" />
                </Form.Item>
                <Form.Item
                  label="Remember"
                  name="remember"
                  valuePropName="checked"
                >
                  <Checkbox>Remember me</Checkbox>
                </Form.Item>
                <Form.Item>
                  <Button color="primary" variant="solid" htmlType="submit">
                    Submit
                  </Button>
                </Form.Item>
              </Form>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Sizes">
        <Flex vertical gap={16}>
          {sizes.map((size) => (
            <Flex key={size} vertical gap={4}>
              <Label>{size}</Label>
              <Form
                name={`size-${size}`}
                size={size}
                layout="horizontal"
                labelCol={{ span: 6 }}
                wrapperCol={{ span: 18 }}
                style={{ maxWidth: 500 }}
              >
                <Form.Item label="Input" name="input">
                  <Input placeholder="Text input" />
                </Form.Item>
                <Form.Item label="Select" name="select">
                  <Select
                    placeholder="Select option"
                    options={[
                      { value: "a", label: "Option A" },
                      { value: "b", label: "Option B" },
                    ]}
                  />
                </Form.Item>
                <Form.Item label="Number" name="number">
                  <InputNumber style={{ width: "100%" }} />
                </Form.Item>
              </Form>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={16}>
          {variants.map((variant) => (
            <Flex key={variant} vertical gap={4}>
              <Label>{variant}</Label>
              <Form
                name={`variant-${variant}`}
                variant={variant}
                layout="horizontal"
                labelCol={{ span: 6 }}
                wrapperCol={{ span: 18 }}
                style={{ maxWidth: 500 }}
              >
                <Form.Item label="Input" name="input">
                  <Input placeholder={`Variant: ${variant}`} />
                </Form.Item>
                <Form.Item label="Select" name="select">
                  <Select
                    placeholder={`Variant: ${variant}`}
                    options={[
                      { value: "a", label: "Option A" },
                      { value: "b", label: "Option B" },
                    ]}
                  />
                </Form.Item>
              </Form>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Form
          name="disabled-form"
          disabled
          layout="horizontal"
          labelCol={{ span: 6 }}
          wrapperCol={{ span: 18 }}
          style={{ maxWidth: 500 }}
          initialValues={{
            username: "johndoe",
            email: "john@example.com",
            role: "admin",
            active: true,
          }}
        >
          <Form.Item label="Username" name="username">
            <Input />
          </Form.Item>
          <Form.Item label="Email" name="email">
            <Input />
          </Form.Item>
          <Form.Item label="Role" name="role">
            <Select
              options={[
                { value: "admin", label: "Admin" },
                { value: "user", label: "User" },
                { value: "guest", label: "Guest" },
              ]}
            />
          </Form.Item>
          <Form.Item label="Active" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Section>

      <Section title="Validation with Feedback">
        <Form
          name="validation"
          style={{ maxWidth: 400 }}
          onFinish={(_values) => {}}
        >
          <Form.Item
            label="Username"
            name="username"
            rules={[
              { required: true, message: "Please input your username!" },
              { min: 3, message: "Min 3 characters" },
            ]}
            hasFeedback
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: "Please input your email!" },
              { type: "email", message: "Invalid email!" },
            ]}
            hasFeedback
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Password"
            name="password"
            rules={[
              { required: true, message: "Please input your password!" },
              { min: 8, message: "Min 8 characters" },
            ]}
            hasFeedback
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            label="Confirm"
            name="confirm"
            dependencies={["password"]}
            hasFeedback
            rules={[
              { required: true, message: "Please confirm password!" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("password") === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error("Passwords do not match!"))
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item>
            <Button color="primary" variant="solid" htmlType="submit">
              Register
            </Button>
          </Form.Item>
        </Form>
      </Section>

      <Section title="Validation States Matrix">
        <Flex gap={24} wrap>
          <Flex vertical gap={16} style={{ flex: 1, minWidth: 280 }}>
            <Label>Error</Label>
            <Form layout="vertical" style={{ maxWidth: 400 }}>
              <Form.Item
                label="Input"
                validateStatus="error"
                help="This field is required"
              >
                <Input placeholder="Error state" />
              </Form.Item>
              <Form.Item
                label="Select"
                validateStatus="error"
                help="Please select an option"
              >
                <Select
                  placeholder="Error state"
                  options={[
                    { value: "a", label: "Option A" },
                    { value: "b", label: "Option B" },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label="Number"
                validateStatus="error"
                help="Value must be greater than 0"
              >
                <InputNumber style={{ width: "100%" }} />
              </Form.Item>
              <Form.Item
                label="Checkbox"
                validateStatus="error"
                help="You must accept the terms"
              >
                <Checkbox>I agree to the terms</Checkbox>
              </Form.Item>
            </Form>
          </Flex>

          <Flex vertical gap={16} style={{ flex: 1, minWidth: 280 }}>
            <Label>Warning</Label>
            <Form layout="vertical" style={{ maxWidth: 400 }}>
              <Form.Item
                label="Input"
                validateStatus="warning"
                help="This username may already exist"
              >
                <Input placeholder="Warning state" />
              </Form.Item>
              <Form.Item
                label="Select"
                validateStatus="warning"
                help="This selection is uncommon"
              >
                <Select
                  placeholder="Warning state"
                  options={[
                    { value: "a", label: "Option A" },
                    { value: "b", label: "Option B" },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label="Number"
                validateStatus="warning"
                help="Value is close to the limit"
              >
                <InputNumber style={{ width: "100%" }} />
              </Form.Item>
              <Form.Item
                label="Checkbox"
                validateStatus="warning"
                help="Please review this option carefully"
              >
                <Checkbox>Enable experimental feature</Checkbox>
              </Form.Item>
            </Form>
          </Flex>

          <Flex vertical gap={16} style={{ flex: 1, minWidth: 280 }}>
            <Label>Success</Label>
            <Form layout="vertical" style={{ maxWidth: 400 }}>
              <Form.Item
                label="Input"
                validateStatus="success"
                help="Username is available"
              >
                <Input placeholder="Success state" />
              </Form.Item>
              <Form.Item
                label="Select"
                validateStatus="success"
                help="Valid selection"
              >
                <Select
                  placeholder="Success state"
                  options={[
                    { value: "a", label: "Option A" },
                    { value: "b", label: "Option B" },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label="Number"
                validateStatus="success"
                help="Value is within range"
              >
                <InputNumber style={{ width: "100%" }} />
              </Form.Item>
              <Form.Item
                label="Checkbox"
                validateStatus="success"
                help="Configuration saved"
              >
                <Checkbox checked>Notifications enabled</Checkbox>
              </Form.Item>
            </Form>
          </Flex>
        </Flex>
      </Section>

      <Section title="Required Mark Styles">
        <Flex gap={24} wrap>
          <Flex vertical gap={4} style={{ flex: 1, minWidth: 280 }}>
            <Label>optional (default)</Label>
            <Form layout="vertical">
              <Form.Item label="Optional Field" name="opt1">
                <Input placeholder="No required mark" />
              </Form.Item>
              <Form.Item label="Required Field" name="req1" required>
                <Input placeholder="Has required mark" />
              </Form.Item>
            </Form>
          </Flex>
          <Flex vertical gap={4} style={{ flex: 1, minWidth: 280 }}>
            <Label>Required Mark Hidden</Label>
            <Form layout="vertical" requiredMark={false}>
              <Form.Item label="Field" name="f1" required>
                <Input placeholder="No mark shown" />
              </Form.Item>
            </Form>
          </Flex>
          <Flex vertical gap={4} style={{ flex: 1, minWidth: 280 }}>
            <Label>Custom Required Mark</Label>
            <Form
              layout="vertical"
              requiredMark={(label, { required }) => (
                <>
                  {label}
                  {required && (
                    <span style={{ color: "var(--ant-color-error)" }}> *</span>
                  )}
                </>
              )}
            >
              <Form.Item label="Field" name="f2" required>
                <Input placeholder="Custom mark" />
              </Form.Item>
            </Form>
          </Flex>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Form {...args} style={{ maxWidth: 400 }}>
      <Form.Item label="Name" name="name">
        <Input placeholder="Your name" />
      </Form.Item>
      <Form.Item label="Email" name="email">
        <Input placeholder="Your email" />
      </Form.Item>
      <Form.Item>
        <Button color="primary" variant="solid" htmlType="submit">
          Submit
        </Button>
      </Form.Item>
    </Form>
  ),
}

export const Interactive: Story = {
  args: {
    onFinish: fn(),
  },
  render: (args: any) => (
    <Form
      name="interactive"
      style={{ maxWidth: 400 }}
      initialValues={{ username: "" }}
      onFinish={args.onFinish}
    >
      <Form.Item label="Username" name="username">
        <Input placeholder="Type your name" />
      </Form.Item>
      <Form.Item>
        <Button color="primary" variant="solid" htmlType="submit">
          Submit
        </Button>
      </Form.Item>
    </Form>
  ),
  play: async ({ canvasElement, args }: any) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("textbox")
    await input.focus()
    await userEvent.type(input, "test")
    const submitBtn = canvas.getByRole("button", { name: /submit/i })
    await submitBtn.click()
    await expect(args.onFinish).toHaveBeenCalled()
  },
}

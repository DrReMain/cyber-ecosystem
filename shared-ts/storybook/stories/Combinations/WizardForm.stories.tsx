import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Flex,
  Form,
  Input,
  Radio,
  Row,
  Select,
  Steps,
} from "antd"
import { useState } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Combinations/WizardForm",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

function WizardDemo() {
  const [current, setCurrent] = useState(0)
  const [form] = Form.useForm()

  const steps = [
    { title: "Account", description: "Basic info" },
    { title: "Profile", description: "Personal details" },
    { title: "Preferences", description: "Settings" },
  ]

  const next = () => setCurrent(current + 1)
  const prev = () => setCurrent(current - 1)

  return (
    <Card style={{ maxWidth: 700 }}>
      <Steps current={current} items={steps} style={{ marginBottom: 32 }} />
      <Form form={form} layout="vertical">
        {current === 0 && (
          <Flex vertical gap={0}>
            <Form.Item
              label="Username"
              name="username"
              rules={[{ required: true }]}
            >
              <Input placeholder="Choose a username" />
            </Form.Item>
            <Form.Item
              label="Email"
              name="email"
              rules={[{ required: true, type: "email" }]}
            >
              <Input placeholder="your@email.com" />
            </Form.Item>
            <Form.Item
              label="Password"
              name="password"
              rules={[{ required: true, min: 8 }]}
            >
              <Input.Password placeholder="At least 8 characters" />
            </Form.Item>
          </Flex>
        )}
        {current === 1 && (
          <Flex vertical gap={0}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="First Name" name="firstName">
                  <Input placeholder="First name" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Last Name" name="lastName">
                  <Input placeholder="Last name" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Date of Birth" name="dob">
              <DatePicker style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item label="Gender" name="gender">
              <Radio.Group>
                <Radio value="male">Male</Radio>
                <Radio value="female">Female</Radio>
                <Radio value="other">Other</Radio>
              </Radio.Group>
            </Form.Item>
          </Flex>
        )}
        {current === 2 && (
          <Flex vertical gap={0}>
            <Form.Item label="Role" name="role">
              <Select
                placeholder="Select your role"
                options={[
                  { value: "developer", label: "Developer" },
                  { value: "designer", label: "Designer" },
                  { value: "manager", label: "Manager" },
                  { value: "other", label: "Other" },
                ]}
              />
            </Form.Item>
            <Form.Item label="Bio" name="bio">
              <Input.TextArea rows={3} placeholder="Tell us about yourself" />
            </Form.Item>
            <Form.Item name="agree" valuePropName="checked">
              <Checkbox>I agree to the terms and conditions</Checkbox>
            </Form.Item>
          </Flex>
        )}
      </Form>
      <Flex justify="space-between" style={{ marginTop: 16 }}>
        <Button onClick={prev} disabled={current === 0}>
          Previous
        </Button>
        <Flex gap={8}>
          {current < steps.length - 1 && (
            <Button variant="solid" color="primary" onClick={next}>
              Next
            </Button>
          )}
          {current === steps.length - 1 && (
            <Button
              variant="solid"
              color="primary"
              onClick={() => setCurrent(0)}
            >
              Submit
            </Button>
          )}
        </Flex>
      </Flex>
    </Card>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="Multi-Step Wizard Form"
        description="Steps + Form with navigation between steps"
      >
        <WizardDemo />
      </Section>
    </Flex>
  ),
}

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    onFinish: fn(),
  },
  // biome-ignore lint/suspicious/noExplicitAny: Storybook story args need any
  render: (args: any) => (
    <Card style={{ maxWidth: 700 }}>
      <Steps
        current={0}
        items={[{ title: "Account" }, { title: "Profile" }, { title: "Done" }]}
        style={{ marginBottom: 24 }}
      />
      <Form layout="vertical">
        <Form.Item label="Username" name="username">
          <Input id="wizard-username" placeholder="Enter username" />
        </Form.Item>
      </Form>
      <Flex justify="flex-end">
        <Button variant="solid" color="primary" onClick={args.onFinish}>
          Next Step
        </Button>
      </Flex>
    </Card>
  ),
  // biome-ignore lint/suspicious/noExplicitAny: Storybook story args need any
  play: async ({ canvasElement, args }: any) => {
    const canvas = within(canvasElement)
    const input = canvas.getByPlaceholderText("Enter username")
    await userEvent.type(input, "testuser")
    const nextBtn = canvas.getByText("Next Step")
    await nextBtn.click()
    await expect(args.onFinish).toHaveBeenCalled()
  },
}

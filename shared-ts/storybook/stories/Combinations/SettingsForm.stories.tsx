import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Button,
  Card,
  Divider,
  Flex,
  Form,
  Input,
  Select,
  Slider,
  Space,
  Switch,
  Tabs,
  Typography,
} from "antd"

const meta: Meta<typeof Form> = {
  title: "Antd/Combinations/SettingsForm",
  component: Form,
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj<typeof Form>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div style={{ maxWidth: 700 }}>
      <Typography.Title level={4}>Settings</Typography.Title>
      <Tabs
        items={[
          {
            key: "profile",
            label: "Profile",
            children: (
              <Card>
                <Form
                  layout="vertical"
                  initialValues={{
                    name: "Alice Chen",
                    email: "alice@example.com",
                  }}
                >
                  <Form.Item label="Display Name" name="name">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Email" name="email">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Timezone" name="timezone">
                    <Select
                      options={[
                        { value: "utc-8", label: "Pacific Time (UTC-8)" },
                        { value: "utc-5", label: "Eastern Time (UTC-5)" },
                        { value: "utc+0", label: "UTC" },
                        { value: "utc+8", label: "China Standard (UTC+8)" },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item>
                    <Space>
                      <Button type="primary" variant="solid">
                        Save
                      </Button>
                      <Button variant="outlined">Cancel</Button>
                    </Space>
                  </Form.Item>
                </Form>
              </Card>
            ),
          },
          {
            key: "notifications",
            label: "Notifications",
            children: (
              <Card>
                <Flex vertical gap={16}>
                  <Flex justify="space-between" align="center">
                    <div>
                      <Typography.Text strong>
                        Email notifications
                      </Typography.Text>
                      <br />
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        Receive updates via email
                      </Typography.Text>
                    </div>
                    <Switch defaultChecked />
                  </Flex>
                  <Divider style={{ margin: 0 }} />
                  <Flex justify="space-between" align="center">
                    <div>
                      <Typography.Text strong>
                        Push notifications
                      </Typography.Text>
                      <br />
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        Browser push notifications
                      </Typography.Text>
                    </div>
                    <Switch />
                  </Flex>
                  <Divider style={{ margin: 0 }} />
                  <Flex justify="space-between" align="center">
                    <div>
                      <Typography.Text strong>
                        Notification frequency
                      </Typography.Text>
                      <br />
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        How often to receive digest emails
                      </Typography.Text>
                    </div>
                  </Flex>
                  <Slider defaultValue={30} />
                </Flex>
              </Card>
            ),
          },
        ]}
      />
    </div>
  ),
}

export const Playground: Story = {
  render: () => (
    <div style={{ maxWidth: 700 }}>
      <Typography.Title level={4}>Settings</Typography.Title>
      <Tabs
        items={[
          {
            key: "profile",
            label: "Profile",
            children: (
              <Card>
                <Form
                  layout="vertical"
                  initialValues={{
                    name: "Alice Chen",
                    email: "alice@example.com",
                  }}
                >
                  <Form.Item label="Display Name" name="name">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Email" name="email">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Timezone" name="timezone">
                    <Select
                      options={[
                        { value: "utc-8", label: "Pacific Time (UTC-8)" },
                        { value: "utc-5", label: "Eastern Time (UTC-5)" },
                        { value: "utc+0", label: "UTC" },
                        { value: "utc+8", label: "China Standard (UTC+8)" },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item>
                    <Space>
                      <Button type="primary" variant="solid">
                        Save
                      </Button>
                      <Button variant="outlined">Cancel</Button>
                    </Space>
                  </Form.Item>
                </Form>
              </Card>
            ),
          },
          {
            key: "notifications",
            label: "Notifications",
            children: (
              <Card>
                <Flex vertical gap={16}>
                  <Flex justify="space-between" align="center">
                    <div>
                      <Typography.Text strong>
                        Email notifications
                      </Typography.Text>
                      <br />
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        Receive updates via email
                      </Typography.Text>
                    </div>
                    <Switch defaultChecked />
                  </Flex>
                  <Divider style={{ margin: 0 }} />
                  <Slider defaultValue={30} />
                </Flex>
              </Card>
            ),
          },
        ]}
      />
    </div>
  ),
}

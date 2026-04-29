import { RocketOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Alert, Button, Flex, Space } from "antd"
import { expect, fn, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta<typeof Alert> = {
  title: "Antd/Feedback/Alert",
  component: Alert,
  parameters: { layout: "padded" },
  args: {
    type: "info",
    showIcon: true,
    closable: false,
    banner: false,
    title: "Alert Title",
  },
  argTypes: {
    type: {
      control: "select",
      options: ["success", "info", "warning", "error"],
    },
    showIcon: { control: "boolean" },
    closable: { control: "boolean" },
    banner: { control: "boolean" },
    title: { control: "text" },
    description: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof Alert>

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const types = ["success", "info", "warning", "error"] as const
    return (
      <Flex vertical gap={24}>
        <Section title="Types">
          <Flex vertical gap={12}>
            {types.map((t) => (
              <Alert
                key={t}
                type={t}
                title={`${t.charAt(0).toUpperCase() + t.slice(1)} Text`}
                showIcon
              />
            ))}
          </Flex>
        </Section>

        <Section title="With Description">
          <Flex vertical gap={12}>
            {types.map((t) => (
              <Alert
                key={t}
                type={t}
                title={`${t.charAt(0).toUpperCase() + t.slice(1)} Tips`}
                description={`Detailed description and advice about ${t} copywriting.`}
                showIcon
              />
            ))}
          </Flex>
        </Section>

        <Section title="Closable">
          <Flex vertical gap={12}>
            <Alert type="warning" title="Warning Text" closable />
            <Alert
              type="error"
              title="Error Text"
              description="More description about error."
              closable
              showIcon
            />
            <Alert
              type="info"
              title="Custom close icon"
              closable={{ closeIcon: "Close Now" }}
              showIcon
            />
          </Flex>
        </Section>

        <Section title="Banner Mode">
          <Flex vertical gap={12}>
            <Alert banner title="Warning text" />
            <Alert banner type="success" title="Success text" />
            <Alert banner type="error" title="Error text" closable />
            <Alert
              banner
              type="info"
              title="Banner with icon"
              showIcon
              icon={<RocketOutlined />}
            />
          </Flex>
        </Section>

        <Section title="With Action">
          <Flex vertical gap={12}>
            <Alert
              type="error"
              title="Connection Lost"
              action={
                <Button size="small" danger>
                  Retry
                </Button>
              }
              showIcon
            />
            <Alert
              type="warning"
              title="You have unsaved changes"
              description="Your changes will be lost if you navigate away."
              action={
                <Space>
                  <Button size="small" type="primary">
                    Save
                  </Button>
                  <Button size="small">Discard</Button>
                </Space>
              }
              showIcon
              closable
            />
            <Alert
              type="info"
              title="New version available"
              action={
                <Button size="small" type="link">
                  Upgrade
                </Button>
              }
              showIcon
              closable
            />
          </Flex>
        </Section>
      </Flex>
    )
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {}

// ── Interactive ──────────────────────────────────────────────────

export const Interactive: Story = {
  args: {
    type: "warning",
    title: "Closable Alert",
    closable: true,
    showIcon: true,
    onClose: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const closeBtn = canvas.getByRole("button", { name: /close/i })
    await closeBtn.click()
    await expect(args.onClose).toHaveBeenCalled()
  },
}

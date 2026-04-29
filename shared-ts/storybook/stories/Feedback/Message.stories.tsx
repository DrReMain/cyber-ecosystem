import type { Meta, StoryObj } from "@storybook/react-vite"
import { App, Button, Flex, Space } from "antd"
import { useEffect } from "react"
import { PageContent, Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Feedback/Message",
  parameters: { layout: "padded" },
  args: {
    type: "info",
    content: "This is a message",
    duration: 3,
  },
  argTypes: {
    type: {
      control: "select",
      options: ["success", "info", "warning", "error", "loading"],
    },
    content: { control: "text" },
    duration: { control: "number" },
  },
  decorators: [
    (Story) => (
      <App>
        <Story />
      </App>
    ),
  ],
}

export default meta
type Story = StoryObj

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const { message } = App.useApp()
    const types = [
      { key: "success", label: "Success" },
      { key: "info", label: "Info" },
      { key: "warning", label: "Warning" },
      { key: "error", label: "Error" },
      { key: "loading", label: "Loading" },
    ] as const

    return (
      <Flex vertical gap={24}>
        <Section title="Types">
          <Space wrap>
            {types.map((t) => (
              <Button
                key={t.key}
                onClick={() =>
                  message[t.key](`This is a ${t.label.toLowerCase()} message`)
                }
              >
                {t.label}
              </Button>
            ))}
          </Space>
        </Section>
      </Flex>
    )
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  render: () => {
    const { message } = App.useApp()
    return (
      <Button type="primary" onClick={() => message.info("Test")}>
        Show Message
      </Button>
    )
  },
}

// ── StaticOpen ───────────────────────────────────────────────────

function StaticMessages() {
  const { message } = App.useApp()

  useEffect(() => {
    message.success({
      content: "This is a success message",
      duration: 0,
      key: "static-success",
    })
    message.info({
      content: "This is an info message",
      duration: 0,
      key: "static-info",
    })
    message.warning({
      content: "This is a warning message",
      duration: 0,
      key: "static-warning",
    })
    message.error({
      content: "This is an error message",
      duration: 0,
      key: "static-error",
    })
    message.loading({
      content: "This is a loading message",
      duration: 0,
      key: "static-loading",
    })
  }, [message])

  return null
}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <>
      <PageContent />
      <StaticMessages />
    </>
  ),
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { App, Button, Flex, Space } from "antd"
import { useEffect } from "react"
import { PageContent, Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Feedback/Notification",
  parameters: { layout: "padded" },
  args: {
    type: "info",
    title: "Notification Title",
    description: "This is a notification.",
    duration: 4.5,
    placement: "topRight",
  },
  argTypes: {
    type: {
      control: "select",
      options: ["success", "info", "warning", "error"],
    },
    title: { control: "text" },
    description: { control: "text" },
    duration: { control: "number" },
    placement: {
      control: "select",
      options: [
        "top",
        "topLeft",
        "topRight",
        "bottom",
        "bottomLeft",
        "bottomRight",
      ],
    },
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
    const { notification } = App.useApp()
    const types = [
      { key: "success", label: "Success" },
      { key: "info", label: "Info" },
      { key: "warning", label: "Warning" },
      { key: "error", label: "Error" },
    ] as const

    return (
      <Flex vertical gap={24}>
        <Section title="Types">
          <Space wrap>
            {types.map((t) => (
              <Button
                key={t.key}
                onClick={() =>
                  notification[t.key]({
                    title: t.label,
                    description: `This is a ${t.label.toLowerCase()} notification.`,
                  })
                }
              >
                {t.label}
              </Button>
            ))}
          </Space>
        </Section>

        <Section title="Placements">
          <Space wrap>
            {(
              [
                "top",
                "topLeft",
                "topRight",
                "bottom",
                "bottomLeft",
                "bottomRight",
              ] as const
            ).map((placement) => (
              <Button
                key={placement}
                onClick={() =>
                  notification.open({
                    title: placement,
                    description: `This notification appears at ${placement}.`,
                    placement,
                  })
                }
              >
                {placement}
              </Button>
            ))}
          </Space>
        </Section>

        <Section title="With Progress Bar">
          <Space wrap>
            <Button
              onClick={() =>
                notification.success({
                  title: "Progress Bar",
                  description: "This notification has a progress bar.",
                  showProgress: true,
                })
              }
            >
              Success with progress
            </Button>
            <Button
              onClick={() =>
                notification.info({
                  title: "Progress + Pause on Hover",
                  description:
                    "Hover over this notification to pause the timer.",
                  showProgress: true,
                  pauseOnHover: true,
                })
              }
            >
              Pause on hover
            </Button>
          </Space>
        </Section>
      </Flex>
    )
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  render: () => {
    const { notification } = App.useApp()
    return (
      <Button
        type="primary"
        onClick={() => notification.info({ message: "Test" })}
      >
        Show Notification
      </Button>
    )
  },
}

// ── StaticOpen ───────────────────────────────────────────────────

function StaticNotifications() {
  const { notification } = App.useApp()

  useEffect(() => {
    notification.success({
      key: "static-success",
      message: "Success",
      description: "This is a success notification.",
      duration: 0,
    })
    notification.info({
      key: "static-info",
      message: "Info",
      description: "This is an info notification.",
      duration: 0,
    })
    notification.warning({
      key: "static-warning",
      message: "Warning",
      description: "This is a warning notification.",
      duration: 0,
    })
    notification.error({
      key: "static-error",
      message: "Error",
      description: "This is an error notification.",
      duration: 0,
    })

    const timer = setTimeout(() => {
      const container = document.querySelector(".ant-notification-stack")
      container?.classList.remove("ant-notification-stack")
      void container
    }, 100)

    return () => {
      clearTimeout(timer)
      notification.destroy("static-success")
      notification.destroy("static-info")
      notification.destroy("static-warning")
      notification.destroy("static-error")
    }
  }, [notification])

  return null
}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <>
      <PageContent />
      <StaticNotifications />
    </>
  ),
}

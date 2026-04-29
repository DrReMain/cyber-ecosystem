import type { Meta, StoryObj } from "@storybook/react-vite"
import { App, Button, Flex, Space } from "antd"
import { expect, userEvent, within } from "storybook/test"
import { Label, Section } from "../helpers"

const { useApp } = App

function StaticMethodsDemo() {
  const { message } = useApp()
  return (
    <Space wrap>
      <Button
        onClick={() => {
          message.success("Success!")
        }}
      >
        message.success
      </Button>
      <Button
        onClick={() => {
          message.error("Error!")
        }}
      >
        message.error
      </Button>
      <Button
        onClick={() => {
          message.warning("Warning!")
        }}
      >
        message.warning
      </Button>
      <Button
        onClick={() => {
          message.info("Info!")
        }}
      >
        message.info
      </Button>
    </Space>
  )
}

function NotificationDemo() {
  const { notification } = useApp()
  return (
    <Button
      onClick={() => {
        notification.open({
          message: "Notification Title",
          description:
            "This is the content of the notification rendered via App.useApp().",
        })
      }}
    >
      notification.open
    </Button>
  )
}

function ModalDemo() {
  const { modal } = useApp()
  return (
    <Button
      onClick={() => {
        modal.warning({
          title: "Warning",
          content: "This is a warning modal rendered via App.useApp().",
        })
      }}
    >
      modal.warning
    </Button>
  )
}

const meta: Meta<typeof App> = {
  title: "Antd/General/App",
  component: App,
  parameters: { layout: "padded" },
  argTypes: {},
}

export default meta
type Story = StoryObj<typeof App>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <App>
      <Flex vertical gap={24}>
        <Section title="Static Methods">
          <Flex vertical gap={4}>
            <Label>
              message.success / message.error / message.warning / message.info
              via App.useApp()
            </Label>
            <StaticMethodsDemo />
          </Flex>
        </Section>

        <Section title="Notification">
          <Flex vertical gap={4}>
            <Label>notification.open via App.useApp()</Label>
            <NotificationDemo />
          </Flex>
        </Section>

        <Section title="Modal">
          <Flex vertical gap={4}>
            <Label>modal.warning via App.useApp()</Label>
            <ModalDemo />
          </Flex>
        </Section>
      </Flex>
    </App>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <App {...args}>
      <Flex vertical gap={12}>
        <StaticMethodsDemo />
        <NotificationDemo />
        <ModalDemo />
      </Flex>
    </App>
  ),
}

export const Interactive: Story = {
  render: () => (
    <App>
      <InteractiveInner />
    </App>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const btn = canvas.getByRole("button", { name: "Trigger success" })
    await userEvent.click(btn)
    // Verify the message appears in the DOM
    const msg = await canvas.findByText("Interactive success!")
    await expect(msg).toBeInTheDocument()
  },
}

function InteractiveInner() {
  const { message } = useApp()
  return (
    <Button
      onClick={() => {
        message.success("Interactive success!")
      }}
    >
      Trigger success
    </Button>
  )
}

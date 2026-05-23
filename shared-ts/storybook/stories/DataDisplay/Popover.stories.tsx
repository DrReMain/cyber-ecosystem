import type { Meta, StoryObj } from "@storybook/react-vite"
import { Avatar, Button, Flex, Input, Popover, Tag, Typography } from "antd"
import { useRef } from "react"
import { expect, fn, within } from "storybook/test"
import { PageContent, Section } from "../helpers"

const { Text } = Typography

const meta: Meta<typeof Popover> = {
  title: "Antd/Data Display/Popover",
  component: Popover,
  parameters: { layout: "padded" },
  args: {
    title: "Popover Title",
    content: "Popover content",
    trigger: "hover",
  },
  argTypes: {
    trigger: {
      control: "radio",
      options: ["hover", "click", "focus"],
    },
    placement: {
      control: "select",
      options: [
        "top",
        "topLeft",
        "topRight",
        "bottom",
        "bottomLeft",
        "bottomRight",
        "left",
        "leftTop",
        "leftBottom",
        "right",
        "rightTop",
        "rightBottom",
      ],
    },
    onOpenChange: { action: "openChanged" },
  },
}

export default meta
type Story = StoryObj<typeof Popover>

const textContent = (
  <div>
    <p style={{ margin: 0 }}>Some content for the popover.</p>
    <p style={{ margin: 0 }}>It can hold any React node.</p>
  </div>
)

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Popover title="Popover Title" content={textContent}>
          <Button type="primary">Hover me</Button>
        </Popover>
      </Section>

      <Section
        title="With Action Buttons"
        description="Popover containing a confirmation form with actions"
      >
        <Flex gap={16}>
          <Popover
            title="Delete this item?"
            content={
              <Flex vertical gap={8}>
                <Text type="secondary">This action cannot be undone.</Text>
                <Flex justify="flex-end" gap={8}>
                  <Button size="small">Cancel</Button>
                  <Button size="small" danger variant="solid">
                    Delete
                  </Button>
                </Flex>
              </Flex>
            }
            trigger="click"
          >
            <Button danger>Delete</Button>
          </Popover>
          <Popover
            title="Share document"
            content={
              <Flex vertical gap={8}>
                <Input placeholder="Enter email address" size="small" />
                <Flex justify="flex-end">
                  <Button size="small" variant="solid" color="primary">
                    Send
                  </Button>
                </Flex>
              </Flex>
            }
            trigger="click"
          >
            <Button>Share</Button>
          </Popover>
        </Flex>
      </Section>

      <Section
        title="Custom Content"
        description="Popover with rich content like user cards or mini tables"
      >
        <Flex gap={16}>
          <Popover
            title={null}
            content={
              <Flex gap={12} style={{ padding: 4 }}>
                <Avatar style={{ backgroundColor: "#1677ff" }}>JD</Avatar>
                <Flex vertical>
                  <Text strong>John Doe</Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    john@example.com
                  </Text>
                  <Tag color="blue" style={{ marginTop: 4 }}>
                    Admin
                  </Tag>
                </Flex>
              </Flex>
            }
            trigger="hover"
          >
            <Button type="link">Hover for user info</Button>
          </Popover>
          <Popover
            title="Recent Activity"
            trigger="click"
            content={
              <Flex vertical gap={8} style={{ minWidth: 200 }}>
                {[
                  { text: "Updated profile", time: "2 min ago" },
                  { text: "Created project", time: "1 hour ago" },
                  { text: "Uploaded file", time: "3 hours ago" },
                ].map((item) => (
                  <Flex key={item.text} justify="space-between">
                    <Text style={{ fontSize: 12 }}>{item.text}</Text>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      {item.time}
                    </Text>
                  </Flex>
                ))}
              </Flex>
            }
          >
            <Button type="link">Click for activity</Button>
          </Popover>
        </Flex>
      </Section>

      <Section
        title="Trigger Variants"
        description="Click vs hover trigger modes"
      >
        <Flex gap={16}>
          <Popover content="Triggered by click" trigger="click">
            <Button>Click me</Button>
          </Popover>
          <Popover content="Triggered by hover" trigger="hover">
            <Button>Hover me</Button>
          </Popover>
          <Popover content="Triggered by focus" trigger="focus">
            <Button>Focus me</Button>
          </Popover>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const StaticOpen: Story = {
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div
        ref={containerRef}
        style={{ position: "relative", minHeight: 720, padding: "120px 0" }}
      >
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Flex vertical gap={160} align="center">
            {(["top", "bottom", "left", "right"] as const).map((placement) => (
              <Popover
                key={placement}
                title={`Static ${placement}`}
                content={textContent}
                placement={placement}
                open
                getPopupContainer={() => containerRef.current ?? document.body}
              >
                <Button>{placement}</Button>
              </Popover>
            ))}
          </Flex>
        </div>
      </div>
    )
  },
}

export const Playground: Story = {
  render: (args) => (
    <Popover {...args}>
      <Button type="primary">Hover me</Button>
    </Popover>
  ),
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    title: "Interactive Title",
    content: "Interactive content",
    trigger: "click",
    onOpenChange: fn(),
  },
  render: (args) => (
    <Popover {...args}>
      <Button>Click me</Button>
    </Popover>
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.getByText("Click me").click()
    await expect(args.onOpenChange).toHaveBeenCalled()
  },
}

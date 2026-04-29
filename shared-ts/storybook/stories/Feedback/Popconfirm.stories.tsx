import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Popconfirm, Space } from "antd"
import { useRef } from "react"
import { expect, fn, within } from "storybook/test"
import { PageContent, Section } from "../helpers"

const meta: Meta<typeof Popconfirm> = {
  title: "Antd/Feedback/Popconfirm",
  component: Popconfirm,
  parameters: { layout: "padded" },
  args: {
    title: "Are you sure?",
    okText: "Yes",
    cancelText: "No",
  },
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
    okText: { control: "text" },
    cancelText: { control: "text" },
    okType: {
      control: "select",
      options: ["primary", "danger", "default", "dashed", "link", "text"],
    },
    onConfirm: { action: "confirmed" },
    onCancel: { action: "cancelled" },
  },
}

export default meta
type Story = StoryObj<typeof Popconfirm>

const placements = [
  "top",
  "topLeft",
  "topRight",
  "bottom",
  "bottomLeft",
  "bottomRight",
] as const

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Popconfirm
          title="Delete the task"
          onConfirm={() => {}}
          onCancel={() => {}}
        >
          <Button danger>Delete</Button>
        </Popconfirm>
      </Section>

      <Section title="With Description">
        <Popconfirm
          title="Delete the task"
          description="Are you sure to delete this task? This action cannot be undone."
          onConfirm={() => {}}
          onCancel={() => {}}
        >
          <Button danger>Delete</Button>
        </Popconfirm>
      </Section>

      <Section title="Placements">
        <Space wrap>
          {placements.map((placement) => (
            <Popconfirm
              key={placement}
              title={`Placement: ${placement}`}
              placement={placement}
              onConfirm={() => {}}
            >
              <Button>{placement}</Button>
            </Popconfirm>
          ))}
        </Space>
      </Section>
    </Flex>
  ),
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  args: {
    title: "Are you sure?",
    onConfirm: fn(),
    onCancel: fn(),
  },
  render: (args) => (
    <Popconfirm {...args}>
      <Button danger>Delete</Button>
    </Popconfirm>
  ),
}

// ── StaticOpen ───────────────────────────────────────────────────

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div ref={containerRef} style={{ position: "relative", minHeight: 200 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Popconfirm
            title="Delete the task"
            description="Are you sure to delete this task? This action cannot be undone."
            open
            onConfirm={() => {}}
            onCancel={() => {}}
            getPopupContainer={() => containerRef.current!}
          >
            <Button danger>Delete</Button>
          </Popconfirm>
        </div>
      </div>
    )
  },
}

// ── Interactive story with play function ──────────────────────────

export const Interactive: Story = {
  args: {
    title: "Delete this item?",
    description: "This action cannot be undone.",
    onConfirm: fn(),
    onCancel: fn(),
  },
  render: (args) => (
    <Popconfirm {...args}>
      <Button danger>Delete</Button>
    </Popconfirm>
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole("button", { name: "Delete" })
    await button.click()
    const confirmBtn = await within(document.body).findByRole("button", {
      name: /yes/i,
    })
    await confirmBtn.click()
    await expect(args.onConfirm).toHaveBeenCalledTimes(1)
  },
}

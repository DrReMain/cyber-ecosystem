import type { Meta, StoryObj } from "@storybook/react-vite"
import { Descriptions, Flex, Space } from "antd"
import { expect, within } from "storybook/test"
import { borderedArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Descriptions> = {
  title: "Antd/Data Display/Descriptions",
  component: Descriptions,
  parameters: { layout: "padded" },
  args: {
    title: "User Info",
    items: [
      { key: "1", label: "Product", children: "Cloud Database" },
      { key: "2", label: "Billing", children: "Prepaid" },
      { key: "3", label: "Time", children: "18:00:00" },
      { key: "4", label: "Amount", children: "$80.00" },
    ],
  },
  argTypes: {
    size: sizeArg(["small", "medium", "large"]),
    bordered: borderedArg,
    layout: {
      control: "radio",
      options: ["horizontal", "vertical"],
    },
    column: { control: { type: "number", min: 1, max: 6 } },
  },
}

export default meta
type Story = StoryObj<typeof Descriptions>

const items = [
  { key: "1", label: "Product", children: "Cloud Database" },
  { key: "2", label: "Billing", children: "Prepaid" },
  { key: "3", label: "Time", children: "18:00:00" },
  { key: "4", label: "Amount", children: "$80.00" },
  { key: "5", label: "Discount", children: "$20.00" },
  { key: "6", label: "Official", children: "$60.00" },
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Bordered" description="Table-like layout with borders">
        <Descriptions title="User Info" bordered items={items} />
      </Section>

      <Section title="Sizes" description="small / medium / large (default)">
        <Flex vertical gap={16}>
          {(["small", "medium", "large"] as const).map((size) => (
            <Space
              key={size}
              orientation="vertical"
              size={4}
              style={{ width: "100%" }}
            >
              <Label>{size}</Label>
              <Descriptions
                title={`Size: ${size}`}
                size={size}
                bordered
                items={items.slice(0, 4)}
              />
            </Space>
          ))}
        </Flex>
      </Section>

      <Section title="Column Count" description="Fixed column count (2, 3, 4)">
        <Flex vertical gap={16}>
          {[2, 3, 4].map((col) => (
            <Space
              key={col}
              orientation="vertical"
              size={4}
              style={{ width: "100%" }}
            >
              <Label>{col} columns</Label>
              <Descriptions
                title={`${col} Columns`}
                bordered
                column={col}
                items={items}
              />
            </Space>
          ))}
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    title: "User Info",
    items: items.slice(0, 4),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const labels = canvas.getAllByText(/Product|Billing|Time|Amount/)
    await expect(labels.length).toBeGreaterThan(0)
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Collapse, Flex, Space } from "antd"
import { expect, fn, within } from "storybook/test"
import { borderedArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Collapse> = {
  title: "Antd/Data Display/Collapse",
  component: Collapse,
  parameters: { layout: "padded" },
  args: {
    items: [
      { key: "1", label: "Panel 1", children: <p>Content of panel 1.</p> },
      { key: "2", label: "Panel 2", children: <p>Content of panel 2.</p> },
      { key: "3", label: "Panel 3", children: <p>Content of panel 3.</p> },
    ],
  },
  argTypes: {
    size: sizeArg(["small", "medium", "large"]),
    bordered: borderedArg,
    accordion: { control: "boolean" },
    ghost: { control: "boolean" },
    expandIconPlacement: {
      control: "radio",
      options: ["start", "end"],
    },
  },
}

export default meta
type Story = StoryObj<typeof Collapse>

const baseItems = [
  { key: "1", label: "Panel 1", children: <p>Content of panel 1.</p> },
  { key: "2", label: "Panel 2", children: <p>Content of panel 2.</p> },
  { key: "3", label: "Panel 3", children: <p>Content of panel 3.</p> },
]

const sizes: Array<React.ComponentProps<typeof Collapse>["size"]> = [
  "small",
  "medium",
  "large",
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes" description="small / medium (default) / large">
        <Flex vertical gap={16}>
          {sizes.map((size) => (
            <Space
              key={size}
              orientation="vertical"
              size={4}
              style={{ width: "100%" }}
            >
              <Label>{size}</Label>
              <Collapse size={size} items={baseItems} />
            </Space>
          ))}
        </Flex>
      </Section>

      <Section title="Accordion" description="Only one panel open at a time">
        <Collapse accordion items={baseItems} />
      </Section>

      <Section title="Ghost" description="No border, transparent background">
        <Collapse ghost items={baseItems} />
      </Section>

      <Section title="Ghost x Accordion">
        <Collapse ghost accordion items={baseItems} />
      </Section>

      <Section title="Nested" description="Collapse inside a collapse panel">
        <Collapse
          items={[
            {
              key: "outer-1",
              label: "Outer Panel 1",
              children: (
                <Collapse
                  size="small"
                  items={[
                    {
                      key: "inner-1",
                      label: "Inner Panel 1",
                      children: <p>Nested content 1</p>,
                    },
                    {
                      key: "inner-2",
                      label: "Inner Panel 2",
                      children: <p>Nested content 2</p>,
                    },
                  ]}
                />
              ),
            },
            {
              key: "outer-2",
              label: "Outer Panel 2",
              children: <p>Regular content without nesting.</p>,
            },
          ]}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    items: [
      {
        key: "1",
        label: "Click to expand",
        children: <p>Expanded content for panel 1.</p>,
      },
      {
        key: "2",
        label: "Another panel",
        children: <p>Expanded content for panel 2.</p>,
      },
    ],
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const headers = canvas.getAllByRole("button")
    await headers[0].click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Tabs } from "antd"
import { expect, fn, within } from "storybook/test"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const basicItems = [
  { key: "1", label: "Tab 1", children: "Content of Tab 1" },
  { key: "2", label: "Tab 2", children: "Content of Tab 2" },
  { key: "3", label: "Tab 3", children: "Content of Tab 3" },
]

const sizes = ["small", "medium", "large"] as const

const meta: Meta<typeof Tabs> = {
  title: "Antd/Navigation/Tabs",
  component: Tabs,
  parameters: { layout: "padded" },
  args: {
    items: basicItems,
  },
  argTypes: {
    type: {
      control: "radio",
      options: ["line", "card", "editable-card"],
    },
    size: sizeArg(["small", "medium", "large"]),
    tabPlacement: {
      control: "radio",
      options: ["top", "bottom", "start", "end"],
    },
    centered: { control: "boolean" },
    animated: { control: "boolean" },
    onChange: { action: "changed" },
    onEdit: { action: "edited" },
  },
}

export default meta
type Story = StoryObj<typeof Tabs>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      {/* 1. Basic */}
      <Section title="Basic (line type)">
        <Tabs items={basicItems} />
      </Section>

      {/* 2. Sizes */}
      <Section title="Sizes">
        <Flex vertical gap={16}>
          {sizes.map((s) => (
            <Flex key={s} vertical gap={4}>
              <Label>{s}</Label>
              <Tabs size={s} items={basicItems} />
            </Flex>
          ))}
        </Flex>
      </Section>

      {/* 3. Card Type */}
      <Section title="Card Type">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>card</Label>
            <Tabs type="card" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>editable-card</Label>
            <Tabs type="editable-card" items={basicItems} onEdit={() => {}} />
          </Flex>
        </Flex>
      </Section>

      {/* 4. Placement */}
      <Section title="Placement">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>top</Label>
            <Tabs tabPlacement="top" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>bottom</Label>
            <Tabs tabPlacement="bottom" items={basicItems} />
          </Flex>
          <Flex gap={24}>
            <Flex vertical gap={4} style={{ flex: 1 }}>
              <Label>start (left)</Label>
              <Tabs tabPlacement="start" items={basicItems} />
            </Flex>
            <Flex vertical gap={4} style={{ flex: 1 }}>
              <Label>end (right)</Label>
              <Tabs tabPlacement="end" items={basicItems} />
            </Flex>
          </Flex>
        </Flex>
      </Section>

      {/* 5. Disabled Tab */}
      <Section title="Disabled Tab">
        <Tabs
          items={[
            { key: "1", label: "Active", children: "Content of active tab" },
            {
              key: "2",
              label: "Disabled",
              children: "Content of disabled tab",
              disabled: true,
            },
            { key: "3", label: "Another", children: "Content of another tab" },
          ]}
        />
      </Section>

      <Section title="States">
        <PseudoStates>
          <Tabs items={basicItems} />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    items: basicItems,
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const tab2 = canvas.getByRole("tab", { name: "Tab 2" })
    await tab2.click()
    await expect(args.onChange).toHaveBeenCalledWith("2")
  },
}

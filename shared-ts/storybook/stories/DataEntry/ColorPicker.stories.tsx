import type { Meta, StoryObj } from "@storybook/react-vite"
import { ColorPicker, Flex } from "antd"
import { useRef } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { callbackArgs, disabledArg } from "../argTypes"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof ColorPicker> = {
  title: "Antd/Data Entry/ColorPicker",
  component: ColorPicker,
  parameters: { layout: "padded" },
  args: { defaultValue: "#1677ff" },
  argTypes: {
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof ColorPicker>

const sizes = ["small", "medium", "large"] as const
const modes = [
  { label: "single", value: "single" as const },
  { label: "gradient", value: "gradient" as const },
  {
    label: "both (switchable)",
    value: ["single", "gradient"] as ["single", "gradient"],
  },
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <ColorPicker defaultValue="#1677ff" />
      </Section>

      <Section title="Sizes">
        <Flex gap={16} align="center">
          {sizes.map((size) => (
            <Flex key={size} vertical gap={4} align="center">
              <Label>{size}</Label>
              <ColorPicker size={size} defaultValue="#1677ff" />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled States">
        <Flex gap={16} align="center">
          <Flex vertical gap={4} align="center">
            <Label>disabled</Label>
            <ColorPicker disabled defaultValue="#1677ff" />
          </Flex>
          <Flex vertical gap={4} align="center">
            <Label>disabledAlpha</Label>
            <ColorPicker disabledAlpha defaultValue="#1677ff" />
          </Flex>
          <Flex vertical gap={4} align="center">
            <Label>disabledFormat</Label>
            <ColorPicker disabledFormat defaultValue="#1677ff" />
          </Flex>
        </Flex>
      </Section>

      <Section title="Modes">
        <Flex gap={16} align="center">
          {modes.map((mode) => (
            <Flex key={mode.label} vertical gap={4} align="center">
              <Label>{mode.label}</Label>
              <ColorPicker mode={mode.value} defaultValue="#1677ff" />
            </Flex>
          ))}
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div ref={containerRef} style={{ position: "relative", minHeight: 400 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <ColorPicker
            open
            defaultValue="#1677ff"
            getPopupContainer={() => containerRef.current!}
          />
        </div>
      </div>
    )
  },
}

export const Interactive: Story = {
  args: {
    defaultValue: "#1677ff",
    onChange: fn(),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const trigger = canvas.getByRole("button")
    await userEvent.click(trigger)
    const popup = await within(document.body).findByRole("dialog")
    await expect(popup).toBeInTheDocument()
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Rate } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Rate> = {
  title: "Antd/Data Entry/Rate",
  component: Rate,
  parameters: { layout: "padded" },
  args: { defaultValue: 3, count: 5 },
  argTypes: {
    count: { control: { type: "number", min: 1, max: 20 } },
    allowHalf: { control: "boolean" },
    allowClear: { control: "boolean" },
    disabled: disabledArg,
    tooltips: { control: "object" },
    size: sizeArg(["small", "medium", "large"]),
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Rate>

const sizes = ["small", "medium", "large"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Rate defaultValue={3} />
      </Section>

      <Section title="Allow Half">
        <Rate allowHalf defaultValue={3.5} />
      </Section>

      <Section title="Custom Count">
        <Flex vertical gap={8}>
          <Label>3 stars</Label>
          <Rate count={3} defaultValue={2} />
          <Label>5 stars (default)</Label>
          <Rate count={5} defaultValue={3} />
          <Label>10 stars</Label>
          <Rate count={10} defaultValue={6} />
        </Flex>
      </Section>

      <Section title="Sizes">
        <Flex vertical gap={8}>
          {sizes.map((size) => (
            <Flex key={size} align="center" gap={8}>
              <Label>{size}</Label>
              <Rate size={size} defaultValue={3} />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex vertical gap={8}>
          <Rate disabled defaultValue={3} />
          <Rate disabled allowHalf defaultValue={2.5} />
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Rate defaultValue={3} />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { defaultValue: 3 },
}

export const Interactive: Story = {
  args: {
    defaultValue: 0,
    count: 5,
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const stars = canvas.getAllByRole("radio")
    await stars[2].click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

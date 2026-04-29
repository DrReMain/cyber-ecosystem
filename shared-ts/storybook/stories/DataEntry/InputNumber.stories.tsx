import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, InputNumber } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  callbackArgs,
  disabledArg,
  sizeArg,
  statusArg,
  variantArg,
} from "../argTypes"
import { Label, Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof InputNumber> = {
  title: "Antd/Data Entry/InputNumber",
  component: InputNumber,
  parameters: { layout: "padded" },
  args: { defaultValue: 0 },
  argTypes: {
    size: sizeArg(),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof InputNumber>

const sizes = ["large", "medium", "small"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["warning", "error"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex vertical gap={12}>
          {sizes.map((size) => (
            <InputNumber
              key={size}
              size={size}
              defaultValue={3}
              placeholder={size.charAt(0).toUpperCase() + size.slice(1)}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12}>
          {variants.map((variant) => (
            <InputNumber
              key={variant}
              variant={variant}
              defaultValue={100}
              placeholder={variant.charAt(0).toUpperCase() + variant.slice(1)}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12}>
          {statuses.map((status) => (
            <InputNumber
              key={status}
              status={status}
              defaultValue={5}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Disabled & ReadOnly">
        <Flex vertical gap={12}>
          <InputNumber disabled defaultValue={5} style={{ width: W.input }} />
          <InputNumber readOnly defaultValue={42} style={{ width: W.input }} />
        </Flex>
      </Section>

      <Section title="Controls">
        <Flex vertical gap={12}>
          <Label>With controls (default)</Label>
          <InputNumber defaultValue={5} style={{ width: W.input }} />
          <Label>Without controls</Label>
          <InputNumber
            controls={false}
            defaultValue={5}
            style={{ width: W.input }}
          />
          <Label>Custom control icons</Label>
          <InputNumber
            defaultValue={5}
            controls={{
              upIcon: <span>&#9650;</span>,
              downIcon: <span>&#9660;</span>,
            }}
            style={{ width: W.input }}
          />
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <InputNumber defaultValue={0} />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    defaultValue: 0,
    onChange: fn(),
    style: { width: 200 },
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("spinbutton")
    await input.focus()
    await userEvent.type(input, "42")
    await expect(args.onChange).toHaveBeenCalled()
  },
}

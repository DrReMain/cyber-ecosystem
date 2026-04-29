import { InfoCircleOutlined, UserOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Input } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  callbackArgs,
  disabledArg,
  sizeArg,
  statusArg,
  variantArg,
} from "../argTypes"
import { Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Input> = {
  title: "Antd/Data Entry/Input",
  component: Input,
  parameters: { layout: "padded" },
  args: { placeholder: "Type something" },
  argTypes: {
    size: sizeArg(),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof Input>

const sizes = ["large", "medium", "small"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["warning", "error"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex vertical gap={12} style={{ width: W.input }}>
          {sizes.map((size) => (
            <Input
              key={size}
              placeholder={size.charAt(0).toUpperCase() + size.slice(1)}
              size={size}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12} style={{ width: W.input }}>
          {variants.map((variant) => (
            <Input
              key={variant}
              variant={variant}
              placeholder={variant.charAt(0).toUpperCase() + variant.slice(1)}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12} style={{ width: W.input }}>
          {statuses.map((status) => (
            <Input
              key={status}
              status={status}
              placeholder={`${status} status`}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Prefix & Suffix">
        <Flex vertical gap={12} style={{ width: W.input }}>
          <Input prefix={<UserOutlined />} placeholder="With prefix" />
          <Input suffix={<InfoCircleOutlined />} placeholder="With suffix" />
          <Input
            prefix={<UserOutlined />}
            suffix={<InfoCircleOutlined />}
            placeholder="Both prefix and suffix"
          />
        </Flex>
      </Section>

      <Section title="Disabled & ReadOnly">
        <Flex vertical gap={12} style={{ width: W.input }}>
          <Input disabled placeholder="Disabled input" />
          <Input disabled defaultValue="Disabled with value" />
          <Input readOnly defaultValue="Read-only value" />
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Input placeholder="Type something" style={{ width: W.pseudo }} />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    placeholder: "Type here",
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("textbox")
    await input.focus()
    await userEvent.type(input, "test")
    await expect(args.onChange).toHaveBeenCalled()
  },
}

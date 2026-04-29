import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Input } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { disabledArg, sizeArg, statusArg, variantArg } from "../argTypes"
import { Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const { TextArea } = Input

const meta: Meta<typeof TextArea> = {
  title: "Antd/Data Entry/TextArea",
  component: TextArea,
  parameters: { layout: "padded" },
  args: { placeholder: "Type something" },
  argTypes: {
    size: sizeArg(),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
  },
}

export default meta
type Story = StoryObj<typeof TextArea>

const sizes = ["large", "medium", "small"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["warning", "error"] as const

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          {sizes.map((size) => (
            <TextArea
              key={size}
              placeholder={size.charAt(0).toUpperCase() + size.slice(1)}
              size={size}
              rows={2}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          {variants.map((variant) => (
            <TextArea
              key={variant}
              variant={variant}
              placeholder={variant.charAt(0).toUpperCase() + variant.slice(1)}
              rows={2}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          {statuses.map((status) => (
            <TextArea
              key={status}
              status={status}
              placeholder={`${status} status`}
              rows={2}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Auto Size">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          <TextArea
            placeholder="Auto size (min 2, max 6)"
            autoSize={{ minRows: 2, maxRows: 6 }}
          />
          <TextArea placeholder="Auto size (true)" autoSize />
        </Flex>
      </Section>

      <Section title="With Count">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          <TextArea placeholder="With count" showCount rows={2} />
          <TextArea
            placeholder="With count and max length"
            showCount
            maxLength={100}
            rows={2}
          />
        </Flex>
      </Section>

      <Section title="Disabled & ReadOnly">
        <Flex vertical gap={12} style={{ width: W.textarea }}>
          <TextArea disabled placeholder="Disabled textarea" rows={2} />
          <TextArea disabled defaultValue="Disabled with value" rows={2} />
          <TextArea readOnly defaultValue="Read-only value" rows={2} />
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <TextArea
            placeholder="Type something"
            rows={2}
            style={{ width: W.pseudo }}
          />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    placeholder: "Type here",
    rows: 3,
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const textarea = canvas.getByRole("textbox")
    await textarea.focus()
    await userEvent.type(textarea, "hello")
    await expect(args.onChange).toHaveBeenCalled()
  },
}

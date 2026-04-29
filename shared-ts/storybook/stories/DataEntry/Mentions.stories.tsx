import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Mentions } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  callbackArgs,
  disabledArg,
  sizeArg,
  statusArg,
  variantArg,
} from "../argTypes"
import { Section, W } from "../helpers"

const mentionOptions = [
  { value: "afc163", label: "afc163" },
  { value: "zombieJ", label: "zombieJ" },
  { value: "yesmeck", label: "yesmeck" },
]

const meta: Meta<typeof Mentions> = {
  title: "Antd/Data Entry/Mentions",
  component: Mentions,
  parameters: { layout: "padded" },
  args: {
    placeholder: "Type @ to mention someone",
    style: { width: W.textarea },
    options: mentionOptions,
  },
  argTypes: {
    size: sizeArg(),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof Mentions>

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
            <Mentions
              key={size}
              size={size}
              placeholder={size.charAt(0).toUpperCase() + size.slice(1)}
              style={{ width: W.textarea }}
              options={mentionOptions}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12}>
          {variants.map((variant) => (
            <Mentions
              key={variant}
              variant={variant}
              placeholder={variant.charAt(0).toUpperCase() + variant.slice(1)}
              style={{ width: W.textarea }}
              options={mentionOptions}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12}>
          {statuses.map((status) => (
            <Mentions
              key={status}
              status={status}
              placeholder={`${status} status`}
              style={{ width: W.textarea }}
              options={mentionOptions}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Mentions
          disabled
          placeholder="Disabled mentions"
          style={{ width: W.textarea }}
          options={mentionOptions}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    options: mentionOptions,
    onChange: fn(),
    placeholder: "Type @ to mention someone",
    style: { width: W.textarea },
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const textarea = canvas.getByRole("textbox")
    await textarea.focus()
    await userEvent.type(textarea, "@")
    await expect(args.onChange).toHaveBeenCalled()
  },
}

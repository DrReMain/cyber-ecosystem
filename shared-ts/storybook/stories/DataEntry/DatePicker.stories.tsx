import type { Meta, StoryObj } from "@storybook/react-vite"
import { DatePicker, Flex } from "antd"
import { useRef } from "react"
import { expect, fn, within } from "storybook/test"
import {
  callbackArgs,
  disabledArg,
  sizeArg,
  statusArg,
  variantArg,
} from "../argTypes"
import { Label, PageContent, Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof DatePicker> = {
  title: "Antd/Data Entry/DatePicker",
  component: DatePicker,
  parameters: { layout: "padded" },
  args: { placeholder: "Select date" },
  argTypes: {
    size: sizeArg(["small", "medium", "large"]),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof DatePicker>

const pickers = ["date", "week", "month", "quarter", "year"] as const
const sizes = ["small", "medium", "large"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["warning", "error"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex gap={12} align="center">
          {sizes.map((size) => (
            <Flex key={size} vertical gap={4} align="center">
              <Label>{size}</Label>
              <DatePicker
                size={size}
                placeholder={size}
                style={{ width: W.input }}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12}>
          {variants.map((variant) => (
            <Flex key={variant} align="center" gap={12}>
              <Label>{variant}</Label>
              <DatePicker
                variant={variant}
                placeholder={variant}
                style={{ width: W.input }}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex gap={12} align="center">
          {statuses.map((status) => (
            <DatePicker
              key={status}
              status={status}
              placeholder={status}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <DatePicker
          disabled
          placeholder="Disabled"
          style={{ width: W.input }}
        />
      </Section>

      <Section title="Picker Types">
        <Flex vertical gap={12}>
          {pickers.map((picker) => (
            <Flex key={picker} align="center" gap={12}>
              <Label>{picker}</Label>
              <DatePicker
                picker={picker}
                placeholder={`Select ${picker}`}
                style={{ width: W.input }}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <DatePicker />
        </PseudoStates>
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
      <div ref={containerRef} style={{ position: "relative", minHeight: 700 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Flex vertical>
            <div style={{ marginBottom: 320 }}>
              <DatePicker
                open
                getPopupContainer={() => containerRef.current!}
                placeholder="DatePicker open"
              />
            </div>
            <DatePicker.RangePicker
              open
              getPopupContainer={() => containerRef.current!}
            />
          </Flex>
        </div>
      </div>
    )
  },
}

export const Interactive: Story = {
  args: {
    onChange: fn(),
    placeholder: "Click to open",
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("textbox")
    await input.click()
    const body = within(document.body)
    const todayCell = body.getByTitle("Today")
    await todayCell.click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

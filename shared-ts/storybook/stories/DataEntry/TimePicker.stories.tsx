import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, TimePicker } from "antd"
import { useRef } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { disabledArg, sizeArg, statusArg, variantArg } from "../argTypes"
import { Label, PageContent, Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof TimePicker> = {
  title: "Antd/Data Entry/TimePicker",
  component: TimePicker,
  parameters: { layout: "padded" },
  args: { placeholder: "Select time" },
  argTypes: {
    format: { control: "text" },
    use12Hours: { control: "boolean" },
    disabled: disabledArg,
    status: statusArg,
    size: sizeArg(["small", "medium", "large"]),
    variant: variantArg(),
    showNow: { control: "boolean" },
    onChange: { action: "changed" },
    onOpenChange: { action: "openChanged" },
  },
}

export default meta
type Story = StoryObj<typeof TimePicker>

const { RangePicker } = TimePicker

const sizes = ["small", "medium", "large"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex gap={16} align="center">
          {sizes.map((size) => (
            <TimePicker
              key={size}
              size={size}
              placeholder={size}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex gap={16} align="center">
          {variants.map((variant) => (
            <TimePicker
              key={variant}
              variant={variant}
              placeholder={variant}
              style={{ width: W.input }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12}>
          <Flex gap={16} align="center">
            <Label>Default</Label>
            <TimePicker
              status="error"
              placeholder="Error"
              style={{ width: W.input }}
            />
            <TimePicker
              status="warning"
              placeholder="Warning"
              style={{ width: W.input }}
            />
          </Flex>
          <Flex gap={16} align="center">
            <Label>Range</Label>
            <RangePicker status="error" />
            <RangePicker status="warning" />
          </Flex>
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex gap={16} align="center">
          <TimePicker
            disabled
            placeholder="Disabled"
            style={{ width: W.input }}
          />
          <RangePicker disabled />
        </Flex>
      </Section>

      <Section title="RangePicker">
        <Flex vertical gap={12}>
          <Flex gap={16} align="center">
            <Label>Basic</Label>
            <RangePicker />
          </Flex>
          <Flex vertical gap={8}>
            {sizes.map((size) => (
              <Flex key={size} align="center" gap={8}>
                <Label>{size}</Label>
                <RangePicker size={size} />
              </Flex>
            ))}
          </Flex>
          <Flex vertical gap={8}>
            {variants.map((variant) => (
              <Flex key={variant} align="center" gap={8}>
                <Label>{variant}</Label>
                <RangePicker
                  variant={variant}
                  placeholder={[variant, variant]}
                />
              </Flex>
            ))}
          </Flex>
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <TimePicker />
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
      <div ref={containerRef} style={{ position: "relative", minHeight: 600 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Flex vertical>
            <div style={{ marginBottom: 260 }}>
              <TimePicker
                open
                getPopupContainer={() => containerRef.current ?? document.body}
                placeholder="TimePicker open"
              />
            </div>
            <RangePicker
              open
              getPopupContainer={() => containerRef.current ?? document.body}
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
    onOpenChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("textbox")
    await userEvent.click(input)
    await expect(args.onOpenChange).toHaveBeenCalled()
    const dropdown = within(document.body)
    const nowBtn = dropdown.getByText("Now")
    await userEvent.click(nowBtn)
    await expect(args.onChange).toHaveBeenCalled()
  },
}

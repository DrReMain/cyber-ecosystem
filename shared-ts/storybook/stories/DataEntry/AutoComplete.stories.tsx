import type { Meta, StoryObj } from "@storybook/react-vite"
import { AutoComplete, Flex } from "antd"
import { useRef, useState } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  callbackArgs,
  disabledArg,
  sizeArg,
  statusArg,
  variantArg,
} from "../argTypes"
import { Label, PageContent, Section, W } from "../helpers"

const meta: Meta<typeof AutoComplete> = {
  title: "Antd/Data Entry/AutoComplete",
  component: AutoComplete,
  parameters: { layout: "padded" },
  args: {
    placeholder: "Try typing 'burn', 'dark', or 'gentle'",
    style: { width: W.cascader },
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
type Story = StoryObj<typeof AutoComplete>

const mockOptions = [
  { value: "burning ashes" },
  { value: "burning legion" },
  { value: "burning crusade" },
  { value: "calm before the storm" },
  { value: "calm waters" },
  { value: "dark knight" },
  { value: "dark matter" },
  { value: "dark souls" },
  { value: "gentle breeze" },
  { value: "gentle rain" },
]

const filterOptions = (text: string) =>
  text
    ? mockOptions.filter((o) => o.value.includes(text.toLowerCase()))
    : mockOptions

function useFilteredOptions() {
  const [options, setOptions] = useState(mockOptions)
  const handleChange = (value: string) => setOptions(filterOptions(value))
  return { options, handleChange }
}

const sizes = ["large", "medium", "small"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["warning", "error"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const sizeDemos = sizes.map(() => useFilteredOptions())
    const variantDemos = variants.map(() => useFilteredOptions())
    const statusDemos = statuses.map(() => useFilteredOptions())

    return (
      <Flex vertical gap={24}>
        <Section title="Sizes">
          <Flex vertical gap={12}>
            {sizes.map((size, i) => (
              <Flex key={size} align="center" gap={12}>
                <Label>{size}</Label>
                <AutoComplete
                  options={sizeDemos[i].options}
                  onChange={sizeDemos[i].handleChange}
                  placeholder={`Size: ${size}`}
                  size={size}
                  style={{ width: W.cascader }}
                />
              </Flex>
            ))}
          </Flex>
        </Section>

        <Section title="Variants">
          <Flex vertical gap={12}>
            {variants.map((variant, i) => (
              <Flex key={variant} align="center" gap={12}>
                <Label>{variant}</Label>
                <AutoComplete
                  options={variantDemos[i].options}
                  onChange={variantDemos[i].handleChange}
                  placeholder={`Variant: ${variant}`}
                  variant={variant}
                  style={{ width: W.cascader }}
                />
              </Flex>
            ))}
          </Flex>
        </Section>

        <Section title="Status">
          <Flex vertical gap={12}>
            {statuses.map((status, i) => (
              <Flex key={status} align="center" gap={12}>
                <Label>{status}</Label>
                <AutoComplete
                  options={statusDemos[i].options}
                  onChange={statusDemos[i].handleChange}
                  placeholder={`Status: ${status}`}
                  status={status}
                  style={{ width: W.cascader }}
                />
              </Flex>
            ))}
          </Flex>
        </Section>

        <Section title="Disabled">
          <AutoComplete
            options={mockOptions}
            placeholder="Disabled"
            disabled
            style={{ width: W.cascader }}
          />
        </Section>
      </Flex>
    )
  },
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
          <AutoComplete
            open
            options={mockOptions.filter((o) => o.value.includes("burn"))}
            value="burn"
            placeholder="Static open dropdown"
            style={{ width: W.cascader }}
            getPopupContainer={() => containerRef.current!}
          />
        </div>
      </div>
    )
  },
}

export const Interactive: Story = {
  args: {
    options: mockOptions,
    onChange: fn(),
    placeholder: "Type 'burn' to see options",
    style: { width: W.cascader },
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const input = canvas.getByRole("combobox")
    await input.focus()
    await userEvent.type(input, "bur")
    await expect(args.onChange).toHaveBeenCalled()
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Select } from "antd"
import { useRef } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { disabledArg, sizeArg, statusArg, variantArg } from "../argTypes"
import { Label, PageContent, Section, W } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Select> = {
  title: "Antd/Data Entry/Select",
  component: Select,
  parameters: { layout: "padded" },
  args: {
    style: { width: W.select },
    placeholder: "Select...",
  },
  argTypes: {
    mode: {
      control: "select",
      options: ["multiple", "tags"],
    },
    showSearch: { control: "boolean" },
    disabled: disabledArg,
    status: statusArg,
    size: sizeArg(["large", "medium", "small"]),
    variant: variantArg(),
    loading: { control: "boolean" },
    onChange: { action: "changed" },
    onOpenChange: { action: "openChanged" },
  },
}

export default meta
type Story = StoryObj<typeof Select>

const basicOptions = [
  { value: "apple", label: "Apple" },
  { value: "banana", label: "Banana" },
  { value: "cherry", label: "Cherry" },
  { value: "durian", label: "Durian" },
]

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
            <Select
              key={size}
              size={size}
              options={basicOptions}
              defaultValue="apple"
              style={{ width: W.select }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex vertical gap={12}>
          {variants.map((variant) => (
            <Select
              key={variant}
              variant={variant}
              options={basicOptions}
              defaultValue="apple"
              style={{ width: W.select }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12}>
          {statuses.map((status) => (
            <Select
              key={status}
              status={status}
              options={basicOptions}
              defaultValue="apple"
              style={{ width: W.select }}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex gap={12} wrap>
          <Select
            disabled
            options={basicOptions}
            defaultValue="apple"
            style={{ width: W.select }}
          />
          <Select
            disabled
            options={basicOptions}
            placeholder="Disabled empty"
            style={{ width: W.select }}
          />
        </Flex>
      </Section>

      <Section title="Mode: Multiple">
        <Select
          mode="multiple"
          options={basicOptions}
          defaultValue={["apple", "banana"]}
          style={{ width: "100%" }}
          placeholder="Select fruits"
          allowClear
        />
      </Section>

      <Section title="Mode: Tags">
        <Select
          mode="tags"
          options={basicOptions}
          defaultValue={["apple", "custom-tag"]}
          style={{ width: "100%" }}
          placeholder="Type to add tags"
          allowClear
        />
      </Section>

      <Section title="Loading">
        <Select
          loading
          options={[]}
          style={{ width: W.select }}
          placeholder="Loading..."
        />
      </Section>

      <Section title="States">
        <PseudoStates>
          <Select
            options={basicOptions}
            defaultValue="apple"
            style={{ width: W.select }}
          />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { options: basicOptions },
}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    const groupedOptions = [
      {
        label: "Manager",
        options: [
          { value: "jack", label: "Jack" },
          { value: "lucy", label: "Lucy" },
        ],
      },
      {
        label: "Engineer",
        options: [
          { value: "tom", label: "Tom" },
          { value: "jerry", label: "Jerry" },
        ],
      },
    ]
    return (
      <div ref={containerRef} style={{ position: "relative", minHeight: 500 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Flex vertical>
            <div style={{ marginBottom: 220 }}>
              <Select
                open
                options={basicOptions}
                defaultValue="apple"
                style={{ width: W.select }}
                getPopupContainer={() => containerRef.current ?? document.body}
              />
            </div>
            <Label>With grouped options</Label>
            <Select
              open
              options={groupedOptions}
              defaultValue="lucy"
              style={{ width: W.select }}
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
    options: basicOptions,
    defaultValue: "apple",
    onChange: fn(),
    onOpenChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const trigger = canvas.getByRole("combobox")
    await userEvent.click(trigger)
    await expect(args.onOpenChange).toHaveBeenCalled()
    const option = await within(document.body).findByRole("option", {
      name: "Banana",
    })
    await userEvent.click(option)
    await expect(args.onChange).toHaveBeenCalled()
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Cascader, Flex } from "antd"
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

const options = [
  {
    value: "zhejiang",
    label: "Zhejiang",
    children: [
      {
        value: "hangzhou",
        label: "Hangzhou",
        children: [
          { value: "xihu", label: "West Lake" },
          { value: "xiasha", label: "Xia Sha" },
        ],
      },
      {
        value: "ningbo",
        label: "Ningbo",
        children: [{ value: "jiangbei", label: "Jiang Bei" }],
      },
    ],
  },
  {
    value: "jiangsu",
    label: "Jiangsu",
    children: [
      {
        value: "nanjing",
        label: "Nanjing",
        children: [{ value: "zhonghuamen", label: "Zhong Hua Men" }],
      },
      {
        value: "suzhou",
        label: "Suzhou",
        children: [{ value: "gusu", label: "Gu Su" }],
      },
    ],
  },
]

const meta: Meta<typeof Cascader> = {
  title: "Antd/Data Entry/Cascader",
  component: Cascader,
  parameters: { layout: "padded" },
  args: { options, placeholder: "Please select", style: { width: W.cascader } },
  argTypes: {
    size: sizeArg(),
    variant: variantArg(),
    status: statusArg,
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof Cascader>

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
            <Flex key={size} align="center" gap={12}>
              <Label>{size}</Label>
              <Cascader
                options={options}
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
          {variants.map((variant) => (
            <Flex key={variant} align="center" gap={12}>
              <Label>{variant}</Label>
              <Cascader
                options={options}
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
          {statuses.map((status) => (
            <Flex key={status} align="center" gap={12}>
              <Label>{status}</Label>
              <Cascader
                options={options}
                placeholder={`Status: ${status}`}
                status={status}
                style={{ width: W.cascader }}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Cascader
          options={options}
          placeholder="Disabled"
          disabled
          style={{ width: W.cascader }}
        />
      </Section>

      <Section title="Multiple">
        <Flex vertical gap={12}>
          <Flex align="center" gap={12}>
            <Label>multiple</Label>
            <Cascader
              options={options}
              placeholder="Multiple select"
              multiple
              maxTagCount={2}
              style={{ width: W.cascader }}
            />
          </Flex>
          <Flex align="center" gap={12}>
            <Label>SHOW_CHILD</Label>
            <Cascader
              options={options}
              placeholder="Show child strategy"
              multiple
              showCheckedStrategy={Cascader.SHOW_CHILD}
              style={{ width: W.cascader }}
            />
          </Flex>
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Cascader options={options} placeholder="Select" />
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
      <div ref={containerRef} style={{ position: "relative", minHeight: 400 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Cascader
            open
            options={options}
            showSearch={false}
            placeholder="Static open cascader"
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
    options,
    onChange: fn(),
    placeholder: "Click to open",
    style: { width: W.cascader },
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const combobox = canvas.getByRole("combobox")
    await combobox.click()
    const dropdown = within(document.body)
    await expect(dropdown.getByText("Zhejiang")).toBeInTheDocument()
  },
}

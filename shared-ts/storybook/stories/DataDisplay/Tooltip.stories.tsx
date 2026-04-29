import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Space, Tooltip } from "antd"
import { useRef } from "react"
import { expect, fn, within } from "storybook/test"
import { PageContent, Section } from "../helpers"

const meta: Meta<typeof Tooltip> = {
  title: "Antd/Data Display/Tooltip",
  component: Tooltip,
  parameters: { layout: "padded" },
  args: {
    title: "prompt text",
    placement: "top",
    trigger: "hover",
  },
  argTypes: {
    placement: {
      control: "select",
      options: [
        "top",
        "topLeft",
        "topRight",
        "right",
        "rightTop",
        "rightBottom",
        "bottom",
        "bottomLeft",
        "bottomRight",
        "left",
        "leftTop",
        "leftBottom",
      ],
    },
    trigger: {
      control: "radio",
      options: ["hover", "click", "focus"],
    },
    color: {
      control: "select",
      options: [
        "default",
        "pink",
        "red",
        "yellow",
        "orange",
        "cyan",
        "green",
        "blue",
        "purple",
        "geekblue",
        "magenta",
        "volcano",
        "gold",
        "lime",
      ],
    },
    arrow: {
      control: "radio",
      options: ["show", "hide"],
    },
  },
}

export default meta
type Story = StoryObj<typeof Tooltip>

const presetColors = [
  "pink",
  "red",
  "yellow",
  "orange",
  "cyan",
  "green",
  "blue",
  "purple",
  "geekblue",
  "magenta",
  "volcano",
  "gold",
  "lime",
]

const customColors = ["#f50", "#2db7f5", "#87d068", "#108ee9"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Colors">
        <Flex vertical gap={8}>
          <Space wrap>
            {presetColors.map((color) => (
              <Tooltip title={`Color: ${color}`} color={color} key={color}>
                <Button>{color}</Button>
              </Tooltip>
            ))}
          </Space>
          <Space wrap>
            {customColors.map((color) => (
              <Tooltip title={`Color: ${color}`} color={color} key={color}>
                <Button>{color}</Button>
              </Tooltip>
            ))}
          </Space>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const StaticOpen: Story = {
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div
        ref={containerRef}
        style={{ position: "relative", minHeight: 640, padding: "100px 0" }}
      >
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Flex vertical gap={120} align="center">
            {(["top", "bottom", "left", "right"] as const).map((placement) => (
              <Tooltip
                key={placement}
                title={`Static ${placement} tooltip`}
                placement={placement}
                open
                getPopupContainer={() => containerRef.current!}
              >
                <Button>{placement}</Button>
              </Tooltip>
            ))}
          </Flex>
        </div>
      </div>
    )
  },
}

export const Playground: Story = {
  render: (args) => (
    <Tooltip {...args}>
      <Button>Hover me</Button>
    </Tooltip>
  ),
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    title: "Tooltip text",
    trigger: "click",
    afterOpenChange: fn(),
  },
  render: (args) => (
    <Tooltip {...args}>
      <Button>Click me</Button>
    </Tooltip>
  ),
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByText("Click me")
    await button.click()
    const tooltip = await within(document.body).findByRole("tooltip")
    await expect(tooltip).toBeInTheDocument()
    await expect(tooltip).toHaveTextContent("Tooltip text")
    await expect(args.afterOpenChange).toHaveBeenCalledWith(true)
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Pagination } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Pagination> = {
  title: "Antd/Navigation/Pagination",
  component: Pagination,
  parameters: { layout: "padded" },
  args: {
    defaultCurrent: 1,
    total: 50,
  },
  argTypes: {
    size: sizeArg(["small", "medium", "large"]),
    disabled: disabledArg,
    showQuickJumper: { control: "boolean" },
    showSizeChanger: { control: "boolean" },
    showLessItems: { control: "boolean" },
    simple: { control: "boolean" },
    align: {
      control: "radio",
      options: ["start", "center", "end"],
    },
    onChange: { action: "changed" },
    onShowSizeChange: { action: "sizeChanged" },
  },
}

export default meta
type Story = StoryObj<typeof Pagination>

const sizes = ["small", "medium", "large"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      {/* 1. Basic */}
      <Section title="Basic">
        <Pagination defaultCurrent={1} total={50} />
      </Section>

      {/* 2. Sizes */}
      <Section title="Sizes">
        <Flex vertical gap={12}>
          {sizes.map((s) => (
            <Flex key={s} vertical gap={4}>
              <Label>{s}</Label>
              <Pagination size={s} defaultCurrent={1} total={50} />
            </Flex>
          ))}
        </Flex>
      </Section>

      {/* 3. Disabled */}
      <Section title="Disabled">
        <Pagination disabled defaultCurrent={3} total={50} />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    defaultCurrent: 1,
    total: 50,
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const page3 = canvas.getByTitle("3")
    await page3.click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

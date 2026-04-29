import type { Meta, StoryObj } from "@storybook/react-vite"
import { Divider, Flex, Typography } from "antd"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Divider> = {
  title: "Antd/Layout/Divider",
  component: Divider,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    variant: {
      control: "select",
      options: ["solid", "dashed", "dotted"],
    },
    orientation: {
      control: "radio",
      options: ["horizontal", "vertical"],
    },
    titlePlacement: {
      control: "radio",
      options: ["start", "center", "end"],
    },
    size: sizeArg(["small", "medium", "large"]),
    plain: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Divider>

const variants = ["solid", "dashed", "dotted"] as const
const sizes = ["small", "medium", "large"] as const
const placements = ["start", "center", "end"] as const

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Variants">
        <Flex vertical gap={8}>
          {variants.map((v) => (
            <Flex key={v} vertical gap={4}>
              <Label>{v}</Label>
              <Divider variant={v} />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Title Placement">
        <Flex vertical gap={8}>
          {placements.map((p) => (
            <Divider key={p} titlePlacement={p}>
              {p}
            </Divider>
          ))}
        </Flex>
      </Section>

      <Section title="Sizes">
        <Flex vertical gap={8}>
          {sizes.map((s) => (
            <Flex key={s} vertical gap={4}>
              <Label>{s}</Label>
              <Divider size={s} />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Vertical">
        <div>
          Text
          <Divider orientation="vertical" />
          <Typography.Link>Link</Typography.Link>
          <Divider orientation="vertical" />
          <Typography.Link>Link</Typography.Link>
        </div>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

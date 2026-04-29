import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Card, Flex, Space } from "antd"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Space> = {
  title: "Antd/Layout/Space",
  component: Space,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    size: sizeArg(["small", "middle", "large"]),
    align: {
      control: "radio",
      options: ["start", "center", "end", "baseline"],
    },
    wrap: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Space>

const sizes = ["small", "middle", "large"] as const
const aligns = ["start", "center", "end", "baseline"] as const

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex vertical gap={8}>
          {sizes.map((s) => (
            <Flex key={s} vertical gap={4}>
              <Label>{s}</Label>
              <Space size={s}>
                <Button>Btn 1</Button>
                <Button>Btn 2</Button>
                <Button>Btn 3</Button>
              </Space>
            </Flex>
          ))}
          <Flex vertical gap={4}>
            <Label>number (24px)</Label>
            <Space size={24}>
              <Button>Btn 1</Button>
              <Button>Btn 2</Button>
              <Button>Btn 3</Button>
            </Space>
          </Flex>
        </Flex>
      </Section>

      <Section title="Vertical">
        <Space orientation="vertical" style={{ width: 200 }}>
          <Button color="primary" variant="solid" block>
            Button 1
          </Button>
          <Button block>Button 2</Button>
          <Button variant="dashed" block>
            Button 3
          </Button>
        </Space>
      </Section>

      <Section title="Align">
        <Flex vertical gap={8}>
          {aligns.map((a) => (
            <Flex key={a} vertical gap={4}>
              <Label>{a}</Label>
              <Space align={a}>
                <Card size="small">Card content</Card>
                <Button>Button</Button>
                {a === "baseline" && (
                  <>
                    <span style={{ fontSize: 24 }}>Large</span>
                    <span style={{ fontSize: 12 }}>Small</span>
                  </>
                )}
              </Space>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Wrap">
        <Space wrap style={{ maxWidth: 400 }}>
          {Array.from({ length: 12 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static demo items
            <Button key={i}>Button {i + 1}</Button>
          ))}
        </Space>
      </Section>

      <Section title="Compact">
        <Flex vertical gap={8}>
          <Flex vertical gap={4}>
            <Label>horizontal</Label>
            <Space.Compact>
              <Button>Btn 1</Button>
              <Button>Btn 2</Button>
              <Button>Btn 3</Button>
            </Space.Compact>
          </Flex>
          <Flex vertical gap={4}>
            <Label>vertical</Label>
            <Space.Compact orientation="vertical">
              <Button>Btn 1</Button>
              <Button>Btn 2</Button>
              <Button>Btn 3</Button>
            </Space.Compact>
          </Flex>
          <Flex vertical gap={4}>
            <Label>block</Label>
            <Space.Compact block>
              <Button>Btn 1</Button>
              <Button>Btn 2</Button>
              <Button>Btn 3</Button>
            </Space.Compact>
          </Flex>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Space {...args}>
      <Button color="primary" variant="solid">
        Primary
      </Button>
      <Button variant="outlined">Default</Button>
      <Button variant="dashed">Dashed</Button>
    </Space>
  ),
}

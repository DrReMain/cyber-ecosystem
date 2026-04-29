import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Watermark } from "antd"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Watermark> = {
  title: "Antd/Feedback/Watermark",
  component: Watermark,
  parameters: { layout: "padded" },
  args: {
    content: "Ant Design",
    gap: [100, 100],
    rotate: -22,
    font: { fontSize: 16, color: "rgba(0, 0, 0, 0.15)" },
  },
  argTypes: {
    content: { control: "text" },
    rotate: { control: { type: "range", min: -180, max: 180, step: 1 } },
    gap: { control: "object" },
    font: { control: "object" },
  },
}

export default meta
type Story = StoryObj<typeof Watermark>

const placeholder = (height = 300) => <div style={{ height }} />

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic Text">
        <Watermark content="Ant Design">{placeholder()}</Watermark>
      </Section>

      <Section title="Multi-Line">
        <Watermark content={["Ant Design", "Happy Working"]}>
          {placeholder()}
        </Watermark>
      </Section>

      <Section title="With Image">
        <Watermark
          image="https://mdn.alipayobjects.com/huamei_7oahex/afts/img/A*lk8RQ1YX5c8AAAAAAAAAAAAADoJFAQ/original"
          width={130}
          height={40}
        >
          {placeholder()}
        </Watermark>
      </Section>

      <Section title="Custom Gap">
        <Flex gap={16}>
          <div style={{ flex: 1 }}>
            <Label>gap: [100, 100] (default)</Label>
            <Watermark content="Default Gap" gap={[100, 100]}>
              {placeholder(200)}
            </Watermark>
          </div>
          <div style={{ flex: 1 }}>
            <Label>gap: [200, 200]</Label>
            <Watermark content="Wide Gap" gap={[200, 200]}>
              {placeholder(200)}
            </Watermark>
          </div>
          <div style={{ flex: 1 }}>
            <Label>gap: [50, 50]</Label>
            <Watermark content="Tight" gap={[50, 50]}>
              {placeholder(200)}
            </Watermark>
          </div>
        </Flex>
      </Section>

      <Section title="Custom Rotation">
        <Flex gap={16}>
          {[-45, -22, 0, 22, 45].map((rotate) => (
            <div key={rotate} style={{ flex: 1 }}>
              <Label>rotate: {rotate}</Label>
              <Watermark content={`${rotate}deg`} rotate={rotate}>
                {placeholder(200)}
              </Watermark>
            </div>
          ))}
        </Flex>
      </Section>
    </Flex>
  ),
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  render: (args) => <Watermark {...args}>{placeholder()}</Watermark>,
}

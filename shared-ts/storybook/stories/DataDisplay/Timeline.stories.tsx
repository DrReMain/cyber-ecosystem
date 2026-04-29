import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Timeline } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Timeline> = {
  title: "Antd/Data Display/Timeline",
  component: Timeline,
  parameters: { layout: "padded" },
  args: {
    mode: "start",
    reverse: false,
  },
  argTypes: {
    mode: {
      control: "radio",
      options: ["start", "alternate", "end"],
    },
    reverse: { control: "boolean" },
    pending: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof Timeline>

const basicItems = [
  { key: "1", content: "Create a services site 2015-09-01" },
  { key: "2", content: "Solve initial network problems 2015-09-01" },
  { key: "3", content: "Technical testing 2015-09-01" },
  { key: "4", content: "Network problems being solved 2015-09-01" },
]

const presetColors = ["blue", "red", "green", "gray"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Preset Colors">
        <Timeline
          items={presetColors.map((color) => ({
            key: color,
            content: `Item with color: ${color}`,
            color,
          }))}
        />
      </Section>

      <Section title="Custom Colors (Hex)">
        <Timeline
          items={[
            { key: "1", content: "Custom color #f50", color: "#f50" },
            { key: "2", content: "Custom color #2db7f5", color: "#2db7f5" },
            { key: "3", content: "Custom color #87d068", color: "#87d068" },
            { key: "4", content: "Custom color #108ee9", color: "#108ee9" },
          ]}
        />
      </Section>

      <Section title="Custom Icons (icon)">
        <Timeline
          items={[
            {
              key: "1",
              content: "Create a services site",
              icon: <CheckCircleOutlined />,
            },
            {
              key: "2",
              content: "Solve initial network problems",
              icon: <CheckCircleOutlined style={{ color: "green" }} />,
            },
            {
              key: "3",
              content: "Technical testing",
              icon: <SyncOutlined spin style={{ color: "blue" }} />,
            },
            {
              key: "4",
              content: "Network problems being solved",
              icon: <ClockCircleOutlined style={{ color: "red" }} />,
            },
          ]}
        />
      </Section>

      <Section title="Mode: alternate">
        <Timeline
          mode="alternate"
          items={[
            { key: "1", content: "Create a services site 2015-09-01" },
            {
              key: "2",
              content: "Solve initial network problems 2015-09-01",
              color: "green",
            },
            {
              key: "3",
              content: "Technical testing 2015-09-01",
              icon: <ClockCircleOutlined style={{ fontSize: 16 }} />,
            },
            {
              key: "4",
              content: "Network problems being solved 2015-09-01",
              color: "red",
            },
          ]}
        />
      </Section>

      <Section title="Mode: end (right)">
        <Timeline mode="end" items={basicItems} />
      </Section>

      <Section title="Variant: filled">
        <Timeline
          variant="filled"
          items={[
            { key: "1", content: "Create a services site", color: "green" },
            {
              key: "2",
              content: "Solve initial network problems",
              color: "green",
            },
            { key: "3", content: "Technical testing", color: "red" },
            { key: "4", content: "Network problems being solved" },
          ]}
        />
      </Section>

      <Section title="Variant: outlined (default)">
        <Timeline variant="outlined" items={basicItems} />
      </Section>

      <Section title="Orientation: horizontal">
        <Timeline
          orientation="horizontal"
          items={[
            { key: "1", content: "Step 1" },
            { key: "2", content: "Step 2", color: "green" },
            { key: "3", content: "Step 3", color: "red" },
            { key: "4", content: "Step 4" },
          ]}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: {
    items: basicItems,
  },
}

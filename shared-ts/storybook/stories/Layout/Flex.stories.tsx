import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Card, Flex, Tag } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Flex> = {
  title: "Antd/Layout/Flex",
  component: Flex,
  parameters: { layout: "padded" },
  args: { gap: "medium" },
  argTypes: {
    vertical: { control: "boolean" },
    wrap: { control: "boolean" },
    gap: {
      control: "select",
      options: ["small", "medium", "large", 8, 16, 24],
    },
    align: {
      control: "radio",
      options: ["start", "center", "end", "stretch", "baseline"],
    },
    justify: {
      control: "radio",
      options: [
        "start",
        "center",
        "end",
        "space-around",
        "space-between",
        "space-evenly",
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof Flex>

const boxStyle: React.CSSProperties = {
  width: 80,
  height: 80,
  borderRadius: 8,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  fontSize: 14,
  fontWeight: 500,
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Flex gap="medium">
          <Card size="small">1</Card>
          <Card size="small">2</Card>
          <Card size="small">3</Card>
          <Card size="small">4</Card>
        </Flex>
      </Section>
      <Section title="Gap: small">
        <Flex gap="small">
          <Button type="primary">Btn 1</Button>
          <Button>Btn 2</Button>
          <Button>Btn 3</Button>
        </Flex>
      </Section>
      <Section title="Gap: medium">
        <Flex gap="medium">
          <Button type="primary">Btn 1</Button>
          <Button>Btn 2</Button>
          <Button>Btn 3</Button>
        </Flex>
      </Section>
      <Section title="Gap: large">
        <Flex gap="large">
          <Button type="primary">Btn 1</Button>
          <Button>Btn 2</Button>
          <Button>Btn 3</Button>
        </Flex>
      </Section>
      <Section title="Gap: 24 (number)">
        <Flex gap={24}>
          <Tag color="blue">Tag 1</Tag>
          <Tag color="green">Tag 2</Tag>
          <Tag color="orange">Tag 3</Tag>
        </Flex>
      </Section>
      <Section title="Vertical">
        <Flex vertical gap="medium">
          <Card size="small">1</Card>
          <Card size="small">2</Card>
          <Card size="small">3</Card>
          <Card size="small">4</Card>
        </Flex>
      </Section>
      <Section title="Align: start">
        <Flex gap="small" align="start">
          <div style={boxStyle}>
            <Card size="small">1</Card>
          </div>
          <div style={{ ...boxStyle, height: 120 }}>
            <Card size="small">2</Card>
          </div>
          <div style={boxStyle}>
            <Card size="small">3</Card>
          </div>
        </Flex>
      </Section>
      <Section title="Align: center">
        <Flex gap="small" align="center">
          <Card size="small" style={{ height: 80 }}>
            1
          </Card>
          <Card size="small" style={{ height: 120 }}>
            2
          </Card>
          <Card size="small" style={{ height: 80 }}>
            3
          </Card>
        </Flex>
      </Section>
      <Section title="Align: end">
        <Flex gap="small" align="end">
          <Card size="small" style={{ height: 80 }}>
            1
          </Card>
          <Card size="small" style={{ height: 120 }}>
            2
          </Card>
          <Card size="small" style={{ height: 80 }}>
            3
          </Card>
        </Flex>
      </Section>
      <Section title="Align: stretch">
        <Flex
          gap="small"
          align="stretch"
          style={{
            background: "var(--ant-color-bg-layout)",
            padding: 8,
            borderRadius: 8,
          }}
        >
          <Card size="small" style={{ height: "auto" }}>
            1
          </Card>
          <Card size="small" style={{ height: 120 }}>
            2
          </Card>
          <Card size="small" style={{ height: "auto" }}>
            3
          </Card>
        </Flex>
      </Section>
      <Section title="Justify: center">
        <Flex gap="small" justify="center">
          <Button type="primary">1</Button>
          <Button>2</Button>
          <Button>3</Button>
        </Flex>
      </Section>
      <Section title="Justify: space-between">
        <Flex gap="small" justify="space-between">
          <Button type="primary">1</Button>
          <Button>2</Button>
          <Button>3</Button>
        </Flex>
      </Section>
      <Section title="No wrap (default)">
        <Flex gap="small" style={{ width: 300, overflow: "hidden" }}>
          <Card size="small">1</Card>
          <Card size="small">2</Card>
          <Card size="small">3</Card>
          <Card size="small">4</Card>
          <Card size="small">5</Card>
        </Flex>
      </Section>
      <Section title="With wrap">
        <Flex gap="small" wrap style={{ width: 300 }}>
          <Card size="small">1</Card>
          <Card size="small">2</Card>
          <Card size="small">3</Card>
          <Card size="small">4</Card>
          <Card size="small">5</Card>
        </Flex>
      </Section>
      <Section title="Nested (row containing columns)">
        <Flex gap="medium">
          <Flex vertical gap="small" style={{ flex: 1 }}>
            <Card size="small">Column A - Row 1</Card>
            <Card size="small">Column A - Row 2</Card>
          </Flex>
          <Flex vertical gap="small" style={{ flex: 2 }}>
            <Card size="small">Column B - Row 1 (wider)</Card>
            <Card size="small">Column B - Row 2 (wider)</Card>
          </Flex>
          <Flex vertical gap="small" style={{ flex: 1 }}>
            <Card size="small">Column C - Row 1</Card>
            <Card size="small">Column C - Row 2</Card>
          </Flex>
        </Flex>
      </Section>
      <Section title="Flex prop">
        <Flex gap="small">
          <Card size="small" style={{ flex: 1 }}>
            flex: 1
          </Card>
          <Card size="small" style={{ flex: 2 }}>
            flex: 2
          </Card>
          <Card size="small" style={{ flex: 1 }}>
            flex: 1
          </Card>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Flex {...args}>
      <Button type="primary">Btn 1</Button>
      <Button>Btn 2</Button>
      <Button>Btn 3</Button>
    </Flex>
  ),
}

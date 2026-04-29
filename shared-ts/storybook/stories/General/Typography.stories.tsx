import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Typography } from "antd"
import { expect, fn, within } from "storybook/test"
import { Section } from "../helpers"

const { Title, Text, Paragraph } = Typography

const textTypes = ["secondary", "success", "warning", "danger"] as const

const meta: Meta<typeof Typography> = {
  title: "Antd/General/Typography",
  component: Typography,
  parameters: { layout: "padded" },
  argTypes: {},
}

export default meta
type Story = StoryObj<typeof Typography>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Title Levels">
        <Flex vertical gap={8}>
          <Title level={1}>h1. Title Level 1</Title>
          <Title level={2}>h2. Title Level 2</Title>
          <Title level={3}>h3. Title Level 3</Title>
          <Title level={4}>h4. Title Level 4</Title>
          <Title level={5}>h5. Title Level 5</Title>
        </Flex>
      </Section>

      <Section title="Title Types">
        <Flex vertical gap={4}>
          <Title level={4}>Default</Title>
          {textTypes.map((t) => (
            <Title key={t} level={4} type={t}>
              {t}
            </Title>
          ))}
        </Flex>
      </Section>

      <Section title="Title Styles">
        <Flex vertical gap={4}>
          <Title level={4} mark>
            Marked title
          </Title>
          <Title level={4} code>
            Code title
          </Title>
          <Title level={4} underline>
            Underlined title
          </Title>
          <Title level={4} delete>
            Deleted title
          </Title>
          <Title level={4} underline>
            Strong + underline
          </Title>
          <Title level={4} keyboard>
            Keyboard title
          </Title>
        </Flex>
      </Section>

      <Section title="Text Types">
        <Flex vertical gap={4}>
          <Text>Default text</Text>
          {textTypes.map((t) => (
            <Text key={t} type={t}>
              {t} text
            </Text>
          ))}
          <Text disabled>Disabled text</Text>
        </Flex>
      </Section>

      <Section title="Text Styles">
        <Flex vertical gap={4}>
          <Text strong>Bold (strong)</Text>
          <Text italic>Italic</Text>
          <Text underline>Underlined</Text>
          <Text delete>Deleted (strikethrough)</Text>
          <Text mark>Marked (highlight)</Text>
          <Text code>Code (monospace)</Text>
          <Text keyboard>Keyboard (Ctrl+C)</Text>
        </Flex>
      </Section>

      <Section title="Combinations">
        <Flex vertical gap={4}>
          <Text strong mark>
            Bold and marked
          </Text>
          <Text code delete>
            Code and deleted
          </Text>
          <Text strong underline type="success">
            Bold, underlined, success
          </Text>
          <Text italic type="warning">
            Italic warning
          </Text>
          <Text strong italic underline>
            Bold italic underline
          </Text>
        </Flex>
      </Section>

      <Section title="Paragraph Types">
        <Flex vertical gap={4}>
          <Paragraph>Default paragraph with some body text content.</Paragraph>
          {textTypes.map((t) => (
            <Paragraph key={t} type={t}>
              {t} paragraph
            </Paragraph>
          ))}
          <Paragraph disabled>Disabled paragraph</Paragraph>
        </Flex>
      </Section>

      <Section title="Paragraph Styles">
        <Flex vertical gap={4}>
          <Paragraph strong>Bold paragraph</Paragraph>
          <Paragraph italic>Italic paragraph</Paragraph>
          <Paragraph underline>Underlined paragraph</Paragraph>
          <Paragraph mark>Marked paragraph</Paragraph>
          <Paragraph code>Code paragraph</Paragraph>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: {
    children: "Default text",
  },
  render: (args) => <Text {...args} />,
}

export const Interactive: Story = {
  args: {
    children: "Click me",
    onClick: fn(),
    style: { cursor: "pointer", color: "#1677ff" },
  },
  render: (args) => <Text {...args} />,
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.getByText("Click me").click()
    await expect(args.onClick).toHaveBeenCalledTimes(1)
  },
}

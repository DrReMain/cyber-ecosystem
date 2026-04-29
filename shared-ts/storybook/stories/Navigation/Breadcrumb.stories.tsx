import { HomeOutlined, UserOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Breadcrumb, Flex } from "antd"
import { Section } from "../helpers"

const basicItems = [
  { title: "Home" },
  { title: "Application Center" },
  { title: "Application List" },
  { title: "An Application" },
]

const meta: Meta<typeof Breadcrumb> = {
  title: "Antd/Navigation/Breadcrumb",
  component: Breadcrumb,
  parameters: { layout: "padded" },
  args: {
    items: basicItems,
  },
  argTypes: {
    separator: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof Breadcrumb>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Breadcrumb items={basicItems} />
      </Section>
      <Section title="Custom Separator">
        <Flex vertical gap={12}>
          <Breadcrumb separator=">" items={basicItems} />
          <Breadcrumb separator="-" items={basicItems} />
          <Breadcrumb separator=">>>" items={basicItems} />
        </Flex>
      </Section>
      <Section title="With Icons">
        <Breadcrumb
          items={[
            {
              title: (
                <>
                  <HomeOutlined /> Home
                </>
              ),
            },
            {
              title: (
                <>
                  <UserOutlined /> User Center
                </>
              ),
            },
            { title: "Profile" },
          ]}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

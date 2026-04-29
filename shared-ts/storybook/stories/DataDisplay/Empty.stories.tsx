import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Empty, Flex, Space } from "antd"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Empty> = {
  title: "Antd/Data Display/Empty",
  component: Empty,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    description: { control: "text" },
    image: {
      control: "select",
      options: ["default", "simple"],
      mapping: {
        default: undefined,
        simple: Empty.PRESENTED_IMAGE_SIMPLE,
      },
    },
  },
}

export default meta
type Story = StoryObj<typeof Empty>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Default">
        <Empty />
      </Section>

      <Section title="Simple">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </Section>

      <Section title="With Children (Action)">
        <Space orientation="vertical" size={16}>
          <div>
            <Label>Single action</Label>
            <Empty description="No items yet">
              <Button type="primary">Create New</Button>
            </Empty>
          </div>
          <div>
            <Label>Multiple actions</Label>
            <Empty description="Something went wrong">
              <Space>
                <Button type="primary">Retry</Button>
                <Button>Go Back</Button>
              </Space>
            </Empty>
          </div>
        </Space>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

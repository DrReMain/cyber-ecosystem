import { UserOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Avatar, Flex, Space } from "antd"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

type AvatarProps = React.ComponentProps<typeof Avatar>

const meta: Meta<typeof Avatar> = {
  title: "Antd/Data Display/Avatar",
  component: Avatar,
  parameters: { layout: "padded" },
  args: { icon: <UserOutlined /> },
  argTypes: {
    shape: {
      control: "radio",
      options: ["circle", "square"],
    },
    size: sizeArg(["small", "medium", "large"]),
    gap: { control: { type: "number", min: 0, max: 10 } },
  },
}

export default meta
type Story = StoryObj<typeof Avatar>

const avatarSrc = "https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"

const shapes: Array<AvatarProps["shape"]> = ["circle", "square"]
const sizes: Array<AvatarProps["size"]> = [20, 32, 40, 64, 80]
const presetSizes: Array<AvatarProps["size"]> = ["small", "medium", "large"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Shapes" description="Circle (default) vs Square">
        <Space size={24}>
          {shapes.map((shape) => (
            <Space key={shape} orientation="vertical" align="center">
              <Label>{shape}</Label>
              <Space size={12}>
                <Avatar shape={shape} icon={<UserOutlined />} />
                <Avatar shape={shape} src={avatarSrc} />
                <Avatar shape={shape} style={{ backgroundColor: "#f56a00" }}>
                  AB
                </Avatar>
              </Space>
            </Space>
          ))}
        </Space>
      </Section>

      <Section title="Preset Sizes" description="small / medium / large">
        <Space size={16} align="center">
          {presetSizes.map((size) => (
            <Space key={String(size)} orientation="vertical" align="center">
              <Label>{String(size)}</Label>
              <Avatar size={size} icon={<UserOutlined />} />
            </Space>
          ))}
        </Space>
      </Section>

      <Section title="Custom Number Sizes" description="Numeric pixel values">
        <Space size={16} align="center">
          {sizes.map((size) => (
            <Space key={String(size)} orientation="vertical" align="center">
              <Label>{String(size)}px</Label>
              <Avatar size={size} icon={<UserOutlined />} />
            </Space>
          ))}
        </Space>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

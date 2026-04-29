import {
  AlignLeftOutlined,
  BoldOutlined,
  ItalicOutlined,
  UnderlineOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Affix,
  Button,
  Flex,
  Menu,
  Space,
  Tooltip,
  Typography,
  theme,
} from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Affix> = {
  title: "Antd/Other/Affix",
  component: Affix,
  parameters: { layout: "fullscreen" },
  args: { offsetTop: 0 },
  argTypes: {
    offsetTop: {
      control: { type: "number", min: 0, max: 300 },
      description: "Distance from top of viewport to affix.",
    },
    offsetBottom: {
      control: { type: "number", min: 0, max: 300 },
      description: "Distance from bottom of viewport to affix.",
    },
    target: {
      control: false,
      description: "Function returning a scrollable container element.",
    },
    onChange: {
      action: "onChange",
      description: "Called when affix state changes.",
    },
  },
}

export default meta
type Story = StoryObj<typeof Affix>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => <AffixGallery />,
}

function AffixGallery() {
  const { token } = theme.useToken()

  return (
    <Flex vertical gap={24}>
      <div
        style={{
          height: 2000,
          padding: "60px 24px",
        }}
      >
        <Flex vertical gap={48}>
          <Section
            title="Basic"
            description="Scroll down to see the affix button stick to the top."
          >
            <Affix offsetTop={0}>
              <Button type="primary">Affixed to Top</Button>
            </Affix>
          </Section>

          <Section
            title="Top with Offset"
            description="Affix with 60px offset from top."
          >
            <Affix offsetTop={60}>
              <Button type="primary">Affixed at 60px from Top</Button>
            </Affix>
          </Section>

          <Section
            title="Top with Large Offset"
            description="Affix with 120px offset from top."
          >
            <Affix offsetTop={120}>
              <Button>Affixed at 120px from Top</Button>
            </Affix>
          </Section>

          <Section
            title="Bottom Affix"
            description="Scroll to the bottom to see the affix button."
          >
            <Affix offsetBottom={0}>
              <Button type="primary">Affixed to Bottom</Button>
            </Affix>
          </Section>

          <Section
            title="Bottom with Offset"
            description="Affix with 40px offset from bottom."
          >
            <Affix offsetBottom={40}>
              <Button>Affixed at 40px from Bottom</Button>
            </Affix>
          </Section>

          <Section
            title="Wrapping a Toolbar"
            description="Affix wrapping a small toolbar instead of a single Button."
          >
            <Affix offsetTop={0}>
              <Space
                style={{
                  padding: "8px 12px",
                  background: token.colorBgElevated,
                  borderRadius: 6,
                  border: `1px solid ${token.colorBorder}`,
                }}
              >
                <Tooltip title="Bold">
                  <Button type="text" icon={<BoldOutlined />} />
                </Tooltip>
                <Tooltip title="Italic">
                  <Button type="text" icon={<ItalicOutlined />} />
                </Tooltip>
                <Tooltip title="Underline">
                  <Button type="text" icon={<UnderlineOutlined />} />
                </Tooltip>
                <Tooltip title="Align Left">
                  <Button type="text" icon={<AlignLeftOutlined />} />
                </Tooltip>
              </Space>
            </Affix>
          </Section>

          <Section
            title="Affix in Scrollable Container"
            description="An Affix pinned inside a scrollable container using target."
          >
            <div
              id="scrollable-container"
              style={{
                height: 200,
                overflowY: "auto",
                border: `1px solid ${token.colorBorder}`,
                borderRadius: 8,
                padding: 12,
              }}
            >
              <div style={{ height: 600 }}>
                <Affix
                  offsetTop={0}
                  target={() =>
                    document.getElementById(
                      "scrollable-container",
                    ) as HTMLElement
                  }
                >
                  <Menu
                    mode="horizontal"
                    items={[
                      { key: "1", label: "Tab 1" },
                      { key: "2", label: "Tab 2" },
                      { key: "3", label: "Tab 3" },
                    ]}
                    style={{ borderRadius: 6 }}
                  />
                </Affix>
                <Typography.Text
                  type="secondary"
                  style={{ display: "block", marginTop: 16 }}
                >
                  Scroll inside this container to see the Menu affix at the top.
                </Typography.Text>
              </div>
            </div>
          </Section>
        </Flex>
      </div>
    </Flex>
  )
}

export const Playground: Story = {
  render: (args) => (
    <div style={{ padding: "60px 24px", minHeight: 2000 }}>
      <Affix {...args}>
        <Button type="primary">Affixed to Top</Button>
      </Affix>
      <div style={{ height: 1600 }} />
    </div>
  ),
}

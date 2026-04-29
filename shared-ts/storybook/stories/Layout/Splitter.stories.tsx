import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Splitter, theme } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Splitter> = {
  title: "Antd/Layout/Splitter",
  component: Splitter,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    orientation: {
      control: "radio",
      options: ["horizontal", "vertical"],
    },
    lazy: { control: "boolean" },
    onResize: { action: "resized" },
  },
}

export default meta
type Story = StoryObj<typeof Splitter>

const panelStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  height: "100%",
}

function ThemeBorder({
  children,
  style,
}: {
  children: React.ReactNode
  style?: React.CSSProperties
}) {
  const { token } = theme.useToken()
  return (
    <div
      style={{
        border: `1px solid ${token.colorBorder}`,
        borderRadius: token.borderRadiusSM,
        ...style,
      }}
    >
      {children}
    </div>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic (horizontal)">
        <ThemeBorder>
          <Splitter style={{ height: 300 }}>
            <Splitter.Panel defaultSize="40%" min="20%">
              <div style={panelStyle}>Panel 1</div>
            </Splitter.Panel>
            <Splitter.Panel>
              <div style={panelStyle}>Panel 2</div>
            </Splitter.Panel>
          </Splitter>
        </ThemeBorder>
      </Section>

      <Section title="Vertical">
        <ThemeBorder>
          <Splitter orientation="vertical" style={{ height: 300 }}>
            <Splitter.Panel defaultSize="50%">
              <div style={panelStyle}>Top Panel</div>
            </Splitter.Panel>
            <Splitter.Panel>
              <div style={panelStyle}>Bottom Panel</div>
            </Splitter.Panel>
          </Splitter>
        </ThemeBorder>
      </Section>

      <Section title="Collapsible">
        <ThemeBorder>
          <Splitter style={{ height: 300 }}>
            <Splitter.Panel collapsible defaultSize="40%" min="20%">
              <div style={panelStyle}>Collapsible Left</div>
            </Splitter.Panel>
            <Splitter.Panel collapsible>
              <div style={panelStyle}>Collapsible Right</div>
            </Splitter.Panel>
          </Splitter>
        </ThemeBorder>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <ThemeBorder>
      <Splitter {...args} style={{ height: 300 }}>
        <Splitter.Panel defaultSize="40%" min="20%">
          <div style={panelStyle}>Panel 1</div>
        </Splitter.Panel>
        <Splitter.Panel>
          <div style={panelStyle}>Panel 2</div>
        </Splitter.Panel>
      </Splitter>
    </ThemeBorder>
  ),
}

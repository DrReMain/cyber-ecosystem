import { LoadingOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Alert, Button, Card, Flex, Space, Spin } from "antd"
import { useState } from "react"
import { expect, within } from "storybook/test"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof Spin> = {
  title: "Antd/Feedback/Spin",
  component: Spin,
  parameters: { layout: "padded" },
  args: {
    size: "medium",
    spinning: true,
    tip: undefined,
  },
  argTypes: {
    size: {
      control: "radio",
      options: ["small", "medium", "large"],
    },
    spinning: { control: "boolean" },
    tip: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof Spin>

const sizes = ["small", "medium", "large"] as const

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Spin />
      </Section>

      <Section title="Sizes">
        <Space size={24}>
          {sizes.map((size) => (
            <div key={size} style={{ textAlign: "center" }}>
              <Spin size={size} />
              <Label>{size}</Label>
            </div>
          ))}
        </Space>
      </Section>

      <Section title="Custom Indicator">
        <Space size={24}>
          <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
          <Spin indicator={<span style={{ fontSize: 20 }}>⏳</span>} />
        </Space>
      </Section>

      <Section title="With Content Wrapping">
        <Spin>
          <Alert
            type="info"
            title="Alert message title"
            description="Further details about the context of this alert."
          />
        </Spin>
      </Section>
    </Flex>
  ),
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {}

// ── StaticOpen ───────────────────────────────────────────────────

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <>
      <PageContent />
      <Spin fullscreen />
    </>
  ),
}

// ── Interactive story with play function ──────────────────────────

export const Interactive: Story = {
  args: { size: "medium" },
  render: (args) => {
    function SpinDemo() {
      const [spinning, setSpinning] = useState(true)
      return (
        <Flex vertical gap={16}>
          <Button onClick={() => setSpinning((s) => !s)}>
            Toggle Spinning
          </Button>
          <Spin {...args} spinning={spinning}>
            <Card style={{ width: 300 }}>
              <p>Card content that can be toggled.</p>
            </Card>
          </Spin>
        </Flex>
      )
    }
    return <SpinDemo />
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole("button", { name: "Toggle Spinning" })
    await button.click()
    const content = canvas.getByText("Card content that can be toggled.")
    await expect(content).toBeInTheDocument()
  },
}

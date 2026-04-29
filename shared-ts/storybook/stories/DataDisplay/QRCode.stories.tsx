import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, QRCode, Space } from "antd"
import { useState } from "react"
import { expect, within } from "storybook/test"
import { borderedArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof QRCode> = {
  title: "Antd/Data Display/QRCode",
  component: QRCode,
  parameters: { layout: "padded" },
  args: {
    value: "https://ant.design",
    size: 160,
  },
  argTypes: {
    value: { control: "text" },
    size: { control: { type: "number", min: 40, max: 400 } },
    color: { control: "color" },
    bgColor: { control: "color" },
    errorLevel: {
      control: "radio",
      options: ["L", "M", "Q", "H"],
    },
    bordered: borderedArg,
    type: {
      control: "radio",
      options: ["canvas", "svg"],
    },
  },
}

export default meta
type Story = StoryObj<typeof QRCode>

const antIcon =
  "https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg"

// ── Args-driven stories ───────────────────────────────────────

function StatusSection() {
  const [status, setStatus] = useState<
    "active" | "expired" | "loading" | "scanned"
  >("active")
  return (
    <Section
      title="Interactive Status"
      description="Toggle between QR code statuses"
    >
      <Space orientation="vertical" align="center">
        <Space>
          {(["active", "expired", "loading", "scanned"] as const).map((s) => (
            <Button
              key={s}
              size="small"
              type={status === s ? "primary" : "default"}
              onClick={() => setStatus(s)}
            >
              {s}
            </Button>
          ))}
        </Space>
        <QRCode value="https://ant.design" status={status} />
      </Space>
    </Section>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes" description="Different size values">
        <Space size={16} align="end">
          {[80, 120, 160, 200, 260].map((s) => (
            <div key={s} style={{ textAlign: "center" }}>
              <Label>{s}px</Label>
              <QRCode value="https://ant.design" size={s} />
            </div>
          ))}
        </Space>
      </Section>

      <Section title="Error Levels" description="L, M, Q, H">
        <Space size={16}>
          {(["L", "M", "Q", "H"] as const).map((level) => (
            <div key={level} style={{ textAlign: "center" }}>
              <Label>{level}</Label>
              <QRCode value="https://ant.design" errorLevel={level} />
            </div>
          ))}
        </Space>
      </Section>

      <Section title="With Icon" description="Embedded icon in the center">
        <Space size={24} align="end">
          <div>
            <Label>icon default size</Label>
            <QRCode value="https://ant.design" icon={antIcon} size={200} />
          </div>
          <div>
            <Label>iconSize: 60</Label>
            <QRCode
              value="https://ant.design"
              icon={antIcon}
              iconSize={60}
              size={200}
            />
          </div>
          <div>
            <Label>iconSize: {`{w:80, h:40}`}</Label>
            <QRCode
              value="https://ant.design"
              icon={antIcon}
              iconSize={{ width: 80, height: 40 }}
              size={200}
            />
          </div>
        </Space>
      </Section>

      <Section title="Colors" description="Custom fg/bg color combinations">
        <Space size={16}>
          <div style={{ textAlign: "center" }}>
            <Label>Default (black)</Label>
            <QRCode value="https://ant.design" />
          </div>
          <div style={{ textAlign: "center" }}>
            <Label>Blue</Label>
            <QRCode value="https://ant.design" color="#1677ff" />
          </div>
          <div style={{ textAlign: "center" }}>
            <Label>Green on light</Label>
            <QRCode
              value="https://ant.design"
              color="#52c41a"
              bgColor="#f6ffed"
              bordered
            />
          </div>
          <div style={{ textAlign: "center" }}>
            <Label>Dark bg</Label>
            <QRCode value="https://ant.design" color="#fff" bgColor="#222" />
          </div>
        </Space>
      </Section>

      <StatusSection />
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    value: "https://ant.design",
    size: 160,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const expiredBtn = canvas.getByText("expired")
    await expiredBtn.click()
    await expect(
      canvas.getByRole("button", { name: /refresh/i }),
    ).toBeInTheDocument()
  },
  render: () => <StatusSection />,
}

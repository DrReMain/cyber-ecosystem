import { ArrowUpOutlined, LikeOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Card, Col, Flex, Row, Space, Statistic, Tag, Typography } from "antd"
import { expect, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta<typeof Statistic> = {
  title: "Antd/Data Display/Statistic",
  component: Statistic,
  parameters: { layout: "padded" },
  args: {
    title: "Active Users",
    value: 112893,
  },
  argTypes: {
    title: { control: "text" },
    value: { control: "number" },
    precision: { control: { type: "number", min: 0, max: 10 } },
    prefix: { control: "text" },
    suffix: { control: "text" },
    loading: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Statistic>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Prefix" description="Icon or text before the value">
        <Space size={32}>
          <Statistic
            title="Income"
            value={112893}
            prefix={<ArrowUpOutlined />}
          />
          <Statistic title="Feedback" value={1128} prefix={<LikeOutlined />} />
          <Statistic title="Price" value={2.79} prefix="$" precision={2} />
        </Space>
      </Section>

      <Section title="Suffix" description="Icon or text after the value">
        <Space size={32}>
          <Statistic title="Account Balance" value={112893} suffix="Yuan" />
          <Statistic title="Rate" value={5.6} suffix="%" precision={1} />
          <Statistic title="Items" value={42} suffix="units" />
        </Space>
      </Section>

      <Section title="Precision Levels" description="Control decimal places">
        <Space size={32}>
          {[0, 2, 4, 6].map((p) => (
            <Statistic
              key={p}
              title={`Precision ${p}`}
              value={Math.PI}
              precision={p}
            />
          ))}
        </Space>
      </Section>

      <Section title="Loading" description="Skeleton placeholder while loading">
        <Space size={32}>
          <Card style={{ width: 180 }}>
            <Statistic title="Active Users" value={112893} loading />
          </Card>
          <Card style={{ width: 180 }}>
            <Statistic title="Active Users" value={112893} />
          </Card>
        </Space>
      </Section>

      <Section
        title="Dashboard KPI Cards"
        description="Realistic dashboard metric row with trend indicators"
      >
        <Row gutter={16}>
          {[
            {
              title: "Total Revenue",
              value: 93200,
              prefix: "$",
              change: "+12.5%",
              up: true,
            },
            { title: "New Users", value: 2847, change: "+8.2%", up: true },
            { title: "Active Orders", value: 532, change: "-3.1%", up: false },
            {
              title: "Conversion Rate",
              value: 3.6,
              suffix: "%",
              change: "+0.4%",
              up: true,
            },
          ].map((item) => (
            <Col key={item.title} span={6}>
              <Card size="small">
                <Statistic
                  title={item.title}
                  value={item.value}
                  prefix={item.prefix}
                  suffix={item.suffix}
                  precision={item.value % 1 !== 0 ? 1 : 0}
                />
                <Flex align="center" gap={4} style={{ marginTop: 8 }}>
                  <Tag
                    color={item.up ? "success" : "error"}
                    style={{ margin: 0 }}
                  >
                    {item.change}
                  </Tag>
                  <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                    vs last month
                  </Typography.Text>
                </Flex>
              </Card>
            </Col>
          ))}
        </Row>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    title: "Active Users",
    value: 112893,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const title = canvas.getByText("Active Users")
    const value = canvas.getByText("112,893")
    await expect(title).toBeInTheDocument()
    await expect(value).toBeInTheDocument()
  },
}

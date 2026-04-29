import { ClockCircleOutlined, UserOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Avatar, Badge, Flex, Space } from "antd"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Badge> = {
  title: "Antd/Data Display/Badge",
  component: Badge,
  parameters: { layout: "padded" },
  args: {
    count: 5,
    children: <Avatar shape="square" icon={<UserOutlined />} />,
  },
  argTypes: {
    count: { control: "number" },
    overflowCount: { control: "number" },
    size: sizeArg(["small", "medium"]),
    dot: { control: "boolean" },
    showZero: { control: "boolean" },
    offset: { control: false },
    status: {
      control: "radio",
      options: ["success", "processing", "default", "error", "warning"],
    },
    color: {
      control: "select",
      options: [
        "pink",
        "red",
        "yellow",
        "orange",
        "cyan",
        "green",
        "blue",
        "purple",
        "geekblue",
        "magenta",
        "volcano",
        "gold",
        "lime",
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof Badge>

const statuses: Array<React.ComponentProps<typeof Badge>["status"]> = [
  "success",
  "processing",
  "default",
  "error",
  "warning",
]

const badgeSizes: Array<React.ComponentProps<typeof Badge>["size"]> = [
  "small",
  "medium",
]

const customColors = [
  { color: "#f50", label: "Magenta" },
  { color: "#2db7f5", label: "Cyan" },
  { color: "#87d068", label: "Green" },
  { color: "#108ee9", label: "Blue" },
  { color: "#f5222d", label: "Red" },
  { color: "#faad14", label: "Gold" },
]

const ribbonPlacements: Array<"start" | "end"> = ["start", "end"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="Basic Count"
        description="Numeric badges with different values and overflow"
      >
        <Space size={24}>
          <Badge count={5}>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
          <Badge count={0} showZero>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
          <Badge count={99}>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
          <Badge count={100}>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
          <Badge count={1000} overflowCount={999}>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
        </Space>
      </Section>

      <Section title="Dot Indicator" description="Dot badges for each status">
        <Space size={24}>
          <Badge dot>
            <Avatar shape="square" icon={<UserOutlined />} />
          </Badge>
          {statuses.map((status) => (
            <Badge key={status} dot status={status}>
              <Avatar shape="square" icon={<UserOutlined />} />
            </Badge>
          ))}
        </Space>
      </Section>

      <Section
        title="Status Standalone"
        description="Status dot with text label"
      >
        <Flex vertical gap={8}>
          {statuses.map((status) => (
            <Badge
              key={status}
              status={status}
              text={`${status?.charAt(0).toUpperCase()}${status?.slice(1)}`}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Sizes" description="small vs medium badge">
        <Space size={24}>
          {badgeSizes.map((size) => (
            <Space key={size} orientation="vertical" align="center">
              <Label>{size}</Label>
              <Badge count={5} size={size}>
                <Avatar shape="square" icon={<UserOutlined />} />
              </Badge>
            </Space>
          ))}
        </Space>
      </Section>

      <Section
        title="Status x Size Matrix"
        description="Each status rendered at each badge size"
      >
        <Flex vertical gap={8}>
          {badgeSizes.map((size) => (
            <Space key={size} align="center">
              <Label>{size}</Label>
              {statuses.map((status) => (
                <Badge key={status} status={status} size={size} text={status} />
              ))}
            </Space>
          ))}
        </Flex>
      </Section>

      <Section title="Custom Colors" description="Custom dot colors with text">
        <Space size={24}>
          {customColors.map(({ color, label }) => (
            <Badge key={color} color={color} text={label}>
              <Avatar shape="square" icon={<UserOutlined />} />
            </Badge>
          ))}
        </Space>
      </Section>

      <Section
        title="Standalone / Text Count"
        description="Badges without a wrapping element, including text and icon counts"
      >
        <Space size={24} align="center">
          <Badge count={25} />
          <Badge count={<ClockCircleOutlined style={{ color: "#f5222d" }} />} />
          <Badge count={0} showZero color="cyan" />
          <Badge count="new" color="green" />
          <Badge count="hot" color="volcano" />
        </Space>
      </Section>

      <Section
        title="Ribbon"
        description="Badge.Ribbon with color and placement variants"
      >
        <Flex vertical gap={16}>
          <Badge.Ribbon text="Default Ribbon">
            <div
              style={{
                background: "var(--ant-color-bg-layout)",
                padding: 24,
                borderRadius: 8,
              }}
            >
              Default ribbon (placement end)
            </div>
          </Badge.Ribbon>
          {customColors.slice(0, 3).map(({ color, label }) => (
            <Badge.Ribbon key={color} text={label} color={color}>
              <div
                style={{
                  background: "var(--ant-color-bg-layout)",
                  padding: 24,
                  borderRadius: 8,
                }}
              >
                {label} ribbon
              </div>
            </Badge.Ribbon>
          ))}
          {ribbonPlacements.map((placement) => (
            <Badge.Ribbon
              key={placement}
              text={`Placement: ${placement}`}
              placement={placement}
            >
              <div
                style={{
                  background: "var(--ant-color-bg-layout)",
                  padding: 24,
                  borderRadius: 8,
                }}
              >
                Ribbon at {placement}
              </div>
            </Badge.Ribbon>
          ))}
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

import {
  BellOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Badge,
  Button,
  Card,
  Empty,
  Flex,
  List,
  Pagination,
  Space,
  Tag,
  Typography,
} from "antd"
import { expect, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Combinations/NotificationCenter",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

const { Text } = Typography

const notifications = [
  {
    id: 1,
    type: "success",
    title: "Deployment completed",
    description: "Project Alpha v2.3.1 deployed to production successfully.",
    time: "5 min ago",
    read: false,
  },
  {
    id: 2,
    type: "warning",
    title: "High memory usage",
    description: "Server node-3 memory usage exceeded 90% threshold.",
    time: "15 min ago",
    read: false,
  },
  {
    id: 3,
    type: "info",
    title: "New team member",
    description: "Bob Smith has joined the Engineering team.",
    time: "1 hour ago",
    read: true,
  },
  {
    id: 4,
    type: "error",
    title: "Build failed",
    description: "CI pipeline failed on branch feature/login-flow.",
    time: "2 hours ago",
    read: true,
  },
  {
    id: 5,
    type: "info",
    title: "Scheduled maintenance",
    description: "Database maintenance scheduled for Sunday 2:00 AM UTC.",
    time: "4 hours ago",
    read: true,
  },
  {
    id: 6,
    type: "success",
    title: "Backup completed",
    description: "Weekly database backup completed successfully.",
    time: "6 hours ago",
    read: true,
  },
]

const typeConfig: Record<
  string,
  { color: string; icon: React.ReactNode; tag: string }
> = {
  success: { color: "#52c41a", icon: <CheckCircleOutlined />, tag: "success" },
  warning: {
    color: "#faad14",
    icon: <ExclamationCircleOutlined />,
    tag: "warning",
  },
  error: {
    color: "#ff4d4f",
    icon: <ExclamationCircleOutlined />,
    tag: "error",
  },
  info: { color: "#1677ff", icon: <InfoCircleOutlined />, tag: "processing" },
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="Notification Center"
        description="List + Badge + Avatar + Tag + Empty combining into a realistic page"
      >
        <Card>
          <Flex
            justify="space-between"
            align="center"
            style={{ marginBottom: 16 }}
          >
            <Flex align="center" gap={8}>
              <Badge count={notifications.filter((n) => !n.read).length}>
                <BellOutlined style={{ fontSize: 20 }} />
              </Badge>
              <Typography.Title level={5} style={{ margin: 0 }}>
                Notifications
              </Typography.Title>
            </Flex>
            <Space>
              <Button size="small">Mark all as read</Button>
              <Button size="small">Settings</Button>
            </Space>
          </Flex>

          <List
            dataSource={notifications}
            renderItem={(item) => {
              const config = typeConfig[item.type]
              return (
                <List.Item
                  style={{
                    background: item.read
                      ? "transparent"
                      : "color-mix(in srgb, var(--ant-color-primary) 4%, transparent)",
                    padding: "12px 16px",
                    borderRadius: 6,
                    marginBottom: 4,
                  }}
                  actions={[
                    <Text type="secondary" key="time" style={{ fontSize: 12 }}>
                      {item.time}
                    </Text>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar
                        style={{ backgroundColor: config.color }}
                        icon={config.icon}
                      />
                    }
                    title={
                      <Flex align="center" gap={8}>
                        <Text strong={!item.read}>{item.title}</Text>
                        <Tag color={config.tag} style={{ fontSize: 10 }}>
                          {item.type}
                        </Tag>
                      </Flex>
                    }
                    description={item.description}
                  />
                </List.Item>
              )
            }}
          />

          <Flex justify="center" style={{ marginTop: 16 }}>
            <Pagination simple total={50} pageSize={10} />
          </Flex>
        </Card>
      </Section>

      <Section
        title="Empty State"
        description="Notification center with no notifications"
      >
        <Card>
          <Flex
            justify="space-between"
            align="center"
            style={{ marginBottom: 16 }}
          >
            <Flex align="center" gap={8}>
              <BellOutlined style={{ fontSize: 20 }} />
              <Typography.Title level={5} style={{ margin: 0 }}>
                Notifications
              </Typography.Title>
            </Flex>
          </Flex>
          <Empty description="No notifications yet" />
        </Card>
      </Section>
    </Flex>
  ),
}

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Card>
      <List
        dataSource={notifications.slice(0, 3)}
        renderItem={(item) => {
          const config = typeConfig[item.type]
          return (
            <List.Item>
              <List.Item.Meta
                avatar={
                  <Avatar
                    style={{ backgroundColor: config.color }}
                    icon={config.icon}
                  />
                }
                title={item.title}
                description={item.description}
              />
            </List.Item>
          )
        }}
      />
    </Card>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText("Deployment completed")).toBeInTheDocument()
  },
}

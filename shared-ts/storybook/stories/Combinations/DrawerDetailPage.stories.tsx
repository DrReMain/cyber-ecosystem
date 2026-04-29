import { EditOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Button,
  Card,
  Descriptions,
  Drawer,
  Flex,
  List,
  Space,
  Steps,
  Table,
  Tabs,
  Tag,
  Timeline,
  Typography,
} from "antd"
import { useState } from "react"

const meta: Meta = {
  title: "Antd/Combinations/DrawerDetailPage",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

function DrawerDetailPage({
  size,
  placement,
}: {
  size?: "default" | "large"
  placement?: "left" | "right" | "top" | "bottom"
}) {
  const [open, setOpen] = useState(true)

  const timelineItems = [
    {
      content: (
        <>
          <Typography.Text strong>Project Created</Typography.Text>
          <div>
            <Typography.Text type="secondary">
              Alice Chen created this project
            </Typography.Text>
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            2025-01-15 09:30
          </Typography.Text>
        </>
      ),
    },
    {
      color: "green",
      content: (
        <>
          <Typography.Text strong>Milestone Reached</Typography.Text>
          <div>
            <Typography.Text type="secondary">
              Phase 1 completed with all deliverables
            </Typography.Text>
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            2025-02-20 14:15
          </Typography.Text>
        </>
      ),
    },
    {
      color: "blue",
      content: (
        <>
          <Typography.Text strong>Team Expanded</Typography.Text>
          <div>
            <Typography.Text type="secondary">
              Bob Smith and Carol Lee joined the project
            </Typography.Text>
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            2025-03-10 11:00
          </Typography.Text>
        </>
      ),
    },
    {
      color: "red",
      content: (
        <>
          <Typography.Text strong>Issue Reported</Typography.Text>
          <div>
            <Typography.Text type="secondary">
              Critical bug found in production environment
            </Typography.Text>
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            2025-04-05 16:45
          </Typography.Text>
        </>
      ),
    },
    {
      content: (
        <>
          <Typography.Text strong>Status Updated</Typography.Text>
          <div>
            <Typography.Text type="secondary">
              Project moved to "On Hold" pending review
            </Typography.Text>
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            2025-05-12 10:20
          </Typography.Text>
        </>
      ),
    },
  ]

  const taskColumns = [
    { title: "Task", dataIndex: "task", key: "task" },
    {
      title: "Assignee",
      dataIndex: "assignee",
      key: "assignee",
      render: (name: string) => (
        <Space>
          <Avatar size="small">{name[0]}</Avatar>
          <span>{name}</span>
        </Space>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        const color =
          status === "Done"
            ? "success"
            : status === "In Progress"
              ? "processing"
              : "default"
        return <Tag color={color}>{status}</Tag>
      },
    },
  ]

  const taskData = [
    {
      key: "1",
      task: "Requirements Analysis",
      assignee: "Alice Chen",
      status: "Done",
    },
    { key: "2", task: "UI Design", assignee: "Bob Smith", status: "Done" },
    {
      key: "3",
      task: "Backend API",
      assignee: "Carol Lee",
      status: "In Progress",
    },
    {
      key: "4",
      task: "Frontend Integration",
      assignee: "David Wang",
      status: "Pending",
    },
    { key: "5", task: "Testing", assignee: "Eve Park", status: "Pending" },
  ]

  const commentData = [
    {
      title: "Alice Chen",
      avatar: "A",
      description: (
        <>
          <Typography.Text>
            Updated the project timeline based on new requirements.
          </Typography.Text>
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              2 hours ago
            </Typography.Text>
          </div>
        </>
      ),
    },
    {
      title: "Bob Smith",
      avatar: "B",
      description: (
        <>
          <Typography.Text>
            Design mockups are ready for review. Please check Figma link.
          </Typography.Text>
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              5 hours ago
            </Typography.Text>
          </div>
        </>
      ),
    },
    {
      title: "Carol Lee",
      avatar: "C",
      description: (
        <>
          <Typography.Text>
            API documentation updated. Breaking changes noted in changelog.
          </Typography.Text>
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              1 day ago
            </Typography.Text>
          </div>
        </>
      ),
    },
  ]

  return (
    <>
      <Button type="primary" onClick={() => setOpen(true)}>
        Open Drawer
      </Button>
      <Drawer
        title={
          <Flex justify="space-between" align="center">
            <Typography.Title level={5} style={{ margin: 0 }}>
              Project Details
            </Typography.Title>
            <Button type="text" icon={<EditOutlined />} size="small">
              Edit
            </Button>
          </Flex>
        }
        placement={placement}
        onClose={() => setOpen(false)}
        open={open}
        size={size}
        width={placement === "left" || placement === "right" ? 640 : undefined}
        height={placement === "top" || placement === "bottom" ? 500 : undefined}
      >
        <Flex vertical gap={24}>
          {/* Project Overview */}
          <Card size="small">
            <Descriptions column={2} size="small">
              <Descriptions.Item label="Name">Project Alpha</Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color="processing">Active</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Category">
                <Tag color="blue">Engineering</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Priority">
                <Tag color="red">High</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Owner">Alice Chen</Descriptions.Item>
              <Descriptions.Item label="Budget">$120k</Descriptions.Item>
              <Descriptions.Item label="Start">2025-01-15</Descriptions.Item>
              <Descriptions.Item label="End">2025-06-30</Descriptions.Item>
            </Descriptions>
          </Card>

          {/* Progress Steps */}
          <Card size="small" title="Progress">
            <Steps
              current={2}
              size="small"
              items={[
                { title: "Planning" },
                { title: "Design" },
                { title: "Development" },
                { title: "Testing" },
                { title: "Release" },
              ]}
            />
          </Card>

          {/* Tabs Section */}
          <Tabs
            defaultActiveKey="tasks"
            items={[
              {
                key: "tasks",
                label: "Tasks",
                children: (
                  <Table
                    columns={taskColumns}
                    dataSource={taskData}
                    pagination={false}
                    size="small"
                  />
                ),
              },
              {
                key: "timeline",
                label: "Timeline",
                children: <Timeline items={timelineItems} />,
              },
              {
                key: "comments",
                label: "Comments",
                children: (
                  <List
                    itemLayout="horizontal"
                    dataSource={commentData}
                    renderItem={(item) => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<Avatar>{item.avatar}</Avatar>}
                          title={item.title}
                          description={item.description}
                        />
                      </List.Item>
                    )}
                  />
                ),
              },
            ]}
          />
        </Flex>
      </Drawer>
    </>
  )
}

export const RightDefault: Story = {
  parameters: { controls: { disable: true } },
  render: () => <DrawerDetailPage />,
}

export const RightLarge: Story = {
  parameters: { controls: { disable: true } },
  render: () => <DrawerDetailPage size="large" />,
}

export const LeftDefault: Story = {
  parameters: { controls: { disable: true } },
  render: () => <DrawerDetailPage placement="left" />,
}

export const Bottom: Story = {
  parameters: { controls: { disable: true } },
  render: () => <DrawerDetailPage placement="bottom" />,
}

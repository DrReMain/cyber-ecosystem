import { EditOutlined, MailOutlined, PhoneOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Flex,
  Row,
  Tabs,
  Tag,
  Typography,
} from "antd"
import { expect, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Combinations/ProfilePage",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

const { Title, Text, Paragraph } = Typography

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="User Profile Page"
        description="Avatar + Descriptions + Tabs combining multiple components"
      >
        <Card>
          <Row gutter={24}>
            <Col span={6}>
              <Flex vertical align="center" gap={12}>
                <Avatar size={96} style={{ fontSize: 36 }}>
                  AJ
                </Avatar>
                <Title level={4} style={{ margin: 0 }}>
                  Alice Johnson
                </Title>
                <Tag color="blue">Admin</Tag>
                <Button icon={<EditOutlined />} size="small">
                  Edit Profile
                </Button>
              </Flex>
            </Col>
            <Col span={18}>
              <Descriptions column={2} size="small">
                <Descriptions.Item label="Email">
                  <Flex align="center" gap={4}>
                    <MailOutlined />
                    alice@example.com
                  </Flex>
                </Descriptions.Item>
                <Descriptions.Item label="Phone">
                  <Flex align="center" gap={4}>
                    <PhoneOutlined />
                    +1 (555) 123-4567
                  </Flex>
                </Descriptions.Item>
                <Descriptions.Item label="Department">
                  Engineering
                </Descriptions.Item>
                <Descriptions.Item label="Location">
                  San Francisco, CA
                </Descriptions.Item>
                <Descriptions.Item label="Joined">
                  January 2024
                </Descriptions.Item>
                <Descriptions.Item label="Status">
                  <Tag color="success">Active</Tag>
                </Descriptions.Item>
              </Descriptions>
            </Col>
          </Row>
        </Card>

        <Card>
          <Tabs
            defaultActiveKey="activity"
            items={[
              {
                key: "activity",
                label: "Activity",
                children: (
                  <Flex vertical gap={16}>
                    {[
                      {
                        action: "Updated project settings",
                        time: "2 hours ago",
                        tag: "Settings",
                      },
                      {
                        action: "Deployed v2.3.1 to production",
                        time: "5 hours ago",
                        tag: "Deploy",
                      },
                      {
                        action: "Reviewed pull request #142",
                        time: "1 day ago",
                        tag: "Review",
                      },
                      {
                        action: "Created new feature branch",
                        time: "2 days ago",
                        tag: "Git",
                      },
                    ].map((item) => (
                      <Flex
                        key={item.action}
                        justify="space-between"
                        align="center"
                      >
                        <Flex align="center" gap={8}>
                          <Tag>{item.tag}</Tag>
                          <Text>{item.action}</Text>
                        </Flex>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {item.time}
                        </Text>
                      </Flex>
                    ))}
                  </Flex>
                ),
              },
              {
                key: "projects",
                label: "Projects",
                children: (
                  <Row gutter={[16, 16]}>
                    {[
                      { name: "Web Platform", role: "Lead", status: "Active" },
                      {
                        name: "Mobile App",
                        role: "Contributor",
                        status: "Active",
                      },
                      {
                        name: "API Gateway",
                        role: "Reviewer",
                        status: "On Hold",
                      },
                    ].map((p) => (
                      <Col key={p.name} span={8}>
                        <Card size="small">
                          <Flex vertical gap={4}>
                            <Text strong>{p.name}</Text>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {p.role}
                            </Text>
                            <Tag
                              color={
                                p.status === "Active" ? "success" : "default"
                              }
                            >
                              {p.status}
                            </Tag>
                          </Flex>
                        </Card>
                      </Col>
                    ))}
                  </Row>
                ),
              },
              {
                key: "settings",
                label: "Settings",
                children: (
                  <Paragraph type="secondary">
                    Profile settings and notification preferences would go here.
                  </Paragraph>
                ),
              },
            ]}
          />
        </Card>
      </Section>
    </Flex>
  ),
}

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Card>
      <Flex align="center" gap={12}>
        <Avatar size={48}>AJ</Avatar>
        <Flex vertical>
          <Text strong>Alice Johnson</Text>
          <Text type="secondary">alice@example.com</Text>
        </Flex>
      </Flex>
    </Card>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText("Alice Johnson")).toBeInTheDocument()
  },
}

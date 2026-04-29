import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  MoreOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Button,
  Card,
  Flex,
  Progress,
  Space,
  Statistic,
  Tag,
  Typography,
} from "antd"

const meta: Meta<typeof Card> = {
  title: "Antd/Combinations/DashboardCard",
  component: Card,
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj<typeof Card>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex gap={16} wrap style={{ maxWidth: 900 }}>
      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="start">
          <Statistic
            title="Total Revenue"
            value={112893}
            precision={2}
            prefix="$"
            styles={{ content: { fontSize: 28 } }}
          />
          <Tag color="green">
            <ArrowUpOutlined /> 12.5%
          </Tag>
        </Flex>
        <Progress percent={75} showInfo={false} style={{ marginTop: 12 }} />
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          75% of monthly target
        </Typography.Text>
      </Card>

      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="start">
          <Statistic
            title="Active Users"
            value={2847}
            styles={{ content: { fontSize: 28 } }}
          />
          <Tag color="red">
            <ArrowDownOutlined /> 3.2%
          </Tag>
        </Flex>
        <Flex gap={8} style={{ marginTop: 12 }}>
          <Avatar.Group max={{ count: 4 }}>
            {["A", "B", "C", "D", "E"].map((s) => (
              <Avatar key={s}>{s}</Avatar>
            ))}
          </Avatar.Group>
        </Flex>
      </Card>

      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="center">
          <Typography.Text strong>Recent Orders</Typography.Text>
          <Button variant="text" size="small" icon={<MoreOutlined />} />
        </Flex>
        <Flex vertical gap={8} style={{ marginTop: 12 }}>
          {[
            { name: "Alice", amount: "$120.00", status: "Completed" as const },
            { name: "Bob", amount: "$85.50", status: "Processing" as const },
            { name: "Carol", amount: "$240.00", status: "Pending" as const },
          ].map((order) => (
            <Flex key={order.name} justify="space-between" align="center">
              <Space>
                <Avatar size="small">{order.name[0]}</Avatar>
                <Typography.Text>{order.name}</Typography.Text>
              </Space>
              <Space>
                <Typography.Text>{order.amount}</Typography.Text>
                <Tag
                  color={
                    order.status === "Completed"
                      ? "green"
                      : order.status === "Processing"
                        ? "blue"
                        : "orange"
                  }
                >
                  {order.status}
                </Tag>
              </Space>
            </Flex>
          ))}
        </Flex>
      </Card>
    </Flex>
  ),
}

export const Playground: Story = {
  render: () => (
    <Flex gap={16} wrap style={{ maxWidth: 900 }}>
      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="start">
          <Statistic
            title="Total Revenue"
            value={112893}
            precision={2}
            prefix="$"
          />
          <Tag color="green">
            <ArrowUpOutlined /> 12.5%
          </Tag>
        </Flex>
        <Progress percent={75} showInfo={false} style={{ marginTop: 12 }} />
      </Card>

      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="start">
          <Statistic title="Active Users" value={2847} />
          <Tag color="red">
            <ArrowDownOutlined /> 3.2%
          </Tag>
        </Flex>
        <Flex gap={8} style={{ marginTop: 12 }}>
          <Avatar.Group max={{ count: 4 }}>
            {["A", "B", "C", "D", "E"].map((s) => (
              <Avatar key={s}>{s}</Avatar>
            ))}
          </Avatar.Group>
        </Flex>
      </Card>

      <Card style={{ flex: "1 1 280px" }}>
        <Flex justify="space-between" align="center">
          <Typography.Text strong>Recent Orders</Typography.Text>
          <Button variant="text" size="small" icon={<MoreOutlined />} />
        </Flex>
        <Flex vertical gap={8} style={{ marginTop: 12 }}>
          {[
            { name: "Alice", amount: "$120.00", status: "Completed" as const },
            { name: "Bob", amount: "$85.50", status: "Processing" as const },
            { name: "Carol", amount: "$240.00", status: "Pending" as const },
          ].map((order) => (
            <Flex key={order.name} justify="space-between" align="center">
              <Space>
                <Avatar size="small">{order.name[0]}</Avatar>
                <Typography.Text>{order.name}</Typography.Text>
              </Space>
              <Space>
                <Typography.Text>{order.amount}</Typography.Text>
                <Tag
                  color={
                    order.status === "Completed"
                      ? "green"
                      : order.status === "Processing"
                        ? "blue"
                        : "orange"
                  }
                >
                  {order.status}
                </Tag>
              </Space>
            </Flex>
          ))}
        </Flex>
      </Card>
    </Flex>
  ),
}

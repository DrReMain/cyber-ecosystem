import { BellOutlined, SearchOutlined, UserOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Flex,
  Menu,
  Space,
  Typography,
} from "antd"

const meta: Meta<typeof Menu> = {
  title: "Antd/Combinations/NavigationHeader",
  component: Menu,
  parameters: { layout: "fullscreen" },
}

export default meta
type Story = StoryObj<typeof Menu>

const menuItems = [
  { key: "dashboard", label: "Dashboard" },
  { key: "projects", label: "Projects" },
  { key: "analytics", label: "Analytics" },
  { key: "settings", label: "Settings" },
]

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical>
      <Flex
        justify="space-between"
        align="center"
        style={{
          padding: "0 24px",
          height: 48,
          borderBottom: "1px solid var(--ant-color-border-secondary)",
        }}
      >
        <Typography.Text strong style={{ fontSize: 16 }}>
          MyApp
        </Typography.Text>
        <Menu
          mode="horizontal"
          defaultSelectedKeys={["dashboard"]}
          items={menuItems}
          style={{ flex: 1, marginLeft: 32, border: "none" }}
        />
        <Space size="middle">
          <Button variant="text" icon={<SearchOutlined />} />
          <Badge count={3} size="small">
            <Button variant="text" icon={<BellOutlined />} />
          </Badge>
          <Avatar icon={<UserOutlined />} />
        </Space>
      </Flex>
      <div style={{ padding: "16px 24px" }}>
        <Breadcrumb
          items={[
            { title: "Home" },
            { title: "Dashboard" },
            { title: "Overview" },
          ]}
        />
      </div>
    </Flex>
  ),
}

export const Playground: Story = {
  render: () => (
    <Flex vertical>
      <Flex
        justify="space-between"
        align="center"
        style={{
          padding: "0 24px",
          height: 48,
          borderBottom: "1px solid var(--ant-color-border-secondary)",
        }}
      >
        <Typography.Text strong style={{ fontSize: 16 }}>
          MyApp
        </Typography.Text>
        <Menu
          mode="horizontal"
          defaultSelectedKeys={["dashboard"]}
          items={menuItems}
          style={{ flex: 1, marginLeft: 32, border: "none" }}
        />
        <Space size="middle">
          <Button variant="text" icon={<SearchOutlined />} />
          <Badge count={3} size="small">
            <Button variant="text" icon={<BellOutlined />} />
          </Badge>
          <Avatar icon={<UserOutlined />} />
        </Space>
      </Flex>
      <div style={{ padding: "16px 24px" }}>
        <Breadcrumb
          items={[
            { title: "Home" },
            { title: "Dashboard" },
            { title: "Overview" },
          ]}
        />
      </div>
    </Flex>
  ),
}

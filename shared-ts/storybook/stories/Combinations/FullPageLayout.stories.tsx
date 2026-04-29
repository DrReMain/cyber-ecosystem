import {
  BellOutlined,
  DashboardOutlined,
  FileOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Card,
  Flex,
  Layout,
  Menu,
  Space,
  Table,
  Tag,
  Typography,
} from "antd"
import { useState } from "react"
import { statusColorMap } from "../fixtures"

const { Header, Sider, Content, Footer } = Layout

const meta: Meta<typeof Layout> = {
  title: "Antd/Combinations/FullPageLayout",
  component: Layout,
  parameters: { layout: "fullscreen" },
}

export default meta
type Story = StoryObj<typeof Layout>

const siderItems = [
  { key: "dashboard", icon: <DashboardOutlined />, label: "Dashboard" },
  { key: "projects", icon: <FileOutlined />, label: "Projects" },
  { key: "team", icon: <TeamOutlined />, label: "Team" },
  { key: "settings", icon: <SettingOutlined />, label: "Settings" },
]

const tableColumns = [
  { title: "Name", dataIndex: "name", key: "name" },
  {
    title: "Status",
    dataIndex: "status",
    key: "status",
    render: (s: string) => {
      const color = statusColorMap[s] ?? "default"
      return <Tag color={color}>{s}</Tag>
    },
  },
  { title: "Role", dataIndex: "role", key: "role" },
]

const tableData = [
  { key: "1", name: "Alice Chen", status: "Active", role: "Admin" },
  { key: "2", name: "Bob Smith", status: "Pending", role: "Editor" },
  { key: "3", name: "Carol Lee", status: "Active", role: "Viewer" },
]

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [collapsed, setCollapsed] = useState(false)

    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          trigger={null}
        >
          <div
            style={{
              height: 48,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              borderBottom: "1px solid rgba(255,255,255,0.1)",
            }}
          >
            <Typography.Text strong style={{ fontSize: collapsed ? 14 : 16 }}>
              {collapsed ? "M" : "MyApp"}
            </Typography.Text>
          </div>
          <Menu
            theme="dark"
            mode="inline"
            defaultSelectedKeys={["dashboard"]}
            items={siderItems}
          />
        </Sider>
        <Layout>
          <Header
            style={{
              background: "var(--ant-color-bg-container)",
              padding: "0 24px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              borderBottom: "1px solid var(--ant-color-border-secondary)",
            }}
          >
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
            />
            <Space size="middle">
              <Badge count={3} size="small">
                <Button variant="text" icon={<BellOutlined />} />
              </Badge>
              <Avatar icon={<UserOutlined />} />
            </Space>
          </Header>
          <Content style={{ padding: 24, overflow: "auto" }}>
            <Breadcrumb
              style={{ marginBottom: 16 }}
              items={[
                { title: "Home" },
                { title: "Dashboard" },
                { title: "Overview" },
              ]}
            />
            <Typography.Title
              level={4}
              style={{ marginTop: 0, marginBottom: 24 }}
            >
              Dashboard Overview
            </Typography.Title>
            <Flex gap={16} wrap style={{ marginBottom: 24 }}>
              {[
                { title: "Total Users", value: 2847, change: "+12.5%" },
                { title: "Revenue", value: "$42,580", change: "+8.2%" },
                { title: "Orders", value: 156, change: "-3.1%" },
              ].map((stat) => (
                <Card
                  key={stat.title}
                  size="small"
                  style={{ flex: "1 1 200px" }}
                >
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {stat.title}
                  </Typography.Text>
                  <div style={{ fontSize: 24, fontWeight: 600 }}>
                    {stat.value}
                  </div>
                  <Tag
                    color={stat.change.startsWith("+") ? "green" : "red"}
                    style={{ marginTop: 4 }}
                  >
                    {stat.change}
                  </Tag>
                </Card>
              ))}
            </Flex>
            <Card title="Team Members" size="small">
              <Table
                columns={tableColumns}
                dataSource={tableData}
                pagination={false}
                size="small"
              />
            </Card>
          </Content>
          <Footer
            style={{
              textAlign: "center",
              padding: "12px 24px",
              background: "var(--ant-color-bg-layout)",
            }}
          >
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              MyApp 2026
            </Typography.Text>
          </Footer>
        </Layout>
      </Layout>
    )
  },
}

export const Playground: Story = {
  render: () => {
    const [collapsed, setCollapsed] = useState(false)

    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          trigger={null}
        >
          <div
            style={{
              height: 48,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <Typography.Text strong>
              {collapsed ? "M" : "MyApp"}
            </Typography.Text>
          </div>
          <Menu
            theme="dark"
            mode="inline"
            defaultSelectedKeys={["dashboard"]}
            items={siderItems}
          />
        </Sider>
        <Layout>
          <Header
            style={{
              background: "var(--ant-color-bg-container)",
              padding: "0 24px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
            />
            <Avatar icon={<UserOutlined />} />
          </Header>
          <Content style={{ padding: 24 }}>
            <Typography.Title level={4}>Dashboard</Typography.Title>
            <Card title="Recent Activity" size="small">
              <Typography.Text>Content area</Typography.Text>
            </Card>
          </Content>
        </Layout>
      </Layout>
    )
  },
}

import {
  DashboardOutlined,
  HomeOutlined,
  SettingOutlined,
  UserOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Layout, Menu, theme } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Layout> = {
  title: "Antd/Layout/Layout",
  component: Layout,
  parameters: { layout: "fullscreen" },
  args: {},
  argTypes: {},
}

export default meta
type Story = StoryObj<typeof Layout>

const { Header, Sider, Content, Footer } = Layout

const siderMenuItems = [
  { key: "1", icon: <HomeOutlined />, label: "Home" },
  { key: "2", icon: <DashboardOutlined />, label: "Dashboard" },
  { key: "3", icon: <UserOutlined />, label: "Users" },
  { key: "4", icon: <SettingOutlined />, label: "Settings" },
]

const headerMenuItems = [
  { key: "1", label: "Nav 1" },
  { key: "2", label: "Nav 2" },
  { key: "3", label: "Nav 3" },
]

function ContentArea({ children }: { children: React.ReactNode }) {
  const { token } = theme.useToken()
  return (
    <div
      style={{
        padding: 24,
        background: token.colorBgContainer,
        minHeight: 200,
        borderRadius: token.borderRadiusLG,
      }}
    >
      {children}
    </div>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24} style={{ padding: 24 }}>
      <Section title="Basic (Header + Content + Footer)">
        <Layout style={{ minHeight: 400 }}>
          <Header style={{ display: "flex", alignItems: "center" }}>
            <div
              style={{
                color: "var(--ant-color-text)",
                fontSize: 18,
                fontWeight: 600,
                marginRight: 40,
              }}
            >
              Logo
            </div>
            <Menu
              theme="dark"
              mode="horizontal"
              defaultSelectedKeys={["1"]}
              items={headerMenuItems}
              style={{ flex: 1, minWidth: 0 }}
            />
          </Header>
          <Content style={{ padding: "0 48px" }}>
            <ContentArea>
              <p>Main content area</p>
            </ContentArea>
          </Content>
          <Footer style={{ textAlign: "center" }}>Footer Content</Footer>
        </Layout>
      </Section>

      <Section title="Sider (dark theme, collapsible)">
        <Layout style={{ minHeight: 400 }}>
          <Sider collapsible>
            <div
              style={{
                height: 32,
                margin: 16,
                background: "var(--ant-color-bg-container)",
                borderRadius: 6,
              }}
            />
            <Menu
              theme="dark"
              defaultSelectedKeys={["1"]}
              mode="inline"
              items={siderMenuItems}
            />
          </Sider>
          <Layout>
            <Header
              style={{
                background: "transparent",
                padding: "0 16px",
              }}
            >
              <span style={{ fontSize: 16, fontWeight: 500 }}>Header</span>
            </Header>
            <Content style={{ margin: "16px" }}>
              <ContentArea>
                <p>Content area</p>
              </ContentArea>
            </Content>
            <Footer style={{ textAlign: "center" }}>Footer</Footer>
          </Layout>
        </Layout>
      </Section>

      <Section title="Sider (light theme)">
        <Layout style={{ minHeight: 300, borderRadius: 8, overflow: "hidden" }}>
          <Sider theme="light">
            <div style={{ padding: 16 }}>Light Sider</div>
            <Menu
              theme="light"
              defaultSelectedKeys={["1"]}
              mode="inline"
              items={siderMenuItems}
            />
          </Sider>
          <Layout>
            <Content style={{ padding: 24 }}>
              <ContentArea>
                <p>Content next to light sider.</p>
              </ContentArea>
            </Content>
          </Layout>
        </Layout>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  parameters: { layout: "fullscreen" },
  render: (args) => (
    <Layout {...args} style={{ minHeight: 400 }}>
      <Header style={{ display: "flex", alignItems: "center" }}>
        <div
          style={{
            color: "var(--ant-color-text)",
            fontSize: 18,
            fontWeight: 600,
            marginRight: 40,
          }}
        >
          Logo
        </div>
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={["1"]}
          items={headerMenuItems}
          style={{ flex: 1, minWidth: 0 }}
        />
      </Header>
      <Content style={{ padding: "0 48px" }}>
        <ContentArea>
          <p>Main content area</p>
        </ContentArea>
      </Content>
      <Footer style={{ textAlign: "center" }}>Footer Content</Footer>
    </Layout>
  ),
}

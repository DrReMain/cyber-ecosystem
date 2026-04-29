import {
  AppstoreOutlined,
  DesktopOutlined,
  FormOutlined,
  HomeOutlined,
  MailOutlined,
  PieChartOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Layout, Menu, Typography } from "antd"
import { expect, fn, within } from "storybook/test"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const horizontalItems = [
  { key: "mail", icon: <MailOutlined />, label: "Navigation One" },
  { key: "app", icon: <AppstoreOutlined />, label: "Navigation Two" },
  { key: "settings", icon: <SettingOutlined />, label: "Navigation Three" },
]

const submenuItems = [
  {
    key: "sub1",
    icon: <MailOutlined />,
    label: "Navigation One",
    children: [
      { key: "1", label: "Option 1" },
      { key: "2", label: "Option 2" },
      { key: "3", label: "Option 3" },
      { key: "4", label: "Option 4" },
    ],
  },
  {
    key: "sub2",
    icon: <AppstoreOutlined />,
    label: "Navigation Two",
    children: [
      { key: "5", label: "Option 5" },
      { key: "6", label: "Option 6" },
      {
        key: "sub3",
        label: "Submenu",
        children: [
          { key: "7", label: "Option 7" },
          { key: "8", label: "Option 8" },
        ],
      },
    ],
  },
  {
    key: "sub4",
    icon: <SettingOutlined />,
    label: "Navigation Three",
    children: [
      { key: "9", label: "Option 9" },
      { key: "10", label: "Option 10" },
      { key: "11", label: "Option 11" },
      { key: "12", label: "Option 12" },
    ],
  },
]

const iconItems = [
  { key: "1", icon: <PieChartOutlined />, label: "Dashboard" },
  { key: "2", icon: <DesktopOutlined />, label: "Workspace" },
  { key: "3", icon: <UserOutlined />, label: "Users" },
  { key: "4", icon: <TeamOutlined />, label: "Teams" },
  { key: "5", icon: <FormOutlined />, label: "Forms" },
]

const meta: Meta<typeof Menu> = {
  title: "Antd/Navigation/Menu",
  component: Menu,
  parameters: { layout: "padded" },
  args: {
    mode: "horizontal",
    defaultSelectedKeys: ["mail"],
    items: horizontalItems,
  },
  argTypes: {
    mode: {
      control: "radio",
      options: ["horizontal", "vertical", "inline"],
    },
    theme: {
      control: "radio",
      options: ["light", "dark"],
    },
    multiple: { control: "boolean" },
    onClick: { action: "clicked" },
    onSelect: { action: "selected" },
    onOpenChange: { action: "openChanged" },
  },
}

export default meta
type Story = StoryObj<typeof Menu>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      {/* 1. Horizontal */}
      <Section title="Horizontal">
        <Menu
          mode="horizontal"
          defaultSelectedKeys={["mail"]}
          items={horizontalItems}
        />
      </Section>

      {/* 2. Vertical */}
      <Section title="Vertical">
        <Menu
          style={{ width: 256 }}
          mode="vertical"
          defaultSelectedKeys={["1"]}
          defaultOpenKeys={["sub1"]}
          items={submenuItems}
        />
      </Section>

      {/* 3. Inline */}
      <Section title="Inline">
        <Menu
          style={{ width: 256 }}
          mode="inline"
          defaultSelectedKeys={["1"]}
          defaultOpenKeys={["sub1"]}
          items={submenuItems}
        />
      </Section>

      {/* 4. Themes */}
      <Section title="Themes">
        <Flex gap={24}>
          <Flex vertical gap={4}>
            <Label>light (default)</Label>
            <Menu
              style={{ width: 256 }}
              mode="inline"
              defaultSelectedKeys={["1"]}
              defaultOpenKeys={["sub1"]}
              items={submenuItems.slice(0, 2)}
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>dark</Label>
            <Menu
              style={{ width: 256 }}
              theme="dark"
              mode="inline"
              defaultSelectedKeys={["1"]}
              defaultOpenKeys={["sub1"]}
              items={submenuItems.slice(0, 2)}
            />
          </Flex>
        </Flex>
      </Section>

      {/* 5. With Icons */}
      <Section title="With Icons">
        <Menu
          style={{ width: 256 }}
          mode="vertical"
          defaultSelectedKeys={["1"]}
          items={iconItems}
        />
      </Section>

      {/* 6. Disabled & Danger Items */}
      <Section title="Disabled & Danger Items">
        <Menu
          style={{ width: 256 }}
          mode="vertical"
          defaultSelectedKeys={["1"]}
          items={[
            { key: "1", label: "Normal Item" },
            { key: "2", label: "Disabled Item", disabled: true },
            { key: "3", label: "Danger Item", danger: true },
            { key: "4", icon: <SettingOutlined />, label: "With Icon" },
            {
              key: "5",
              icon: <SettingOutlined />,
              label: "Disabled + Icon",
              disabled: true,
            },
          ]}
        />
      </Section>

      <Section title="States">
        <PseudoStates>
          <Menu
            mode="horizontal"
            defaultSelectedKeys={["mail"]}
            items={horizontalItems}
          />
        </PseudoStates>
      </Section>

      <Section title="Collapsed Inline">
        <Menu
          style={{ width: 80 }}
          mode="inline"
          inlineCollapsed
          defaultSelectedKeys={["1"]}
          items={iconItems}
        />
      </Section>

      <Section
        title="Sidebar Layout"
        description="Menu in its primary context — sidebar navigation within a page layout"
      >
        <div
          style={{
            height: 320,
            borderRadius: 8,
            overflow: "hidden",
            border: "1px solid var(--ant-color-border)",
          }}
        >
          <Layout style={{ height: "100%" }}>
            <Layout.Sider
              width={200}
              style={{ background: "var(--ant-color-bg-container)" }}
            >
              <Flex vertical style={{ padding: "12px 0" }}>
                <Typography.Text
                  strong
                  style={{ padding: "0 16px 12px", fontSize: 14 }}
                >
                  My App
                </Typography.Text>
                <Menu
                  mode="inline"
                  defaultSelectedKeys={["1"]}
                  items={[
                    { key: "1", icon: <HomeOutlined />, label: "Dashboard" },
                    { key: "2", icon: <UserOutlined />, label: "Users" },
                    { key: "3", icon: <SettingOutlined />, label: "Settings" },
                    {
                      key: "sub1",
                      icon: <MailOutlined />,
                      label: "Messages",
                      children: [
                        { key: "4", label: "Inbox" },
                        { key: "5", label: "Sent" },
                      ],
                    },
                  ]}
                  style={{ border: "none" }}
                />
              </Flex>
            </Layout.Sider>
            <Layout>
              <Layout.Content
                style={{
                  padding: 16,
                  background: "var(--ant-color-bg-layout)",
                }}
              >
                <Typography.Title level={5} style={{ margin: 0 }}>
                  Dashboard
                </Typography.Title>
                <Typography.Text type="secondary">
                  Welcome back! Here is your overview.
                </Typography.Text>
              </Layout.Content>
            </Layout>
          </Layout>
        </div>
      </Section>

      <Section
        title="Dark Sidebar Layout"
        description="Dark theme menu in sidebar context — critical for skin validation"
      >
        <div
          style={{
            height: 320,
            borderRadius: 8,
            overflow: "hidden",
            border: "1px solid var(--ant-color-border)",
          }}
        >
          <Layout style={{ height: "100%" }}>
            <Layout.Sider width={200} theme="dark">
              <Flex vertical style={{ padding: "12px 0" }}>
                <Typography.Text
                  strong
                  style={{
                    padding: "0 16px 12px",
                    fontSize: 14,
                    color: "#fff",
                  }}
                >
                  My App
                </Typography.Text>
                <Menu
                  theme="dark"
                  mode="inline"
                  defaultSelectedKeys={["1"]}
                  items={[
                    { key: "1", icon: <HomeOutlined />, label: "Dashboard" },
                    { key: "2", icon: <UserOutlined />, label: "Users" },
                    { key: "3", icon: <SettingOutlined />, label: "Settings" },
                    {
                      key: "sub1",
                      icon: <MailOutlined />,
                      label: "Messages",
                      children: [
                        { key: "4", label: "Inbox" },
                        { key: "5", label: "Sent" },
                      ],
                    },
                  ]}
                  style={{ border: "none" }}
                />
              </Flex>
            </Layout.Sider>
            <Layout>
              <Layout.Content
                style={{
                  padding: 16,
                  background: "var(--ant-color-bg-layout)",
                }}
              >
                <Typography.Title level={5} style={{ margin: 0 }}>
                  Dashboard
                </Typography.Title>
                <Typography.Text type="secondary">
                  Dark sidebar with skin-aware styling.
                </Typography.Text>
              </Layout.Content>
            </Layout>
          </Layout>
        </div>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    mode: "horizontal",
    defaultSelectedKeys: ["mail"],
    items: horizontalItems,
    onClick: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const item = canvas.getByText("Navigation Two")
    await item.click()
    await expect(args.onClick).toHaveBeenCalledTimes(1)
  },
}

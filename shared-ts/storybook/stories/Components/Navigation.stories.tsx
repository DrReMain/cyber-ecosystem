import type { Meta, StoryObj } from "@storybook/react-vite";
import { Breadcrumb, Layout, Menu, Pagination, Space, Tabs, Typography } from "antd";
import { useState } from "react";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Components/Navigation",
};

export default meta;

type Story = StoryObj;

const MENU_ITEMS = [
  { key: "workbench", label: "Workbench" },
  {
    key: "ops",
    label: "Operations",
    children: [
      { key: "drones", label: "Drones" },
      { key: "missions", label: "Missions" },
      { key: "telemetry", label: "Telemetry" },
    ],
  },
  {
    key: "system",
    label: "System",
    children: [
      { key: "users", label: "Users" },
      { key: "roles", label: "Roles" },
      { key: "diag", label: "Diagnostics" },
    ],
  },
];

function NavigationPage() {
  const [menuKey, setMenuKey] = useState("workbench");
  const [tabKey, setTabKey] = useState("overview");
  return (
    <div style={{ maxWidth: 900 }}>
      <Section
        description="Sider chrome geometry (item height, radius, selected/hover fills, popup)."
        title="Menu in Layout.Sider"
      >
        <Layout
          hasSider
          style={{
            border: "1px solid rgba(128,128,128,0.25)",
            borderRadius: 8,
            overflow: "hidden",
          }}
        >
          <Layout.Sider style={{ background: "transparent" }} theme="light" width={208}>
            <Menu
              items={MENU_ITEMS}
              mode="inline"
              onClick={(e) => setMenuKey(e.key)}
              openKeys={["ops"]}
              selectedKeys={[menuKey]}
              style={{ borderInlineEnd: "none", height: "100%" }}
            />
          </Layout.Sider>
          <Layout.Content style={{ padding: 16 }}>
            <Typography.Text type="secondary">Selected: {menuKey}</Typography.Text>
          </Layout.Content>
        </Layout>
      </Section>

      <Section title="Breadcrumb & tabs">
        <Space vertical size={16} style={{ display: "flex" }}>
          <Breadcrumb
            items={[{ title: "Workbench" }, { title: "Operations" }, { title: "Drones" }]}
          />
          <Tabs
            activeKey={tabKey}
            items={[
              { key: "overview", label: "Overview", children: <p>Overview pane content.</p> },
              { key: "specs", label: "Specs", children: <p>Specs pane content.</p> },
              { key: "logs", label: "Logs", children: <p>Logs pane content.</p> },
            ]}
            onChange={setTabKey}
          />
        </Space>
      </Section>

      <Section title="Pagination">
        <Pagination defaultCurrent={3} showQuickJumper total={87} />
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <NavigationPage />,
};

import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, Card, Input, Segmented, Select, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Combinations/Table Page",
};

export default meta;

type Story = StoryObj;

interface Role {
  key: string;
  name: string;
  code: string;
  grants: number;
  users: number;
  status: "enabled" | "disabled";
}

const DATA: Role[] = [
  { key: "1", name: "Super Admin", code: "superadmin", grants: 0, users: 3, status: "enabled" },
  {
    key: "2",
    name: "Fleet Operator",
    code: "fleet_operator",
    grants: 12,
    users: 8,
    status: "enabled",
  },
  { key: "3", name: "Readonly Auditor", code: "auditor", grants: 4, users: 2, status: "enabled" },
  { key: "4", name: "Intern (legacy)", code: "intern", grants: 2, users: 1, status: "disabled" },
];

const COLUMNS: ColumnsType<Role> = [
  { title: "Role", dataIndex: "name", key: "name" },
  {
    title: "Code",
    dataIndex: "code",
    key: "code",
    render: (v: string) => <Typography.Text code>{v}</Typography.Text>,
  },
  { title: "Grants", dataIndex: "grants", key: "grants", align: "right" },
  { title: "Users", dataIndex: "users", key: "users", align: "right" },
  {
    title: "Status",
    dataIndex: "status",
    key: "status",
    render: (v: Role["status"]) => <Tag color={v === "enabled" ? "success" : "default"}>{v}</Tag>,
  },
];

function TablePage() {
  const [view, setView] = useState("table");
  return (
    <div style={{ maxWidth: 960 }}>
      <Section
        description="Filter row + toolbar + table + pager: the density-heavy layout skins must survive."
        title="Roles list"
      >
        <Card>
          <Space vertical size={16} style={{ display: "flex" }}>
            <Space wrap size={12}>
              <Input placeholder="Keyword" style={{ width: 220 }} />
              <Select
                allowClear
                options={[
                  { value: "enabled", label: "Enabled" },
                  { value: "disabled", label: "Disabled" },
                ]}
                placeholder="Status"
                style={{ width: 140 }}
              />
              <Button type="primary">Search</Button>
              <Button>Reset</Button>
            </Space>
            <Space size={12} style={{ justifyContent: "space-between", display: "flex" }}>
              <Segmented
                onChange={(v) => setView(v as string)}
                options={[
                  { value: "table", label: "Table" },
                  { value: "cards", label: "Cards" },
                ]}
                value={view}
              />
              <Button type="primary">New role</Button>
            </Space>
            <Table
              columns={COLUMNS}
              dataSource={DATA}
              pagination={{ total: 4, showTotal: (t) => `${t} roles` }}
            />
          </Space>
        </Card>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <TablePage />,
};

import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge, Space, Switch, Table, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import { Section } from "../helpers";

const meta: Meta<typeof Table> = {
  title: "Components/Table",
  component: Table,
};

export default meta;

type Story = StoryObj<typeof Table>;

interface Row {
  key: string;
  name: string;
  fleet: string;
  battery: number;
  status: "online" | "patrol" | "charging" | "offline";
  updated: string;
}

const DATA: Row[] = [
  {
    key: "1",
    name: "SPOT-01",
    fleet: "North substation",
    battery: 92,
    status: "patrol",
    updated: "12:41",
  },
  {
    key: "2",
    name: "SPOT-02",
    fleet: "North substation",
    battery: 67,
    status: "online",
    updated: "12:38",
  },
  {
    key: "3",
    name: "ANY-400-A",
    fleet: "East corridor",
    battery: 24,
    status: "charging",
    updated: "12:20",
  },
  {
    key: "4",
    name: "ANY-400-B",
    fleet: "East corridor",
    battery: 0,
    status: "offline",
    updated: "11:57",
  },
  {
    key: "5",
    name: "QUAD-07",
    fleet: "River crossing",
    battery: 81,
    status: "patrol",
    updated: "12:40",
  },
];

const STATUS_TAG: Record<Row["status"], { color: string; label: string }> = {
  online: { color: "success", label: "Online" },
  patrol: { color: "processing", label: "Patrol" },
  charging: { color: "warning", label: "Charging" },
  offline: { color: "default", label: "Offline" },
};

const COLUMNS: ColumnsType<Row> = [
  { title: "Robot", dataIndex: "name", key: "name" },
  { title: "Fleet", dataIndex: "fleet", key: "fleet" },
  {
    title: "Battery",
    dataIndex: "battery",
    key: "battery",
    sorter: (a, b) => a.battery - b.battery,
    render: (v: number) => (
      <Space>
        <Badge
          count={v}
          style={{
            backgroundColor: v > 50 ? "#52c41a" : v > 20 ? "#faad14" : "#ff4d4f",
          }}
        />
        <span>{v}%</span>
      </Space>
    ),
  },
  {
    title: "Status",
    dataIndex: "status",
    key: "status",
    filters: Object.entries(STATUS_TAG).map(([value, v]) => ({
      value,
      text: v.label,
    })),
    onFilter: (value, record) => record.status === value,
    render: (v: Row["status"]) => <Tag color={STATUS_TAG[v].color}>{STATUS_TAG[v].label}</Tag>,
  },
  { title: "Updated", dataIndex: "updated", key: "updated" },
];

function TablePage() {
  const [bordered, setBordered] = useState(false);
  const [stripe, setStripe] = useState(false);
  return (
    <div style={{ maxWidth: 900 }}>
      <Section
        description="Header, rows, hover fill, selected keys, filters and pager as the skin paints them."
        title="Data table"
      >
        <Space vertical size={16} style={{ display: "flex" }}>
          <Space size={20}>
            <Switch
              checked={bordered}
              checkedChildren="bordered"
              onChange={setBordered}
              unCheckedChildren="bordered"
            />
            <Switch
              checked={stripe}
              checkedChildren="striped"
              onChange={setStripe}
              unCheckedChildren="striped"
            />
          </Space>
          <Table
            bordered={bordered}
            columns={COLUMNS}
            dataSource={DATA}
            pagination={{ pageSize: 3, showSizeChanger: true }}
            rowSelection={{}}
            sticky
          />
        </Space>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <TablePage />,
};

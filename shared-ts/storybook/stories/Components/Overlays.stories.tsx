import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Button,
  DatePicker,
  Drawer,
  Dropdown,
  Input,
  Modal,
  Popover,
  Select,
  Space,
  Table,
  Tooltip,
  Typography,
} from "antd";
import { useState } from "react";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Components/Overlays",
};

export default meta;

type Story = StoryObj;

const POPOVER_ROWS = [
  { key: "1", metric: "battery", value: "92%" },
  { key: "2", metric: "link", value: "stable" },
  { key: "3", metric: "temp", value: "41°C" },
];

function OverlaysPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);

  return (
    <div style={{ maxWidth: 760 }}>
      <Section
        description="Elevated surfaces: mask, popup container, radius and shadow."
        title="Modal & drawer"
      >
        <Space>
          <Button onClick={() => setModalOpen(true)}>Open modal</Button>
          <Button onClick={() => setDrawerOpen(true)}>Open drawer</Button>
        </Space>
        <Modal
          cancelText="Cancel"
          okText="Confirm"
          onCancel={() => setModalOpen(false)}
          onOk={() => setModalOpen(false)}
          open={modalOpen}
          title="Confirm deployment"
        >
          <p>
            Modal body copy on the elevated surface. Fields below show how controls read on top of
            it.
          </p>
          <Space vertical size={8} style={{ display: "flex" }}>
            <Input placeholder="Mission name" />
            <DatePicker style={{ width: "100%" }} />
          </Space>
        </Modal>
        <Drawer
          onClose={() => setDrawerOpen(false)}
          open={drawerOpen}
          title="Robot detail"
          width={420}
        >
          <p>Drawer body copy on the elevated surface.</p>
          <Select
            options={[
              { value: 1, label: "Option A" },
              { value: 2, label: "Option B" },
            ]}
            placeholder="Popup from inside the drawer"
            style={{ width: 260 }}
          />
        </Drawer>
      </Section>

      <Section title="Anchored popups">
        <Space size={16} wrap>
          <Tooltip title="Tooltip on the inverse surface">
            <Button>Hover for tooltip</Button>
          </Tooltip>
          <Popover
            content={
              <Table
                columns={[
                  { title: "Metric", dataIndex: "metric" },
                  { title: "Value", dataIndex: "value" },
                ]}
                dataSource={POPOVER_ROWS}
                pagination={false}
                size="small"
              />
            }
            title="Telemetry"
            trigger="click"
          >
            <Button>Click for popover</Button>
          </Popover>
          <Dropdown
            menu={{
              items: [
                { key: "1", label: "Action one" },
                { key: "2", label: "Action two" },
                { type: "divider" },
                { key: "3", label: "Danger action", danger: true },
              ],
            }}
            trigger={["click"]}
          >
            <Button>Click for dropdown</Button>
          </Dropdown>
        </Space>
      </Section>

      <Section title="Static message shapes">
        <Space vertical size={8}>
          <Typography.Text type="secondary">
            (message/toast live under Components/Feedback — this page covers positioned overlays
            only)
          </Typography.Text>
        </Space>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <OverlaysPage />,
};

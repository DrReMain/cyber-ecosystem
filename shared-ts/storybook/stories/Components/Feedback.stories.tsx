import type { Meta, StoryObj } from "@storybook/react-vite";
import { Alert, Badge, Card, Progress, Result, Skeleton, Space, Spin, Tag, Typography } from "antd";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Components/Feedback",
};

export default meta;

type Story = StoryObj;

const PRESETS = [
  "magenta",
  "red",
  "volcano",
  "orange",
  "gold",
  "lime",
  "green",
  "cyan",
  "blue",
  "geekblue",
  "purple",
] as const;

function FeedbackPage() {
  return (
    <div style={{ maxWidth: 860 }}>
      <Section title="Alerts">
        <Space vertical size={12} style={{ display: "flex" }}>
          <Alert message="Success text" showIcon type="success" />
          <Alert message="Info text" showIcon type="info" />
          <Alert message="Warning text" showIcon type="warning" />
          <Alert message="Error text" showIcon type="error" />
          <Alert banner message="Banner variant — context downgrade" type="warning" />
        </Space>
      </Section>

      <Section
        description="Preset palette is reserved for tags and categorical color."
        title="Tags"
      >
        <Space size={8} wrap>
          {PRESETS.map((p) => (
            <Tag color={p} key={p}>
              {p}
            </Tag>
          ))}
          <Tag>default</Tag>
          <Tag variant="filled">borderless</Tag>
          <Tag closable>closable</Tag>
        </Space>
      </Section>

      <Section title="Badges & progress">
        <Space size={24} wrap align="center">
          <Badge count={5}>
            <div
              style={{
                background: "rgba(128,128,128,0.25)",
                borderRadius: 8,
                height: 40,
                width: 60,
              }}
            />
          </Badge>
          <Badge dot>
            <Typography.Text>Dot target</Typography.Text>
          </Badge>
          <Badge status="success" text="Success" />
          <Badge status="error" text="Error" />
          <Progress percent={62} style={{ width: 160 }} />
          <Progress percent={88} size="small" style={{ width: 120 }} type="circle" />
        </Space>
      </Section>

      <Section title="Loading & empty">
        <Space size={32} wrap align="start">
          <Spin tip="Loading…" />
          <Skeleton active paragraph={{ rows: 2 }} title style={{ width: 220 }} />
          <Result
            status="warning"
            style={{ padding: "12px 24px" }}
            subTitle="Something needs attention"
            title="Result warning"
          />
        </Space>
      </Section>

      <Section title="Card (container surface)">
        <Card size="small" style={{ width: 320 }} title="Card title">
          <p style={{ margin: 0 }}>
            Card body copy — the workhorse container surface with its header, body padding and
            shadow as the skin defines them.
          </p>
        </Card>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <FeedbackPage />,
};

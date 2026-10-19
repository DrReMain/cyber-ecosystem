import type { Meta, StoryObj } from "@storybook/react-vite";
import { Space, Typography } from "antd";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Foundation/Typography",
};

export default meta;

type Story = StoryObj;

function TypographyPage() {
  return (
    <div style={{ maxWidth: 760 }}>
      <Section title="Headings">
        <Space vertical>
          <Typography.Title level={1} style={{ margin: 0 }}>
            Heading 1 — console title
          </Typography.Title>
          <Typography.Title level={2} style={{ margin: 0 }}>
            Heading 2 — page title
          </Typography.Title>
          <Typography.Title level={3} style={{ margin: 0 }}>
            Heading 3 — section title
          </Typography.Title>
          <Typography.Title level={4} style={{ margin: 0 }}>
            Heading 4 — card title
          </Typography.Title>
          <Typography.Title level={5} style={{ margin: 0 }}>
            Heading 5 — minor title
          </Typography.Title>
        </Space>
      </Section>

      <Section title="Body">
        <Typography.Paragraph>
          Body copy at the skin&apos;s base size. Enterprise consoles trade legibility headroom for
          information density; this paragraph shows line height, weight and ink color as resolved by
          the active skin.
        </Typography.Paragraph>
        <Typography.Paragraph type="secondary">
          Secondary paragraph — supporting context that must recede.
        </Typography.Paragraph>
        <div>
          <Typography.Text type="success">Success inline text</Typography.Text>
          <span> · </span>
          <Typography.Text type="warning">Warning inline text</Typography.Text>
          <span> · </span>
          <Typography.Text type="danger">Danger inline text</Typography.Text>
        </div>
        <Typography.Paragraph disabled>Disabled paragraph</Typography.Paragraph>
      </Section>

      <Section title="Inline semantics">
        <Space vertical>
          <Typography.Text>
            Default with <Typography.Text code>inline code</Typography.Text> and{" "}
            <Typography.Text keyboard>⌘K</Typography.Text> and{" "}
            <Typography.Text mark>marked</Typography.Text> and{" "}
            <Typography.Text underline>underlined</Typography.Text> and{" "}
            <Typography.Text delete>deleted</Typography.Text>.
          </Typography.Text>
          <Typography.Text>
            A <Typography.Text strong>strong</Typography.Text> word, an{" "}
            <Typography.Text italic>italic</Typography.Text> word, and a{" "}
            <Typography.Link href="#">link</Typography.Link>.
          </Typography.Text>
        </Space>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <TypographyPage />,
};

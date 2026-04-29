import { Card, Flex, Typography } from "antd"

export const W = {
  input: 300,
  select: 200,
  textarea: 400,
  block: 400,
  cascader: 300,
  pseudo: 280,
} as const

export function Label({ children }: { children: React.ReactNode }) {
  return (
    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
      {children}
    </Typography.Text>
  )
}

export function Section({
  title,
  children,
  description,
}: {
  title: string
  children: React.ReactNode
  description?: string
}) {
  return (
    <div>
      <Typography.Text
        style={{
          display: "block",
          marginBottom: 4,
          fontSize: 13,
          fontWeight: 600,
          textTransform: "uppercase",
          letterSpacing: "0.05em",
        }}
      >
        {title}
      </Typography.Text>
      {description && (
        <Typography.Text
          type="secondary"
          style={{ display: "block", margin: "0 0 8px", fontSize: 12 }}
        >
          {description}
        </Typography.Text>
      )}
      {children}
    </div>
  )
}

const paragraphs = [
  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
  "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
  "Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.",
]

export function PageContent({ children }: { children?: React.ReactNode }) {
  return (
    <div>
      <Typography.Title level={4} style={{ marginBottom: 16 }}>
        Dashboard Overview
      </Typography.Title>
      <Typography.Paragraph type="secondary" style={{ maxWidth: 600 }}>
        Welcome back! Here is a summary of recent activity across your
        workspace. This page demonstrates how overlay components appear in
        context with realistic content.
      </Typography.Paragraph>
      <Flex vertical gap={16} style={{ maxWidth: 800 }}>
        <Card size="small">
          <Typography.Text strong>Recent Activity</Typography.Text>
          <Typography.Paragraph style={{ marginTop: 8, marginBottom: 0 }}>
            {paragraphs[0]}
          </Typography.Paragraph>
        </Card>
        <Card size="small">
          <Typography.Text strong>Team Updates</Typography.Text>
          <Typography.Paragraph style={{ marginTop: 8, marginBottom: 0 }}>
            {paragraphs[1]}
          </Typography.Paragraph>
        </Card>
        {children}
      </Flex>
    </div>
  )
}

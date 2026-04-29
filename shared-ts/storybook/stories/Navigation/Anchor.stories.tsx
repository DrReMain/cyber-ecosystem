import type { Meta, StoryObj } from "@storybook/react-vite"
import { Anchor, Card, Flex, Typography } from "antd"
import { expect, fn, within } from "storybook/test"
import { Section } from "../helpers"

const anchorItems = [
  { key: "1", href: "#section-1", title: "Section 1" },
  { key: "2", href: "#section-2", title: "Section 2" },
  {
    key: "3",
    href: "#section-3",
    title: "Section 3",
    children: [
      { key: "3-1", href: "#section-3-1", title: "Sub 3.1" },
      { key: "3-2", href: "#section-3-2", title: "Sub 3.2" },
    ],
  },
  { key: "4", href: "#section-4", title: "Section 4" },
]

const meta: Meta<typeof Anchor> = {
  title: "Antd/Navigation/Anchor",
  component: Anchor,
  parameters: { layout: "padded" },
  args: {
    affix: false,
    items: anchorItems,
  },
  argTypes: {
    direction: {
      control: "radio",
      options: ["vertical", "horizontal"],
    },
    affix: { control: "boolean" },
    showInkInFixed: { control: "boolean" },
    offsetTop: { control: "number" },
    onChange: { action: "changed" },
    onClick: { action: "clicked" },
  },
}

export default meta
type Story = StoryObj<typeof Anchor>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      {/* 1. Basic */}
      <Section title="Basic">
        <Anchor affix={false} items={anchorItems} />
      </Section>

      {/* 2. Horizontal */}
      <Section title="Horizontal">
        <Anchor affix={false} direction="horizontal" items={anchorItems} />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const ScrollablePage: Story = {
  parameters: { layout: "fullscreen", controls: { disable: true } },
  render: () => {
    const sections = [
      { id: "overview", title: "Overview" },
      { id: "features", title: "Features" },
      { id: "pricing", title: "Pricing" },
      { id: "faq", title: "FAQ" },
      { id: "contact", title: "Contact" },
    ]

    return (
      <Flex>
        <div
          style={{
            flex: 1,
            height: "100vh",
            overflowY: "auto",
            padding: "24px 32px",
          }}
        >
          <Typography.Title level={2}>Product Documentation</Typography.Title>
          {sections.map((section) => (
            <div
              key={section.id}
              id={section.id}
              style={{ minHeight: 300, marginBottom: 24 }}
            >
              <Typography.Title level={3}>{section.title}</Typography.Title>
              <Card>
                <Typography.Paragraph>
                  This is the {section.title.toLowerCase()} section. Scroll
                  through this page to see the Anchor component in the sidebar
                  highlight the current section automatically. Each section has
                  enough content to require scrolling.
                </Typography.Paragraph>
                <Typography.Paragraph type="secondary">
                  The Anchor component on the right side tracks your scroll
                  position and highlights the corresponding link. Click any link
                  to smoothly scroll to that section.
                </Typography.Paragraph>
              </Card>
            </div>
          ))}
        </div>
        <div
          style={{
            width: 160,
            height: "100vh",
            position: "sticky",
            top: 0,
            borderLeft: "1px solid var(--ant-color-border)",
            padding: "24px 0 24px 16px",
          }}
        >
          <Anchor
            offsetTop={24}
            items={sections.map((s) => ({
              key: s.id,
              href: `#${s.id}`,
              title: s.title,
            }))}
          />
        </div>
      </Flex>
    )
  },
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    items: anchorItems,
    onClick: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const link = canvas.getByText("Section 1")
    await link.click()
    await expect(args.onClick).toHaveBeenCalledTimes(1)
  },
}

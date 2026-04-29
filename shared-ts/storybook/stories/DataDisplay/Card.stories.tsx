import {
  EditOutlined,
  EllipsisOutlined,
  ShareAltOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Avatar, Button, Card, Flex, Space } from "antd"
import { expect, within } from "storybook/test"
import { borderedArg, sizeArg } from "../argTypes"
import { Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Card> = {
  title: "Antd/Data Display/Card",
  component: Card,
  parameters: { layout: "padded" },
  args: {
    title: "Card title",
    children: <p>Card content</p>,
  },
  argTypes: {
    variant: {
      control: "select",
      options: ["outlined", "borderless"],
    },
    size: sizeArg(["small", "medium"]),
    hoverable: { control: "boolean" },
    loading: { control: "boolean" },
    bordered: borderedArg,
  },
}

export default meta
type Story = StoryObj<typeof Card>

const variants: Array<React.ComponentProps<typeof Card>["variant"]> = [
  "outlined",
  "borderless",
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Variants" description="outlined (default) vs borderless">
        <Flex gap={16} wrap="wrap">
          {variants.map((variant) => (
            <Card
              key={variant}
              variant={variant}
              title={`${variant}`}
              style={{ width: 300 }}
            >
              <p>Card content with {variant} variant</p>
            </Card>
          ))}
        </Flex>
      </Section>

      <Section title="Hoverable" description="Lift-up shadow on hover">
        <Flex gap={16} wrap="wrap">
          <Card
            hoverable
            variant="outlined"
            title="Outlined + Hover"
            style={{ width: 300 }}
          >
            <p>Hover over this card to see the shadow effect.</p>
          </Card>
          <Card
            hoverable
            variant="borderless"
            title="Borderless + Hover"
            style={{ width: 300 }}
          >
            <p>Hover over this card to see the shadow effect.</p>
          </Card>
        </Flex>
      </Section>

      <Section title="Loading" description="Shows skeleton placeholder">
        <Flex gap={16} wrap="wrap">
          <Card loading style={{ width: 300 }}>
            <Card.Meta
              avatar={<Avatar>J</Avatar>}
              title="Card title"
              description="This is the description"
            />
          </Card>
          <Card loading variant="borderless" style={{ width: 300 }}>
            <Card.Meta
              avatar={<Avatar>J</Avatar>}
              title="Card title"
              description="This is the description"
            />
          </Card>
        </Flex>
      </Section>

      <Section title="With Actions" description="Action bar at bottom of card">
        <Card
          title="Card with Actions"
          style={{ width: 360 }}
          actions={[
            <Space key="edit">
              <EditOutlined /> Edit
            </Space>,
            <Space key="share">
              <ShareAltOutlined /> Share
            </Space>,
            <Space key="more">
              <EllipsisOutlined /> More
            </Space>,
          ]}
        >
          <p>Card content with action bar at bottom.</p>
        </Card>
      </Section>

      <Section title="Inner Card" description="Nested card with type='inner'">
        <Card title="Outer Card" style={{ width: 400 }}>
          <p>Outer content</p>
          <Card type="inner" title="Inner Card">
            <p>Inner card content</p>
          </Card>
        </Card>
      </Section>

      <Section title="With Tabs" description="Tabbed card header using tabList">
        <Card
          title="Card with Tabs"
          style={{ width: 400 }}
          tabList={[
            { key: "tab1", label: "Tab 1" },
            { key: "tab2", label: "Tab 2" },
          ]}
          defaultActiveTabKey="tab1"
        >
          <p>Switch tabs in the card header.</p>
        </Card>
      </Section>

      <Section title="Card Grid" description="Grid layout within a card">
        <Card title="Card Grid" style={{ width: 480 }}>
          <Card.Grid style={{ width: "50%" }}>Content A</Card.Grid>
          <Card.Grid style={{ width: "50%" }}>Content B</Card.Grid>
          <Card.Grid style={{ width: "50%" }} hoverable={false}>
            Not hoverable
          </Card.Grid>
          <Card.Grid style={{ width: "50%" }}>Content D</Card.Grid>
        </Card>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Card size="small" style={{ width: 200 }}>
            Card content
          </Card>
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { style: { width: 360 } },
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    title: "Interactive Card",
    style: { width: 360 },
  },
  render: (args) => (
    <Card
      {...args}
      actions={[
        <Button key="edit" type="text">
          Edit
        </Button>,
        <Button key="share" type="text">
          Share
        </Button>,
      ]}
    >
      <p>Click the actions below to interact.</p>
    </Card>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const editBtn = canvas.getByText("Edit")
    await editBtn.click()
    await expect(canvas.getByText("Share")).toBeInTheDocument()
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  Button,
  Descriptions,
  Drawer,
  Flex,
  Form,
  Input,
  Select,
  Space,
  Tag,
  Typography,
} from "antd"
import { useState } from "react"
import { expect, fn, within } from "storybook/test"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof Drawer> = {
  title: "Antd/Feedback/Drawer",
  component: Drawer,
  parameters: { layout: "padded" },
  args: {
    open: true,
    title: "Drawer Title",
    placement: "right",
    width: 378,
    closable: true,
    mask: true,
  },
  argTypes: {
    open: { control: "boolean" },
    title: { control: "text" },
    placement: {
      control: "radio",
      options: ["left", "right", "top", "bottom"],
    },
    width: { control: "number" },
    closable: { control: "boolean" },
    mask: { control: "boolean" },
    onClose: { action: "closed" },
  },
}

export default meta
type Story = StoryObj<typeof Drawer>

// ── Gallery ──────────────────────────────────────────────────────

const { Text } = Typography

function FormDrawerDemo() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant="solid" color="primary" onClick={() => setOpen(true)}>
        Open Form Drawer
      </Button>
      <Drawer
        title="Edit User"
        placement="right"
        width={400}
        open={open}
        onClose={() => setOpen(false)}
        extra={
          <Space>
            <Button onClick={() => setOpen(false)}>Cancel</Button>
            <Button
              variant="solid"
              color="primary"
              onClick={() => setOpen(false)}
            >
              Save
            </Button>
          </Space>
        }
      >
        <Form layout="vertical">
          <Form.Item label="Full Name">
            <Input placeholder="Enter full name" defaultValue="Alice Johnson" />
          </Form.Item>
          <Form.Item label="Email">
            <Input placeholder="Enter email" defaultValue="alice@example.com" />
          </Form.Item>
          <Form.Item label="Department">
            <Select
              defaultValue="engineering"
              options={[
                { value: "engineering", label: "Engineering" },
                { value: "design", label: "Design" },
                { value: "marketing", label: "Marketing" },
              ]}
            />
          </Form.Item>
          <Form.Item label="Bio">
            <Input.TextArea rows={4} placeholder="Tell us about yourself" />
          </Form.Item>
        </Form>
      </Drawer>
    </>
  )
}

function DetailDrawerDemo() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open Detail Drawer</Button>
      <Drawer
        title="User Details"
        placement="right"
        width={420}
        open={open}
        onClose={() => setOpen(false)}
      >
        <Flex vertical gap={16}>
          <Flex align="center" gap={12}>
            <Avatar size={48}>AJ</Avatar>
            <Flex vertical>
              <Text strong>Alice Johnson</Text>
              <Text type="secondary">alice@example.com</Text>
            </Flex>
          </Flex>
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="Department">
              Engineering
            </Descriptions.Item>
            <Descriptions.Item label="Role">Senior Developer</Descriptions.Item>
            <Descriptions.Item label="Status">
              <Tag color="success">Active</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Joined">Jan 15, 2024</Descriptions.Item>
          </Descriptions>
        </Flex>
      </Drawer>
    </>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [placementOpen, setPlacementOpen] = useState(false)
    const [placement, setPlacement] = useState<
      "left" | "right" | "top" | "bottom"
    >("right")
    const [sizeOpen, setSizeOpen] = useState(false)
    const [size, setSize] = useState<"default" | "large">("default")
    const [footerOpen, setFooterOpen] = useState(false)
    const [loadingOpen, setLoadingOpen] = useState(false)
    const [noMaskOpen, setNoMaskOpen] = useState(false)

    return (
      <Flex vertical gap={24}>
        <Section title="Placements">
          <Space>
            <Label>
              {placement.charAt(0).toUpperCase() + placement.slice(1)}
            </Label>
            {(["left", "right", "top", "bottom"] as const).map((p) => (
              <Button
                key={p}
                type={placement === p ? "primary" : "default"}
                onClick={() => setPlacement(p)}
              >
                {p.charAt(0).toUpperCase() + p.slice(1)}
              </Button>
            ))}
            <Button type="primary" onClick={() => setPlacementOpen(true)}>
              Open
            </Button>
          </Space>
          <Drawer
            title={`Drawer (${placement})`}
            placement={placement}
            open={placementOpen}
            onClose={() => setPlacementOpen(false)}
          >
            <p>Drawer content with {placement} placement.</p>
          </Drawer>
        </Section>

        <Section title="Sizes">
          <Space>
            <Button
              type={size === "default" ? "primary" : "default"}
              onClick={() => setSize("default")}
            >
              Default (378px)
            </Button>
            <Button
              type={size === "large" ? "primary" : "default"}
              onClick={() => setSize("large")}
            >
              Large (736px)
            </Button>
            <Button type="primary" onClick={() => setSizeOpen(true)}>
              Open
            </Button>
          </Space>
          <Drawer
            title={`Drawer size: ${size}`}
            size={size}
            open={sizeOpen}
            onClose={() => setSizeOpen(false)}
          >
            <p>Content area with {size} size.</p>
          </Drawer>
        </Section>

        <Section title="With Footer">
          <Button type="primary" onClick={() => setFooterOpen(true)}>
            Open with Footer
          </Button>
          <Drawer
            title="Drawer with Footer"
            open={footerOpen}
            onClose={() => setFooterOpen(false)}
            footer={
              <Space style={{ display: "flex", justifyContent: "flex-end" }}>
                <Button onClick={() => setFooterOpen(false)}>Cancel</Button>
                <Button type="primary" onClick={() => setFooterOpen(false)}>
                  Submit
                </Button>
              </Space>
            }
          >
            <p>Drawer with a footer action bar.</p>
          </Drawer>
        </Section>

        <Section title="Loading">
          <Button type="primary" onClick={() => setLoadingOpen(true)}>
            Open Loading Drawer
          </Button>
          <Drawer
            title="Loading"
            open={loadingOpen}
            onClose={() => setLoadingOpen(false)}
            loading
          >
            <p>This content is behind a skeleton loading overlay.</p>
          </Drawer>
        </Section>

        <Section title="No Mask">
          <Button type="primary" onClick={() => setNoMaskOpen(true)}>
            Open No-Mask Drawer
          </Button>
          <Drawer
            title="No Mask"
            open={noMaskOpen}
            onClose={() => setNoMaskOpen(false)}
            mask={false}
          >
            <p>Drawer without a background mask overlay.</p>
          </Drawer>
        </Section>

        <Section
          title="Form Drawer"
          description="Drawer containing an edit form with save/cancel footer"
        >
          <FormDrawerDemo />
        </Section>

        <Section
          title="Detail Drawer"
          description="Drawer displaying a read-only record detail view"
        >
          <DetailDrawerDemo />
        </Section>
      </Flex>
    )
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  render: (args) => (
    <div style={{ position: "relative", height: 400, overflow: "hidden" }}>
      <Drawer {...args} getContainer={false} onClose={() => {}}>
        <p>Some contents...</p>
        <p>Some contents...</p>
        <p>Some contents...</p>
      </Drawer>
    </div>
  ),
}

// ── StaticOpen ───────────────────────────────────────────────────

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div style={{ position: "relative", minHeight: 500, overflow: "hidden" }}>
      <PageContent />
      <Drawer title="Basic Drawer" open getContainer={false} onClose={() => {}}>
        <p>Some contents...</p>
        <p>Some contents...</p>
        <p>Some contents...</p>
      </Drawer>
    </div>
  ),
}

// ── Interactive story with play function ──────────────────────────

export const Interactive: Story = {
  args: {
    title: "Interactive Drawer",
    onClose: fn(),
  },
  render: (args) => {
    function DrawerDemo() {
      const [open, setOpen] = useState(false)
      return (
        <>
          <Button type="primary" onClick={() => setOpen(true)}>
            Open Drawer
          </Button>
          <Drawer
            {...args}
            open={open}
            onClose={(e) => {
              setOpen(false)
              args.onClose?.(e)
            }}
          >
            <p>Drawer content</p>
          </Drawer>
        </>
      )
    }
    return <DrawerDemo />
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole("button", { name: "Open Drawer" })
    await button.click()
    await expect(args.onClose).not.toHaveBeenCalled()
  },
}

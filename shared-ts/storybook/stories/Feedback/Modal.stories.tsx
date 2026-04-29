import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Form, Input, Modal, Select, Space } from "antd"
import { useState } from "react"
import { expect, fn, within } from "storybook/test"
import { PageContent, Section } from "../helpers"

const meta: Meta<typeof Modal> = {
  title: "Antd/Feedback/Modal",
  component: Modal,
  parameters: { layout: "padded" },
  args: {
    open: true,
    title: "Modal Title",
    width: 520,
    centered: false,
    closable: true,
  },
  argTypes: {
    open: { control: "boolean" },
    title: { control: "text" },
    width: { control: "number" },
    centered: { control: "boolean" },
    closable: { control: "boolean" },
    onCancel: { action: "cancelled" },
    onOk: { action: "ok" },
  },
}

export default meta
type Story = StoryObj<typeof Modal>

// ── Gallery ──────────────────────────────────────────────────────

function WidthModal({ width }: { width: number }) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button onClick={() => setOpen(true)}>{width}px</Button>
      <Modal
        title={`Modal width: ${width}px`}
        open={open}
        width={width}
        onOk={() => setOpen(false)}
        onCancel={() => setOpen(false)}
      >
        <p>Modal content with width {width}px.</p>
      </Modal>
    </>
  )
}

function FormDialogDemo() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant="solid" color="primary" onClick={() => setOpen(true)}>
        Open Form Dialog
      </Button>
      <Modal
        title="Create New User"
        open={open}
        onOk={() => setOpen(false)}
        onCancel={() => setOpen(false)}
        okText="Create"
      >
        <Form layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            label="Username"
            required
            validateStatus="error"
            help="Username is required"
          >
            <Input placeholder="Enter username" />
          </Form.Item>
          <Form.Item label="Email" required>
            <Input placeholder="Enter email" type="email" />
          </Form.Item>
          <Form.Item label="Role">
            <Select
              placeholder="Select role"
              options={[
                { value: "admin", label: "Admin" },
                { value: "editor", label: "Editor" },
                { value: "viewer", label: "Viewer" },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [centeredOpen, setCenteredOpen] = useState(false)
    const [customFooterOpen, setCustomFooterOpen] = useState(false)
    const [noFooterOpen, setNoFooterOpen] = useState(false)
    const [confirmLoadingOpen, setConfirmLoadingOpen] = useState(false)
    const [loading, setLoading] = useState(false)
    const [loadingSkeletonOpen, setLoadingSkeletonOpen] = useState(false)

    const handleOk = () => {
      setLoading(true)
      setTimeout(() => {
        setLoading(false)
        setConfirmLoadingOpen(false)
      }, 2000)
    }

    return (
      <Flex vertical gap={24}>
        <Section title="Centered">
          <Button type="primary" onClick={() => setCenteredOpen(true)}>
            Open Centered Modal
          </Button>
          <Modal
            title="Centered Modal"
            centered
            open={centeredOpen}
            onOk={() => setCenteredOpen(false)}
            onCancel={() => setCenteredOpen(false)}
          >
            <p>This modal is vertically centered.</p>
          </Modal>
        </Section>

        <Section title="Custom Footer">
          <Button type="primary" onClick={() => setCustomFooterOpen(true)}>
            Open Custom Footer
          </Button>
          <Modal
            title="Custom Footer"
            open={customFooterOpen}
            footer={(_, { OkBtn, CancelBtn }) => (
              <Flex justify="space-between">
                <Button onClick={() => setCustomFooterOpen(false)}>
                  Do Something
                </Button>
                <Space>
                  <CancelBtn />
                  <OkBtn />
                </Space>
              </Flex>
            )}
          >
            <p>This modal has a custom footer layout.</p>
          </Modal>
        </Section>

        <Section title="No Footer">
          <Button type="primary" onClick={() => setNoFooterOpen(true)}>
            Open No-Footer Modal
          </Button>
          <Modal
            title="No Footer"
            open={noFooterOpen}
            onCancel={() => setNoFooterOpen(false)}
            footer={null}
          >
            <p>This modal has no footer buttons.</p>
          </Modal>
        </Section>

        <Section title="Confirm Loading">
          <Button type="primary" onClick={() => setConfirmLoadingOpen(true)}>
            Open with Loading
          </Button>
          <Modal
            title="Confirm Loading"
            open={confirmLoadingOpen}
            onOk={handleOk}
            onCancel={() => setConfirmLoadingOpen(false)}
            confirmLoading={loading}
          >
            <p>Click OK to trigger a 2-second loading state.</p>
          </Modal>
        </Section>

        <Section title="Loading Skeleton">
          <Button type="primary" onClick={() => setLoadingSkeletonOpen(true)}>
            Open Loading Modal
          </Button>
          <Modal
            title="Loading Skeleton"
            open={loadingSkeletonOpen}
            onOk={() => setLoadingSkeletonOpen(false)}
            onCancel={() => setLoadingSkeletonOpen(false)}
            loading
          >
            <p>Content is hidden behind skeleton.</p>
          </Modal>
        </Section>

        <Section title="Width Variants">
          <Space wrap>
            {[400, 520, 720, 1000].map((w) => (
              <WidthModal key={w} width={w} />
            ))}
          </Space>
        </Section>

        <Section
          title="Form Dialog"
          description="The most common modal usage — form with validation"
        >
          <FormDialogDemo />
        </Section>
      </Flex>
    )
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {
  render: (args) => (
    <div style={{ position: "relative", height: 400, overflow: "hidden" }}>
      <Modal {...args} getContainer={false} onOk={() => {}} onCancel={() => {}}>
        <p>Some contents...</p>
        <p>Some contents...</p>
        <p>Some contents...</p>
      </Modal>
    </div>
  ),
}

// ── StaticOpen ───────────────────────────────────────────────────

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div style={{ height: 500, overflow: "hidden", position: "relative" }}>
      <PageContent />
      <Modal
        title="Static Modal"
        open
        getContainer={false}
        onOk={() => {}}
        onCancel={() => {}}
      >
        <p>This modal is always visible for skin inspection.</p>
        <p>Header, body, footer, close button, and overlay are all rendered.</p>
      </Modal>
    </div>
  ),
}

// ── Interactive story with play function ──────────────────────────

export const Interactive: Story = {
  args: {
    title: "Interactive Modal",
    onCancel: fn(),
    onOk: fn(),
  },
  render: (args) => {
    function ModalDemo() {
      const [open, setOpen] = useState(false)
      return (
        <>
          <Button type="primary" onClick={() => setOpen(true)}>
            Open Modal
          </Button>
          <Modal
            {...args}
            open={open}
            onCancel={(e) => {
              setOpen(false)
              args.onCancel?.(e)
            }}
            onOk={(e) => {
              setOpen(false)
              args.onOk?.(e)
            }}
          >
            <p>Modal content</p>
          </Modal>
        </>
      )
    }
    return <ModalDemo />
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole("button", { name: "Open Modal" })
    await button.click()
    await expect(args.onCancel).not.toHaveBeenCalled()
    await expect(args.onOk).not.toHaveBeenCalled()
  },
}

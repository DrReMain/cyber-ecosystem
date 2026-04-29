import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Button,
  DatePicker,
  Flex,
  Form,
  Input,
  Modal,
  Select,
  Space,
} from "antd"
import { useState } from "react"
import { expect, fn, userEvent, within } from "storybook/test"

const meta: Meta<typeof Modal> = {
  title: "Antd/Combinations/FormModal",
  component: Modal,
  parameters: { layout: "centered" },
}

export default meta
type Story = StoryObj<typeof Modal>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Modal
      open
      title="Create New Project"
      footer={
        <Space>
          <Button>Cancel</Button>
          <Button type="primary" variant="solid">
            Create
          </Button>
        </Space>
      }
    >
      <Form layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item label="Project Name" required>
          <Input placeholder="Enter project name" />
        </Form.Item>
        <Form.Item label="Description">
          <Input.TextArea placeholder="Brief description" rows={3} />
        </Form.Item>
        <Flex gap={16}>
          <Form.Item label="Team" style={{ flex: 1 }}>
            <Select
              placeholder="Select team"
              options={[
                { value: "eng", label: "Engineering" },
                { value: "design", label: "Design" },
                { value: "product", label: "Product" },
              ]}
            />
          </Form.Item>
          <Form.Item label="Priority" style={{ flex: 1 }}>
            <Select
              placeholder="Select priority"
              options={[
                { value: "low", label: "Low" },
                { value: "medium", label: "Medium" },
                { value: "high", label: "High" },
              ]}
            />
          </Form.Item>
        </Flex>
        <Form.Item label="Due Date">
          <DatePicker style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item label="Tags">
          <Select
            mode="tags"
            placeholder="Add tags"
            options={[
              { value: "frontend", label: "Frontend" },
              { value: "backend", label: "Backend" },
              { value: "infra", label: "Infra" },
            ]}
          />
        </Form.Item>
      </Form>
    </Modal>
  ),
}

export const Playground: Story = {
  render: () => (
    <Modal
      open
      title="Create New Project"
      footer={
        <Space>
          <Button>Cancel</Button>
          <Button type="primary" variant="solid">
            Create
          </Button>
        </Space>
      }
    >
      <Form layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item label="Project Name" required>
          <Input placeholder="Enter project name" />
        </Form.Item>
        <Form.Item label="Description">
          <Input.TextArea placeholder="Brief description" rows={3} />
        </Form.Item>
        <Flex gap={16}>
          <Form.Item label="Team" style={{ flex: 1 }}>
            <Select
              placeholder="Select team"
              options={[
                { value: "eng", label: "Engineering" },
                { value: "design", label: "Design" },
              ]}
            />
          </Form.Item>
          <Form.Item label="Due Date" style={{ flex: 1 }}>
            <DatePicker style={{ width: "100%" }} />
          </Form.Item>
        </Flex>
      </Form>
    </Modal>
  ),
}

const interactiveOnFinish = fn()
const interactiveOnCancel = fn()

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    onCancel: fn(),
  },
  render: (args) => {
    function FormModalDemo() {
      const [open, setOpen] = useState(false)
      return (
        <>
          <Button type="primary" onClick={() => setOpen(true)}>
            Open Modal
          </Button>
          <Modal
            open={open}
            title="Create New Project"
            onCancel={(e) => {
              setOpen(false)
              interactiveOnCancel(e)
              args.onCancel?.(e)
            }}
            footer={
              <Space>
                <Button
                  onClick={() => {
                    setOpen(false)
                    interactiveOnCancel()
                  }}
                >
                  Cancel
                </Button>
                <Button
                  type="primary"
                  variant="solid"
                  onClick={() => setOpen(false)}
                  form="interactive-form"
                  htmlType="submit"
                >
                  OK
                </Button>
              </Space>
            }
          >
            <Form
              id="interactive-form"
              layout="vertical"
              style={{ marginTop: 16 }}
              onFinish={() => {
                setOpen(false)
                interactiveOnFinish()
              }}
            >
              <Form.Item
                label="Project Name"
                name="name"
                rules={[{ required: true }]}
              >
                <Input placeholder="Enter project name" />
              </Form.Item>
              <Form.Item label="Description" name="description">
                <Input.TextArea placeholder="Brief description" rows={3} />
              </Form.Item>
            </Form>
          </Modal>
        </>
      )
    }
    return <FormModalDemo />
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const openBtn = canvas.getByRole("button", { name: "Open Modal" })
    await openBtn.click()
    const modal = within(document.body)
    const nameInput = modal.getByPlaceholderText("Enter project name")
    await userEvent.type(nameInput, "My Project")
    const okBtn = modal.getByRole("button", { name: "OK" })
    await okBtn.click()
    await expect(interactiveOnFinish).toHaveBeenCalled()
  },
}

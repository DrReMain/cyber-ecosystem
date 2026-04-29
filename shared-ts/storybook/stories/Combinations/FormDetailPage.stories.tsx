import { SaveOutlined, UndoOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Breadcrumb,
  Button,
  Card,
  Checkbox,
  DatePicker,
  Flex,
  Form,
  Input,
  InputNumber,
  Radio,
  Select,
  Slider,
  Space,
  Steps,
  Tag,
  Typography,
} from "antd"
import { useState } from "react"

const meta: Meta = {
  title: "Antd/Combinations/FormDetailPage",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

interface FormValues {
  name: string
  description: string
  category: string
  priority: string
  status: string
  budget: number
  startDate: unknown
  endDate: unknown
  tags: string[]
  notify: boolean
  publicAccess: boolean
  members: string[]
  progress: number
}

function FormDetailPage({
  validationMode,
  readOnly,
}: {
  validationMode?: "error" | "warning" | "none"
  readOnly?: boolean
}) {
  const [form] = Form.useForm<FormValues>()
  const [currentStep, setCurrentStep] = useState(0)

  const priorityOptions = [
    { label: "Low", value: "low" },
    { label: "Medium", value: "medium" },
    { label: "High", value: "high" },
  ]

  const categoryOptions = [
    { label: "Engineering", value: "engineering" },
    { label: "Design", value: "design" },
    { label: "Marketing", value: "marketing" },
    { label: "Product", value: "product" },
  ]

  const tagOptions = [
    { value: "frontend", label: "Frontend" },
    { value: "backend", label: "Backend" },
    { value: "infra", label: "Infrastructure" },
    { value: "mobile", label: "Mobile" },
    { value: "ai", label: "AI/ML" },
  ]

  const memberOptions = [
    { value: "alice", label: "Alice Chen" },
    { value: "bob", label: "Bob Smith" },
    { value: "carol", label: "Carol Lee" },
    { value: "david", label: "David Wang" },
    { value: "eve", label: "Eve Park" },
  ]

  const validateStatus = validationMode === "none" ? undefined : validationMode

  return (
    <Flex vertical gap={16} style={{ maxWidth: 900 }}>
      {/* Page Header */}
      <Flex justify="space-between" align="center" wrap gap={8}>
        <Flex vertical gap={4}>
          <Breadcrumb
            items={[
              { title: "Workspace" },
              { title: "Projects" },
              { title: "Edit Project" },
            ]}
          />
          <Typography.Title level={4} style={{ margin: 0 }}>
            Project Details
          </Typography.Title>
        </Flex>
        <Space>
          <Button icon={<UndoOutlined />}>Reset</Button>
          <Button type="primary" variant="solid" icon={<SaveOutlined />}>
            Save Changes
          </Button>
        </Space>
      </Flex>

      {/* Steps Wizard */}
      <Card size="small">
        <Steps
          current={currentStep}
          onChange={setCurrentStep}
          items={[
            { title: "Basic Info" },
            { title: "Configuration" },
            { title: "Members" },
          ]}
        />
      </Card>

      <Form
        form={form}
        layout="vertical"
        disabled={readOnly}
        initialValues={{
          name: "Project Alpha",
          description:
            "A comprehensive platform upgrade initiative focusing on performance optimization and user experience improvements.",
          category: "engineering",
          priority: "high",
          status: "active",
          budget: 120000,
          tags: ["frontend", "backend"],
          notify: true,
          publicAccess: false,
          members: ["alice", "bob"],
          progress: 65,
        }}
      >
        {/* Basic Information */}
        <Card
          size="small"
          title={<Typography.Text strong>Basic Information</Typography.Text>}
          style={{ marginBottom: 16 }}
        >
          <Flex gap={16} wrap>
            <Form.Item
              label="Project Name"
              name="name"
              style={{ flex: 1, minWidth: 280 }}
              validateStatus={validateStatus}
              help={
                validationMode === "error"
                  ? "Project name is required"
                  : validationMode === "warning"
                    ? "This name may conflict with existing projects"
                    : undefined
              }
              rules={[{ required: true }]}
            >
              <Input placeholder="Enter project name" />
            </Form.Item>
            <Form.Item
              label="Category"
              name="category"
              style={{ flex: 1, minWidth: 280 }}
            >
              <Select options={categoryOptions} placeholder="Select category" />
            </Form.Item>
          </Flex>

          <Form.Item label="Description" name="description">
            <Input.TextArea
              placeholder="Brief description of the project"
              rows={3}
            />
          </Form.Item>

          <Flex gap={16} wrap>
            <Form.Item
              label="Priority"
              name="priority"
              style={{ flex: 1, minWidth: 280 }}
            >
              <Radio.Group options={priorityOptions} optionType="button" />
            </Form.Item>
            <Form.Item
              label="Status"
              name="status"
              style={{ flex: 1, minWidth: 280 }}
            >
              <Select
                options={[
                  { label: "Active", value: "active" },
                  { label: "Pending", value: "pending" },
                  { label: "Completed", value: "completed" },
                  { label: "On Hold", value: "on-hold" },
                ]}
              />
            </Form.Item>
          </Flex>

          <Flex gap={16} wrap>
            <Form.Item
              label="Budget"
              name="budget"
              style={{ flex: 1, minWidth: 280 }}
            >
              <InputNumber
                style={{ width: "100%" }}
                prefix="$"
                formatter={(value) =>
                  `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                }
                parser={(value) =>
                  value?.replace(/\$\s?|(,*)/g, "") as unknown as number
                }
              />
            </Form.Item>
            <Form.Item
              label="Timeline"
              style={{ flex: 1, minWidth: 280, marginBottom: 0 }}
            >
              <Flex gap={8}>
                <Form.Item name="startDate" noStyle>
                  <DatePicker
                    style={{ width: "100%" }}
                    placeholder="Start date"
                  />
                </Form.Item>
                <Form.Item name="endDate" noStyle>
                  <DatePicker
                    style={{ width: "100%" }}
                    placeholder="End date"
                  />
                </Form.Item>
              </Flex>
            </Form.Item>
          </Flex>
        </Card>

        {/* Configuration */}
        <Card
          size="small"
          title={<Typography.Text strong>Configuration</Typography.Text>}
          style={{ marginBottom: 16 }}
        >
          <Flex vertical gap={16}>
            <Flex gap={24} wrap>
              <Form.Item
                name="notify"
                valuePropName="checked"
                style={{ marginBottom: 0 }}
              >
                <Checkbox>Enable email notifications</Checkbox>
              </Form.Item>
              <Form.Item
                name="publicAccess"
                valuePropName="checked"
                style={{ marginBottom: 0 }}
              >
                <Checkbox>Allow public access</Checkbox>
              </Form.Item>
            </Flex>

            <Form.Item label="Progress" name="progress">
              <Slider
                marks={{
                  0: "0%",
                  25: "25%",
                  50: "50%",
                  75: "75%",
                  100: "100%",
                }}
              />
            </Form.Item>

            <Form.Item label="Tags" name="tags">
              <Select
                mode="tags"
                placeholder="Add tags"
                options={tagOptions}
                tagRender={({ label, closable, onClose }) => (
                  <Tag closable={closable} onClose={onClose}>
                    {label}
                  </Tag>
                )}
              />
            </Form.Item>
          </Flex>
        </Card>

        {/* Members */}
        <Card
          size="small"
          title={<Typography.Text strong>Team Members</Typography.Text>}
        >
          <Form.Item
            label="Assigned Members"
            name="members"
            validateStatus={validateStatus}
            help={
              validationMode === "error"
                ? "At least one member must be assigned"
                : undefined
            }
          >
            <Select
              mode="multiple"
              placeholder="Select team members"
              options={memberOptions}
              style={{ width: "100%" }}
            />
          </Form.Item>
        </Card>
      </Form>
    </Flex>
  )
}

export const Normal: Story = {
  parameters: { controls: { disable: true } },
  render: () => <FormDetailPage />,
}

export const ValidationError: Story = {
  parameters: { controls: { disable: true } },
  render: () => <FormDetailPage validationMode="error" />,
}

export const ValidationWarning: Story = {
  parameters: { controls: { disable: true } },
  render: () => <FormDetailPage validationMode="warning" />,
}

export const ReadOnly: Story = {
  parameters: { controls: { disable: true } },
  render: () => <FormDetailPage readOnly />,
}

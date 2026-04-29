import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Button,
  Flex,
  Input,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from "antd"
import { expect, fn, within } from "storybook/test"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Table> = {
  title: "Antd/Data Display/Table",
  component: Table,
  parameters: { layout: "padded" },
  args: {
    bordered: false,
    loading: false,
    size: "medium",
    showHeader: true,
  },
  argTypes: {
    size: sizeArg(["large", "medium", "small"]),
    bordered: { control: "boolean" },
    loading: { control: "boolean" },
    showHeader: { control: "boolean" },
    sticky: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Table>

interface DataType {
  key: string
  name: string
  age: number
  address: string
  tags: string[]
  description?: string
}

const data: DataType[] = [
  {
    key: "1",
    name: "John Brown",
    age: 32,
    address: "New York No. 1 Lake Park",
    tags: ["nice", "developer"],
    description: "My name is John Brown, I am 32 years old.",
  },
  {
    key: "2",
    name: "Jim Green",
    age: 42,
    address: "London No. 1 Lake Park",
    tags: ["loser"],
    description: "My name is Jim Green, I am 42 years old.",
  },
  {
    key: "3",
    name: "Joe Black",
    age: 28,
    address: "Sydney No. 1 Lake Park",
    tags: ["cool", "teacher"],
    description: "My name is Joe Black, I am 28 years old.",
  },
  {
    key: "4",
    name: "Disabled User",
    age: 35,
    address: "Paris No. 1 Lake Park",
    tags: ["disabled"],
    description: "This user is disabled for selection.",
  },
  {
    key: "5",
    name: "Jane Doe",
    age: 25,
    address: "Berlin No. 1 Lake Park",
    tags: ["cool"],
  },
  {
    key: "6",
    name: "Tom Smith",
    age: 38,
    address: "Tokyo No. 1 Lake Park",
    tags: ["developer"],
  },
  {
    key: "7",
    name: "Alice Wang",
    age: 30,
    address: "Seoul No. 1 Lake Park",
    tags: ["teacher"],
  },
  {
    key: "8",
    name: "Bob Lee",
    age: 45,
    address: "Shanghai No. 1 Lake Park",
    tags: ["nice", "developer"],
  },
]

const basicColumns = [
  { title: "Name", dataIndex: "name", key: "name" },
  { title: "Age", dataIndex: "age", key: "age" },
  { title: "Address", dataIndex: "address", key: "address" },
]

const sizes = ["large", "medium", "small"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes (large / medium / small)">
        <Flex vertical gap={16}>
          {sizes.map((size) => (
            <div key={size}>
              <Label>{size}</Label>
              <Table
                columns={basicColumns}
                dataSource={data.slice(0, 3)}
                size={size}
              />
            </div>
          ))}
        </Flex>
      </Section>

      <Section title="Bordered">
        <Flex vertical gap={16}>
          {sizes.map((size) => (
            <div key={`bordered-${size}`}>
              <Label>{size}</Label>
              <Table
                columns={basicColumns}
                dataSource={data.slice(0, 3)}
                size={size}
                bordered
              />
            </div>
          ))}
        </Flex>
      </Section>

      <Section title="Row Selection">
        <Flex vertical gap={16}>
          <div>
            <Label>Checkbox</Label>
            <Table
              columns={basicColumns}
              dataSource={data}
              rowSelection={{
                type: "checkbox",
                getCheckboxProps: (record) => ({
                  disabled: record.name === "Disabled User",
                }),
                selections: true,
              }}
              pagination={{ pageSize: 4 }}
            />
          </div>
          <div>
            <Label>Radio</Label>
            <Table
              columns={basicColumns}
              dataSource={data.slice(0, 4)}
              rowSelection={{
                type: "radio",
              }}
            />
          </div>
        </Flex>
      </Section>

      <Section title="Expandable Rows">
        <Table
          columns={basicColumns}
          dataSource={data.slice(0, 4)}
          expandable={{
            expandedRowRender: (record) => (
              <p style={{ margin: 0 }}>
                {record.description ?? "No description"}
              </p>
            ),
            rowExpandable: (record) => !!record.description,
          }}
        />
      </Section>

      <Section title="Loading">
        <Flex vertical gap={16}>
          <div>
            <Label>Loading state</Label>
            <Table columns={basicColumns} dataSource={[]} loading />
          </div>
          <div>
            <Label>Custom loading (with tip)</Label>
            <Table
              columns={basicColumns}
              dataSource={[]}
              loading={{ tip: "Loading data...", spinning: true }}
            />
          </div>
        </Flex>
      </Section>

      <Section title="Custom Render & Tags">
        <Table
          columns={[
            { title: "Name", dataIndex: "name", key: "name" },
            { title: "Age", dataIndex: "age", key: "age" },
            { title: "Address", dataIndex: "address", key: "address" },
            {
              title: "Tags",
              key: "tags",
              dataIndex: "tags",
              render: (tags: string[]) => (
                <>
                  {tags.map((tag) => (
                    <Tag
                      color={
                        tag === "loser"
                          ? "volcano"
                          : tag === "cool"
                            ? "green"
                            : "blue"
                      }
                      key={tag}
                    >
                      {tag.toUpperCase()}
                    </Tag>
                  ))}
                </>
              ),
            },
            {
              title: "Action",
              key: "action",
              render: (_: unknown, record: DataType) => (
                <Space size="middle">
                  <Typography.Link>Invite {record.name}</Typography.Link>
                  <Typography.Link>Delete</Typography.Link>
                </Space>
              ),
            },
          ]}
          dataSource={data.slice(0, 4)}
        />
      </Section>

      <Section title="States">
        <PseudoStates>
          <Table
            columns={basicColumns}
            dataSource={data.slice(0, 1)}
            pagination={false}
            size="small"
          />
        </PseudoStates>
      </Section>

      <Section
        title="Data Management"
        description="Realistic table with search, filter, and action toolbar"
      >
        <Flex vertical gap={12}>
          <Flex gap={8} wrap>
            <Input.Search
              placeholder="Search by name..."
              style={{ width: 250 }}
              allowClear
            />
            <Select
              placeholder="Filter by tag"
              style={{ width: 150 }}
              allowClear
              options={[
                { value: "nice", label: "Nice" },
                { value: "developer", label: "Developer" },
                { value: "loser", label: "Loser" },
                { value: "cool", label: "Cool" },
                { value: "teacher", label: "Teacher" },
              ]}
            />
            <Button variant="solid" color="primary">
              Search
            </Button>
            <Button>Reset</Button>
          </Flex>
          <Table
            columns={[
              {
                title: "Name",
                dataIndex: "name",
                key: "name",
              },
              {
                title: "Age",
                dataIndex: "age",
                key: "age",
                width: 80,
              },
              {
                title: "Address",
                dataIndex: "address",
                key: "address",
                ellipsis: true,
              },
              {
                title: "Tags",
                dataIndex: "tags",
                key: "tags",
                render: (tags: string[]) => (
                  <Flex gap={4} wrap>
                    {tags.map((tag) => (
                      <Tag key={tag}>{tag}</Tag>
                    ))}
                  </Flex>
                ),
              },
              {
                title: "Action",
                key: "action",
                width: 120,
                render: () => (
                  <Flex gap={8}>
                    <Button type="link" size="small">
                      Edit
                    </Button>
                    <Button type="link" size="small" danger>
                      Delete
                    </Button>
                  </Flex>
                ),
              },
            ]}
            dataSource={data}
            pagination={{
              pageSize: 5,
              showTotal: (total) => `Total ${total} items`,
            }}
            size="medium"
          />
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Table {...args} columns={basicColumns} dataSource={data.slice(0, 4)} />
  ),
}

export const Interactive: Story = {
  args: {
    columns: basicColumns,
    dataSource: data.slice(0, 3),
    rowSelection: { type: "checkbox" },
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const checkboxes = canvas.getAllByRole("checkbox")
    await checkboxes[1].click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

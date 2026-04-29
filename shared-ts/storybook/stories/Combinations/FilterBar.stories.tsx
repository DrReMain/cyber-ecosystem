import { SearchOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Card, DatePicker, Flex, Input, Select, Table, Tag } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { Label, Section } from "../helpers"

const { RangePicker } = DatePicker

const meta: Meta = {
  title: "Antd/Combinations/FilterBar",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

const mockData = [
  {
    key: "1",
    name: "Project Alpha",
    status: "Active",
    category: "Engineering",
    date: "2025-01-15",
  },
  {
    key: "2",
    name: "Project Beta",
    status: "Pending",
    category: "Design",
    date: "2025-02-20",
  },
  {
    key: "3",
    name: "Project Gamma",
    status: "Completed",
    category: "Engineering",
    date: "2025-03-10",
  },
  {
    key: "4",
    name: "Project Delta",
    status: "Active",
    category: "Marketing",
    date: "2025-04-05",
  },
  {
    key: "5",
    name: "Project Epsilon",
    status: "On Hold",
    category: "Design",
    date: "2025-05-12",
  },
  {
    key: "6",
    name: "Project Zeta",
    status: "Active",
    category: "Engineering",
    date: "2025-06-01",
  },
  {
    key: "7",
    name: "Project Eta",
    status: "Completed",
    category: "Marketing",
    date: "2025-06-15",
  },
  {
    key: "8",
    name: "Project Theta",
    status: "Pending",
    category: "Engineering",
    date: "2025-07-22",
  },
]

const statusColors: Record<string, string> = {
  Active: "processing",
  Pending: "warning",
  Completed: "success",
  "On Hold": "default",
}

const columns = [
  { title: "Name", dataIndex: "name", key: "name" },
  {
    title: "Status",
    dataIndex: "status",
    key: "status",
    render: (status: string) => (
      <Tag color={statusColors[status]}>{status}</Tag>
    ),
  },
  { title: "Category", dataIndex: "category", key: "category" },
  { title: "Date", dataIndex: "date", key: "date" },
  {
    title: "Action",
    key: "action",
    render: () => (
      <Button type="link" size="small">
        View
      </Button>
    ),
  },
]

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="Search & Filter Toolbar"
        description="Input.Search + Select + DatePicker.RangePicker + Button"
      >
        <Card>
          <Flex vertical gap={16}>
            <Flex gap={8} wrap>
              <Input.Search
                placeholder="Search projects..."
                style={{ width: 260 }}
                allowClear
              />
              <Select
                placeholder="Status"
                style={{ width: 140 }}
                allowClear
                options={[
                  { value: "active", label: "Active" },
                  { value: "pending", label: "Pending" },
                  { value: "completed", label: "Completed" },
                  { value: "on-hold", label: "On Hold" },
                ]}
              />
              <Select
                placeholder="Category"
                style={{ width: 150 }}
                allowClear
                options={[
                  { value: "engineering", label: "Engineering" },
                  { value: "design", label: "Design" },
                  { value: "marketing", label: "Marketing" },
                ]}
              />
              <RangePicker />
              <Button variant="solid" color="primary" icon={<SearchOutlined />}>
                Search
              </Button>
              <Button>Reset</Button>
            </Flex>
            <Table
              columns={columns}
              dataSource={mockData}
              pagination={{
                pageSize: 5,
                showTotal: (total) => `Total ${total} items`,
              }}
              size="medium"
            />
          </Flex>
        </Card>
      </Section>

      <Section
        title="Inline Filter"
        description="Compact filter row for quick data filtering"
      >
        <Flex gap={8} align="center" wrap>
          <Label>Quick filter:</Label>
          {["All", "Active", "Pending", "Completed"].map((label, i) => (
            <Button
              key={label}
              size="small"
              type={i === 0 ? "primary" : "default"}
            >
              {label}
            </Button>
          ))}
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Card>
      <Flex vertical gap={16}>
        <Flex gap={8} wrap>
          <Input.Search
            placeholder="Search..."
            style={{ width: 260 }}
            allowClear
          />
          <Select
            placeholder="Filter"
            style={{ width: 140 }}
            allowClear
            options={[]}
          />
          <RangePicker />
          <Button variant="solid" color="primary">
            Apply
          </Button>
          <Button>Reset</Button>
        </Flex>
        <Table
          columns={columns}
          dataSource={mockData}
          pagination={{ pageSize: 5 }}
          size="medium"
        />
      </Flex>
    </Card>
  ),
}

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    onSearch: fn(),
  },
  render: (args: any) => (
    <Card>
      <Flex vertical gap={16}>
        <Flex gap={8} wrap>
          <Input.Search
            id="filter-search"
            placeholder="Search projects..."
            style={{ width: 260 }}
            allowClear
            onSearch={args.onSearch}
          />
          <Select
            id="filter-status"
            placeholder="Status"
            style={{ width: 140 }}
            allowClear
            options={[
              { value: "active", label: "Active" },
              { value: "pending", label: "Pending" },
            ]}
          />
          <Button variant="solid" color="primary">
            Search
          </Button>
        </Flex>
        <Table
          columns={columns}
          dataSource={mockData}
          pagination={{ pageSize: 5 }}
          size="medium"
        />
      </Flex>
    </Card>
  ),
  play: async ({ canvasElement, args }: any) => {
    const canvas = within(canvasElement)
    const input = canvas.getByPlaceholderText("Search projects...")
    await userEvent.type(input, "Alpha")
    await userEvent.keyboard("{Enter}")
    await expect(args.onSearch).toHaveBeenCalledWith("Alpha")
  },
}

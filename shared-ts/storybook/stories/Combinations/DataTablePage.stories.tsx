import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons"
import { Filter } from "@shared/antd/filter"
import { TableToolbar } from "@shared/antd/table"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Badge,
  Breadcrumb,
  Button,
  Card,
  Flex,
  Popconfirm,
  Space,
  Table,
  Tag,
  Typography,
} from "antd"
import type { Key } from "react"
import { useMemo, useState } from "react"

const meta: Meta = {
  title: "Antd/Combinations/DataTablePage",
  parameters: { layout: "padded" },
}

export default meta
type Story = StoryObj

interface Project {
  key: string
  name: string
  status: "Active" | "Pending" | "Completed" | "On Hold"
  category: "Engineering" | "Design" | "Marketing"
  owner: string
  budget: number
  createdAt: string
}

const allData: Project[] = [
  {
    key: "1",
    name: "Project Alpha",
    status: "Active",
    category: "Engineering",
    owner: "Alice Chen",
    budget: 120000,
    createdAt: "2025-01-15",
  },
  {
    key: "2",
    name: "Project Beta",
    status: "Pending",
    category: "Design",
    owner: "Bob Smith",
    budget: 85000,
    createdAt: "2025-02-20",
  },
  {
    key: "3",
    name: "Project Gamma",
    status: "Completed",
    category: "Engineering",
    owner: "Carol Lee",
    budget: 200000,
    createdAt: "2025-03-10",
  },
  {
    key: "4",
    name: "Project Delta",
    status: "Active",
    category: "Marketing",
    owner: "David Wang",
    budget: 150000,
    createdAt: "2025-04-05",
  },
  {
    key: "5",
    name: "Project Epsilon",
    status: "On Hold",
    category: "Design",
    owner: "Eve Park",
    budget: 95000,
    createdAt: "2025-05-12",
  },
  {
    key: "6",
    name: "Project Zeta",
    status: "Active",
    category: "Engineering",
    owner: "Frank Zhao",
    budget: 180000,
    createdAt: "2025-06-01",
  },
  {
    key: "7",
    name: "Project Eta",
    status: "Completed",
    category: "Marketing",
    owner: "Grace Kim",
    budget: 110000,
    createdAt: "2025-06-15",
  },
  {
    key: "8",
    name: "Project Theta",
    status: "Pending",
    category: "Engineering",
    owner: "Henry Liu",
    budget: 135000,
    createdAt: "2025-07-22",
  },
]

const statusColors: Record<string, string> = {
  Active: "processing",
  Pending: "warning",
  Completed: "success",
  "On Hold": "default",
}

const categoryColors: Record<string, string> = {
  Engineering: "blue",
  Design: "purple",
  Marketing: "orange",
}

function formatBudget(v: number) {
  return `$${(v / 1000).toFixed(0)}k`
}

function ProjectTablePage({
  data,
  loading,
  empty,
  rowSelection,
  pagination,
}: {
  data: Project[]
  loading?: boolean
  empty?: boolean
  rowSelection?: object
  pagination?: object | false
}) {
  const [tableSize, setTableSize] = useState<"small" | "middle" | "large">(
    "middle",
  )
  const [refreshing, setRefreshing] = useState(false)
  const [filteredData, setFilteredData] = useState<Project[]>(data)

  const displayData = useMemo(() => {
    if (empty) return []
    return filteredData
  }, [empty, filteredData])

  const columns = useMemo(
    () => [
      {
        title: "Project Name",
        dataIndex: "name",
        key: "name",
        sorter: (a: Project, b: Project) => a.name.localeCompare(b.name),
      },
      {
        title: "Status",
        dataIndex: "status",
        key: "status",
        filters: [
          { text: "Active", value: "Active" },
          { text: "Pending", value: "Pending" },
          { text: "Completed", value: "Completed" },
          { text: "On Hold", value: "On Hold" },
        ],
        onFilter: (value: boolean | Key, record: Project) =>
          record.status === value,
        render: (status: string) => (
          <Badge
            status={
              status === "Active"
                ? "processing"
                : status === "Pending"
                  ? "warning"
                  : status === "Completed"
                    ? "success"
                    : "default"
            }
            text={<Tag color={statusColors[status]}>{status}</Tag>}
          />
        ),
      },
      {
        title: "Category",
        dataIndex: "category",
        key: "category",
        filters: [
          { text: "Engineering", value: "Engineering" },
          { text: "Design", value: "Design" },
          { text: "Marketing", value: "Marketing" },
        ],
        onFilter: (value: boolean | Key, record: Project) =>
          record.category === value,
        render: (category: string) => (
          <Tag color={categoryColors[category]}>{category}</Tag>
        ),
      },
      {
        title: "Owner",
        dataIndex: "owner",
        key: "owner",
      },
      {
        title: "Budget",
        dataIndex: "budget",
        key: "budget",
        sorter: (a: Project, b: Project) => a.budget - b.budget,
        render: (budget: number) => formatBudget(budget),
      },
      {
        title: "Created",
        dataIndex: "createdAt",
        key: "createdAt",
        sorter: (a: Project, b: Project) =>
          a.createdAt.localeCompare(b.createdAt),
      },
      {
        title: "Action",
        key: "action",
        render: () => (
          <Space size="small">
            <Button type="text" size="small" icon={<EditOutlined />} />
            <Popconfirm
              title="Delete this project?"
              okText="Yes"
              cancelText="No"
            >
              <Button
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Space>
        ),
      },
    ],
    [],
  )

  const handleRefresh = () => {
    setRefreshing(true)
    setTimeout(() => setRefreshing(false), 1000)
  }

  return (
    <Flex vertical gap={16}>
      {/* Page Header */}
      <Flex justify="space-between" align="center" wrap gap={8}>
        <Flex vertical gap={4}>
          <Breadcrumb items={[{ title: "Workspace" }, { title: "Projects" }]} />
          <Typography.Title level={4} style={{ margin: 0 }}>
            Project Management
          </Typography.Title>
        </Flex>
        <Button type="primary" variant="solid" icon={<PlusOutlined />}>
          New Project
        </Button>
      </Flex>

      {/* Filter */}
      <Card size="small">
        <Filter<Record<string, unknown>>
          options={[
            {
              label: "Project Name",
              name: "name",
              type: "input",
              placeholder: "Search by name",
            },
            {
              label: "Status",
              name: "status",
              type: "select",
              placeholder: "Select status",
              options: [
                { label: "Active", value: "Active" },
                { label: "Pending", value: "Pending" },
                { label: "Completed", value: "Completed" },
                { label: "On Hold", value: "On Hold" },
              ],
            },
            {
              label: "Category",
              name: "category",
              type: "select",
              placeholder: "Select category",
              options: [
                { label: "Engineering", value: "Engineering" },
                { label: "Design", value: "Design" },
                { label: "Marketing", value: "Marketing" },
              ],
            },
            {
              label: "Date Range",
              name: ["startDate", "endDate"],
              type: "range-date",
              placeholder: ["Start", "End"],
            },
            {
              label: "Budget Range",
              name: ["minBudget", "maxBudget"],
              type: "range-number",
              placeholder: ["Min", "Max"],
            },
          ]}
          onFilter={(values) => {
            let result = [...allData]
            if (values.name) {
              result = result.filter((d) =>
                d.name
                  .toLowerCase()
                  .includes(String(values.name).toLowerCase()),
              )
            }
            if (values.status) {
              result = result.filter((d) => d.status === values.status)
            }
            if (values.category) {
              result = result.filter((d) => d.category === values.category)
            }
            setFilteredData(result)
          }}
          onReset={() => setFilteredData(allData)}
          columns={4}
        />
      </Card>

      {/* Table Section */}
      <Card
        size="small"
        title={
          <Flex justify="space-between" align="center">
            <Typography.Text strong>
              All Projects
              {displayData.length > 0 && (
                <Typography.Text type="secondary" style={{ marginLeft: 8 }}>
                  ({displayData.length})
                </Typography.Text>
              )}
            </Typography.Text>
            <TableToolbar
              size={tableSize}
              onSizeChange={(s) => setTableSize(s)}
              onRefresh={handleRefresh}
              loading={refreshing}
            />
          </Flex>
        }
      >
        <Table
          columns={columns}
          dataSource={displayData}
          loading={loading}
          size={tableSize}
          rowSelection={rowSelection}
          pagination={
            pagination === false
              ? false
              : {
                  pageSize: 5,
                  showSizeChanger: true,
                  showTotal: (total) => `Total ${total} projects`,
                  ...pagination,
                }
          }
        />
      </Card>
    </Flex>
  )
}

export const Normal: Story = {
  parameters: { controls: { disable: true } },
  render: () => <ProjectTablePage data={allData} />,
}

export const Loading: Story = {
  parameters: { controls: { disable: true } },
  render: () => <ProjectTablePage data={allData} loading />,
}

export const Empty: Story = {
  parameters: { controls: { disable: true } },
  render: () => <ProjectTablePage data={[]} empty />,
}

export const FilterEmpty: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <ProjectTablePage
      data={[]}
      empty
      // Simulate that filter was applied but no results
    />
  ),
}

export const WithSelection: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <ProjectTablePage
      data={allData}
      rowSelection={{
        type: "checkbox",
        onChange: (_selectedRowKeys: Key[]) => {
          // selection handled by Table
        },
      }}
    />
  ),
}

export const NoPagination: Story = {
  parameters: { controls: { disable: true } },
  render: () => <ProjectTablePage data={allData} pagination={false} />,
}

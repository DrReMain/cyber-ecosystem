// Table data
export interface TableDataType {
  key: string
  name: string
  age: number
  address: string
  status?: string
}

export const tableData: TableDataType[] = [
  {
    key: "1",
    name: "John Brown",
    age: 32,
    address: "New York No. 1 Lake Park",
    status: "Active",
  },
  {
    key: "2",
    name: "Jim Green",
    age: 42,
    address: "London No. 1 Lake Park",
    status: "Inactive",
  },
  {
    key: "3",
    name: "Joe Black",
    age: 28,
    address: "Sidney No. 1 Lake Park",
    status: "Active",
  },
  {
    key: "4",
    name: "Disabled User",
    age: 35,
    address: "Paris No. 2 Lake Park",
    status: "Pending",
  },
  {
    key: "5",
    name: "Jane Smith",
    age: 24,
    address: "Berlin No. 3 Lake Park",
    status: "Active",
  },
  {
    key: "6",
    name: "Tom Wilson",
    age: 38,
    address: "Tokyo No. 4 Lake Park",
    status: "Inactive",
  },
  {
    key: "7",
    name: "Lucy Davis",
    age: 29,
    address: "Seoul No. 5 Lake Park",
    status: "Active",
  },
  {
    key: "8",
    name: "Mike Johnson",
    age: 45,
    address: "Beijing No. 6 Lake Park",
    status: "Pending",
  },
]

// Tree data
export interface TreeNodeType {
  title: string
  value: string
  key?: string
  children?: TreeNodeType[]
  disabled?: boolean
}

export const treeData: TreeNodeType[] = [
  {
    title: "Node 1",
    value: "1",
    children: [
      { title: "Child 1-1", value: "1-1" },
      { title: "Child 1-2", value: "1-2" },
    ],
  },
  {
    title: "Node 2",
    value: "2",
    children: [
      { title: "Child 2-1", value: "2-1" },
      { title: "Child 2-2", value: "2-2" },
    ],
  },
  {
    title: "Node 3",
    value: "3",
    children: [
      { title: "Child 3-1", value: "3-1" },
      { title: "Child 3-2", value: "3-2" },
    ],
  },
]

// Status color mapping
export const statusColorMap: Record<string, string> = {
  Active: "green",
  Inactive: "red",
  Pending: "orange",
}

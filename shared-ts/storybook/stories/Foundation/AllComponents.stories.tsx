import {
  InboxOutlined,
  PlusOutlined,
  SettingOutlined,
  UploadOutlined,
  UserOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Alert,
  Anchor,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Calendar,
  Card,
  Cascader,
  Checkbox,
  Col,
  Collapse,
  ColorPicker,
  DatePicker,
  Descriptions,
  Divider,
  Dropdown,
  Empty,
  Flex,
  Form,
  Image,
  Input,
  InputNumber,
  List,
  Mentions,
  Menu,
  Pagination,
  Popconfirm,
  Popover,
  Progress,
  QRCode,
  Radio,
  Rate,
  Result,
  Row,
  Segmented,
  Select,
  Skeleton,
  Slider,
  Space,
  Spin,
  Statistic,
  Steps,
  Switch,
  Table,
  Tabs,
  Tag,
  Timeline,
  TimePicker,
  Tooltip,
  Transfer,
  Tree,
  TreeSelect,
  Typography,
  Upload,
} from "antd"
import { Label, Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Foundation/AllComponents",
  parameters: { layout: "padded", controls: { disable: true } },
}

export default meta
type Story = StoryObj

const { Title, Text, Paragraph } = Typography

export const Gallery: Story = {
  render: () => {
    const selectOptions = [
      { value: "apple", label: "Apple" },
      { value: "banana", label: "Banana" },
      { value: "cherry", label: "Cherry" },
    ]

    const cascaderOptions = [
      {
        value: "zhejiang",
        label: "Zhejiang",
        children: [{ value: "hangzhou", label: "Hangzhou" }],
      },
    ]

    const treeData = [
      { title: "Node 1", key: "1", children: [{ title: "Leaf", key: "1-1" }] },
      { title: "Node 2", key: "2" },
    ]

    const tableColumns = [
      { title: "Name", dataIndex: "name", key: "name" },
      { title: "Age", dataIndex: "age", key: "age" },
    ]
    const tableData = [
      { key: "1", name: "Alice", age: 28 },
      { key: "2", name: "Bob", age: 34 },
    ]

    return (
      <Flex vertical gap={40} style={{ maxWidth: 1000 }}>
        <Section title="General">
          <Flex vertical gap={16}>
            <Space wrap>
              <Button variant="solid" color="primary">
                Primary
              </Button>
              <Button>Default</Button>
              <Button variant="dashed">Dashed</Button>
              <Button variant="text">Text</Button>
              <Button variant="link">Link</Button>
              <Button danger>Danger</Button>
              <Button disabled>Disabled</Button>
              <Button variant="solid" color="primary" icon={<PlusOutlined />}>
                Icon
              </Button>
            </Space>
            <Space wrap>
              <Typography>
                <Title level={4}>Typography</Title>
                <Paragraph>Body paragraph text.</Paragraph>
                <Text>Default</Text> <Text type="secondary">Secondary</Text>{" "}
                <Text type="success">Success</Text>{" "}
                <Text type="warning">Warning</Text>{" "}
                <Text type="danger">Danger</Text>
              </Typography>
            </Space>
            <Space>
              <Input placeholder="Input" style={{ width: 180 }} />
              <Input.Search placeholder="Search" style={{ width: 200 }} />
              <Input.Password placeholder="Password" style={{ width: 180 }} />
              <InputNumber defaultValue={0} style={{ width: 120 }} />
            </Space>
          </Flex>
        </Section>

        <Section title="Layout">
          <Flex vertical gap={16}>
            <Card title="Card" size="small" style={{ width: 300 }}>
              <p>Card content</p>
            </Card>
            <Space>
              <Divider style={{ width: 100 }} />
              <Divider vertical />
            </Space>
            <Row gutter={8}>
              <Col span={8}>
                <div
                  style={{
                    background: "var(--ant-color-primary)",
                    opacity: 0.1,
                    padding: 8,
                    textAlign: "center",
                    borderRadius: 4,
                  }}
                >
                  Col 8
                </div>
              </Col>
              <Col span={8}>
                <div
                  style={{
                    background: "var(--ant-color-primary)",
                    opacity: 0.1,
                    padding: 8,
                    textAlign: "center",
                    borderRadius: 4,
                  }}
                >
                  Col 8
                </div>
              </Col>
              <Col span={8}>
                <div
                  style={{
                    background: "var(--ant-color-primary)",
                    opacity: 0.1,
                    padding: 8,
                    textAlign: "center",
                    borderRadius: 4,
                  }}
                >
                  Col 8
                </div>
              </Col>
            </Row>
            <Space>
              <Flex gap={8}>
                <div
                  style={{
                    width: 40,
                    height: 40,
                    background:
                      "color-mix(in srgb, var(--ant-color-primary) 10%, transparent)",
                    borderRadius: 4,
                  }}
                />
                <div
                  style={{
                    width: 40,
                    height: 40,
                    background:
                      "color-mix(in srgb, var(--ant-color-primary) 10%, transparent)",
                    borderRadius: 4,
                  }}
                />
              </Flex>
              <Space>Space item 1 | Space item 2</Space>
            </Space>
          </Flex>
        </Section>

        <Section title="Navigation">
          <Flex vertical gap={16}>
            <Breadcrumb
              items={[{ title: "Home" }, { title: "List" }, { title: "App" }]}
            />
            <Menu
              mode="horizontal"
              defaultSelectedKeys={["mail"]}
              items={[
                { key: "mail", label: "Mail", icon: <InboxOutlined /> },
                { key: "app", label: "App" },
                {
                  key: "settings",
                  label: "Settings",
                  icon: <SettingOutlined />,
                },
              ]}
            />
            <Pagination total={50} pageSize={10} size="small" />
            <Steps
              size="small"
              current={1}
              items={[
                { title: "Step 1" },
                { title: "Step 2" },
                { title: "Step 3" },
              ]}
            />
            <Dropdown
              menu={{
                items: [
                  { key: "1", label: "Item 1" },
                  { key: "2", label: "Item 2" },
                ],
              }}
            >
              <Button>Dropdown</Button>
            </Dropdown>
            <Anchor
              direction="horizontal"
              items={[
                { key: "1", href: "#", title: "Section 1" },
                { key: "2", href: "#", title: "Section 2" },
              ]}
            />
            <Segmented options={["Map", "Transit", "Satellite"]} />
            <Tabs
              items={[
                { key: "1", label: "Tab 1", children: "Content 1" },
                { key: "2", label: "Tab 2", children: "Content 2" },
              ]}
            />
          </Flex>
        </Section>

        <Section title="Data Entry">
          <Flex vertical gap={16}>
            <Flex vertical gap={8}>
              <Label>Text Input</Label>
              <Flex gap={8} wrap>
                <Mentions placeholder="Mentions" style={{ width: 300 }} />
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Selection</Label>
              <Flex gap={8} wrap>
                <Select
                  options={selectOptions}
                  defaultValue="apple"
                  style={{ width: 180 }}
                />
                <TreeSelect
                  treeData={treeData}
                  defaultValue="1"
                  style={{ width: 180 }}
                />
                <Cascader
                  options={cascaderOptions}
                  placeholder="Cascader"
                  style={{ width: 200 }}
                />
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Pickers</Label>
              <Flex gap={8} wrap>
                <DatePicker style={{ width: 180 }} />
                <TimePicker style={{ width: 180 }} />
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Toggles</Label>
              <Flex gap={16} wrap>
                <Space wrap>
                  <Checkbox>Checkbox</Checkbox>
                  <Checkbox checked>Checked</Checkbox>
                  <Checkbox disabled>Disabled</Checkbox>
                </Space>
                <Radio.Group defaultValue="a">
                  <Radio value="a">A</Radio>
                  <Radio value="b">B</Radio>
                  <Radio value="c">C</Radio>
                </Radio.Group>
                <Space wrap>
                  <Switch defaultChecked />
                  <Switch />
                  <Switch disabled />
                </Space>
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Other</Label>
              <Flex gap={8} wrap>
                <Slider defaultValue={30} style={{ width: 200 }} />
                <Rate defaultValue={3} />
                <ColorPicker defaultValue="#1677ff" />
                <Upload>
                  <Button icon={<UploadOutlined />}>Upload</Button>
                </Upload>
              </Flex>
              <Transfer
                dataSource={[
                  { key: "1", title: "Item 1" },
                  { key: "2", title: "Item 2" },
                ]}
                targetKeys={[]}
                render={(item: { key: string; title: string }) =>
                  item.title ?? item.key
                }
                oneWay
              />
            </Flex>
          </Flex>
        </Section>

        <Section title="Data Display">
          <Flex vertical gap={16}>
            <Flex vertical gap={8}>
              <Label>Tags & Badges</Label>
              <Flex gap={16} align="center" wrap>
                <Space wrap>
                  <Tag>Default</Tag>
                  <Tag color="success">Success</Tag>
                  <Tag color="processing">Processing</Tag>
                  <Tag color="error">Error</Tag>
                  <Tag color="warning">Warning</Tag>
                </Space>
                <Space>
                  <Badge count={5}>
                    <div
                      style={{
                        width: 40,
                        height: 40,
                        background:
                          "color-mix(in srgb, var(--ant-color-text) 6%, transparent)",
                        borderRadius: 4,
                      }}
                    />
                  </Badge>
                  <Badge dot>
                    <div
                      style={{
                        width: 40,
                        height: 40,
                        background:
                          "color-mix(in srgb, var(--ant-color-text) 6%, transparent)",
                        borderRadius: 4,
                      }}
                    />
                  </Badge>
                </Space>
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Avatars & Media</Label>
              <Flex gap={16} align="center" wrap>
                <Space>
                  <Avatar icon={<UserOutlined />} aria-label="User avatar" />
                  <Avatar>U</Avatar>
                  <Avatar src="https://api.dicebear.com/7.x/miniavs/svg?seed=1" />
                </Space>
                <QRCode value="https://ant.design" size={80} />
                <Image
                  width={100}
                  src="https://picsum.photos/seed/antd/100/60"
                  preview={false}
                />
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Overlays</Label>
              <Flex gap={8} wrap>
                <Tooltip title="Tooltip text">
                  <Button>Hover me</Button>
                </Tooltip>
                <Popover content="Popover content" title="Title">
                  <Button>Click me</Button>
                </Popover>
                <Popconfirm title="Are you sure?">
                  <Button>Delete</Button>
                </Popconfirm>
              </Flex>
            </Flex>
            <Flex vertical gap={8}>
              <Label>Data Views</Label>
              <Table
                columns={tableColumns}
                dataSource={tableData}
                size="small"
                pagination={false}
              />
              <List
                size="small"
                dataSource={["Item A", "Item B", "Item C"]}
                renderItem={(item: string) => <List.Item>{item}</List.Item>}
              />
              <Collapse
                items={[
                  {
                    key: "1",
                    label: "Panel 1",
                    children: "Content of panel 1",
                  },
                ]}
              />
              <Tree treeData={treeData} defaultExpandAll />
              <Descriptions
                title="Descriptions"
                size="small"
                column={2}
                items={[
                  { key: "1", label: "Name", children: "Alice" },
                  { key: "2", label: "Age", children: "28" },
                ]}
              />
              <Timeline
                items={[
                  { children: "Step 1" },
                  { children: "Step 2", color: "green" },
                  { children: "Step 3" },
                ]}
              />
              <Calendar fullscreen={false} />
            </Flex>
            <Flex vertical gap={8}>
              <Label>Metrics</Label>
              <Flex gap={16} wrap>
                <Statistic title="Statistic" value={12345} />
              </Flex>
            </Flex>
          </Flex>
        </Section>

        <Section title="Feedback">
          <Flex vertical gap={16}>
            <Space vertical wrap style={{ width: "100%" }}>
              <Alert title="Info alert" type="info" showIcon />
              <Alert title="Success alert" type="success" showIcon />
              <Alert title="Warning alert" type="warning" showIcon />
              <Alert title="Error alert" type="error" showIcon />
            </Space>
            <Progress percent={60} style={{ width: 200 }} />
            <Spin />
            <Skeleton active style={{ width: 300 }} />
            <Result
              status="success"
              title="Success"
              subTitle="Operation completed"
              extra={
                <Button variant="solid" color="primary">
                  Done
                </Button>
              }
            />
            <Empty description="No data" />
          </Flex>
        </Section>

        <Section title="Form">
          <Card size="small" style={{ maxWidth: 400 }}>
            <Form layout="vertical" size="small">
              <Form.Item label="Name">
                <Input placeholder="Name" />
              </Form.Item>
              <Form.Item label="Email">
                <Input placeholder="Email" />
              </Form.Item>
              <Form.Item label="Select">
                <Select options={selectOptions} placeholder="Choose" />
              </Form.Item>
              <Form.Item label="Switch">
                <Switch />
              </Form.Item>
              <Form.Item>
                <Button variant="solid" color="primary">
                  Submit
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Section>
      </Flex>
    )
  },
}

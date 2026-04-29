import { DeleteOutlined, EditOutlined, EyeOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Avatar, Button, Flex, List } from "antd"
import { expect, within } from "storybook/test"
import { borderedArg, loadingArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof List> = {
  title: "Antd/Data Display/List",
  component: List,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    size: sizeArg(["small", "default", "large"]),
    bordered: borderedArg,
    split: { control: "boolean" },
    loading: loadingArg,
    grid: { control: false },
  },
}

export default meta
type Story = StoryObj<typeof List>

interface ListItem {
  title: string
  description: string
}

const data: ListItem[] = Array.from({ length: 6 }, (_, i) => ({
  title: `Title ${i + 1}`,
  description: `Description for item ${i + 1}`,
}))

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes" description="small, default, large">
        {(["small", "default", "large"] as const).map((size) => (
          <div key={size} style={{ marginBottom: 12 }}>
            <Label>{size}</Label>
            <List
              size={size}
              dataSource={data.slice(0, 2)}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
                    title={item.title}
                    description={item.description}
                  />
                </List.Item>
              )}
            />
          </div>
        ))}
      </Section>

      <Section title="Bordered">
        <List
          bordered
          dataSource={data.slice(0, 3)}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
                title={item.title}
                description={item.description}
              />
            </List.Item>
          )}
        />
      </Section>

      <Section title="No Split">
        <List
          split={false}
          dataSource={data.slice(0, 3)}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
                title={item.title}
                description={item.description}
              />
            </List.Item>
          )}
        />
      </Section>

      <Section title="With Actions">
        <List
          dataSource={data.slice(0, 3)}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button key="view" type="link" icon={<EyeOutlined />}>
                  View
                </Button>,
                <Button key="edit" type="link" icon={<EditOutlined />}>
                  Edit
                </Button>,
                <Button
                  key="delete"
                  type="link"
                  danger
                  icon={<DeleteOutlined />}
                >
                  Delete
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
                title={item.title}
                description={item.description}
              />
            </List.Item>
          )}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <List
      {...args}
      dataSource={data.slice(0, 3)}
      renderItem={(item) => (
        <List.Item>
          <List.Item.Meta
            avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
            title={item.title}
            description={item.description}
          />
        </List.Item>
      )}
    />
  ),
}

export const Interactive: Story = {
  render: () => (
    <List
      dataSource={data.slice(0, 3)}
      renderItem={(item) => (
        <List.Item
          actions={[
            <Button key="view" type="link" icon={<EyeOutlined />}>
              View
            </Button>,
          ]}
        >
          <List.Item.Meta
            avatar={<Avatar>{item.title.charAt(0)}</Avatar>}
            title={item.title}
            description={item.description}
          />
        </List.Item>
      )}
    />
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const viewButtons = canvas.getAllByText("View")
    await viewButtons[0].click()
    await expect(viewButtons[0]).toBeInTheDocument()
  },
}

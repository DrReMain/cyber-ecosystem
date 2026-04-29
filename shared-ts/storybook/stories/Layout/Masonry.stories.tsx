import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Masonry } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Masonry> = {
  title: "Antd/Layout/Masonry",
  component: Masonry,
  parameters: { layout: "padded" },
  args: { columns: 3, gutter: 16 },
  argTypes: {
    columns: {
      control: "radio",
      options: [2, 3, 4, 5],
    },
    gutter: {
      control: "select",
      options: [0, 8, 12, 16, 24],
    },
  },
}

export default meta
type Story = StoryObj<typeof Masonry>

const colors = [
  "#1677ff",
  "#52c41a",
  "#faad14",
  "#eb2f96",
  "#722ed1",
  "#13c2c2",
  "#f5222d",
  "#2f54eb",
  "#fa8c16",
]

const basicHeights = [
  120, 200, 150, 180, 220, 160, 190, 140, 170, 210, 130, 200,
]
const responsiveHeights = [
  100, 160, 120, 150, 140, 110, 130, 155, 105, 145, 125, 135, 115, 165, 108,
  148, 128, 138,
]
const gutterHeights = [
  120, 160, 140, 180, 110, 150, 130, 170, 125, 165, 115, 155,
]

const basicItems = Array.from({ length: 12 }, (_, i) => ({
  key: `item-${i}`,
  data: { index: i },
  height: basicHeights[i],
}))

const responsiveItems = Array.from({ length: 18 }, (_, i) => ({
  key: `resp-${i}`,
  data: { index: i },
  height: responsiveHeights[i],
}))

const gutterItems = Array.from({ length: 12 }, (_, i) => ({
  key: `gutter-${i}`,
  data: { index: i },
  height: gutterHeights[i],
}))

const renderItem = (item: { index: number; height?: number }) => (
  <div
    style={{
      padding: 16,
      background: colors[item.index % colors.length],
      color: "var(--ant-color-text)",
      borderRadius: 8,
      height: item.height ?? 120,
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      fontWeight: 600,
      fontSize: 16,
    }}
  >
    Item {item.index + 1}
  </div>
)

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Masonry
          columns={3}
          gutter={16}
          items={basicItems}
          itemRender={renderItem}
        />
      </Section>
      <Section
        title="Responsive columns"
        description="Resize the browser to see columns change."
      >
        <Masonry
          columns={{ xs: 1, sm: 2, md: 3, lg: 4, xl: 5 }}
          gutter={12}
          items={responsiveItems}
          itemRender={(item) => (
            <div
              style={{
                padding: 16,
                background: colors[item.index % colors.length],
                color: "var(--ant-color-text)",
                borderRadius: 8,
                height: item.height,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontWeight: 600,
              }}
            >
              Item {item.index + 1}
            </div>
          )}
        />
      </Section>
      <Section
        title="Gutter: [24, 24]"
        description="24px horizontal and vertical gap."
      >
        <Masonry
          columns={3}
          gutter={[24, 24]}
          items={gutterItems}
          itemRender={(item) => (
            <div
              style={{
                padding: 16,
                background: colors[item.index % colors.length],
                color: "var(--ant-color-text)",
                borderRadius: 8,
                height: item.height,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontWeight: 600,
              }}
            >
              Item {item.index + 1}
            </div>
          )}
        />
      </Section>
      <Section title="No gutter" description="Zero gap between items.">
        <Masonry
          columns={4}
          gutter={0}
          items={gutterItems}
          itemRender={(item) => (
            <div
              style={{
                padding: 16,
                background: colors[item.index % colors.length],
                color: "var(--ant-color-text)",
                height: item.height,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontWeight: 600,
              }}
            >
              Item {item.index + 1}
            </div>
          )}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Masonry {...args} items={basicItems} itemRender={renderItem} />
  ),
}

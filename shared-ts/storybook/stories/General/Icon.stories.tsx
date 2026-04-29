import {
  CheckCircleTwoTone,
  EditTwoTone,
  HeartFilled,
  HeartTwoTone,
  HomeOutlined,
  LoadingOutlined,
  SearchOutlined,
  SettingFilled,
  SmileOutlined,
  StarFilled,
  SyncOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Space } from "antd"
import { Section } from "../helpers"

const meta: Meta = {
  title: "Antd/General/Icon",
  parameters: { layout: "padded" },
  args: {
    style: { fontSize: 24 },
  },
  argTypes: {
    spin: { control: "boolean" },
    rotate: { control: { type: "range", min: 0, max: 360, step: 15 } },
    style: { control: "object" },
  },
}

export default meta
type Story = StoryObj

const sizes = [12, 16, 24, 32, 48, 64] as const

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Outlined">
        <Space size="large" wrap>
          <HomeOutlined style={{ fontSize: 24 }} />
          <SmileOutlined style={{ fontSize: 24 }} />
          <SearchOutlined style={{ fontSize: 24 }} />
          <SyncOutlined style={{ fontSize: 24 }} />
        </Space>
      </Section>

      <Section title="Filled">
        <Space size="large" wrap>
          <SettingFilled style={{ fontSize: 24 }} />
          <StarFilled style={{ fontSize: 24 }} />
          <HeartFilled style={{ fontSize: 24 }} />
        </Space>
      </Section>

      <Section title="Sizes">
        <Space size="middle" align="center" wrap>
          {sizes.map((s) => (
            <SmileOutlined key={s} style={{ fontSize: s }} />
          ))}
        </Space>
      </Section>

      <Section title="Spin">
        <Space size="large" wrap>
          <LoadingOutlined style={{ fontSize: 24 }} />
          <SyncOutlined spin style={{ fontSize: 24 }} />
          <SmileOutlined spin style={{ fontSize: 24 }} />
        </Space>
      </Section>

      <Section title="Rotate">
        <Space size="large" wrap>
          {[0, 45, 90, 135, 180, 270].map((deg) => (
            <HomeOutlined key={deg} rotate={deg} style={{ fontSize: 24 }} />
          ))}
        </Space>
      </Section>

      <Section title="Custom Colors">
        <Space size="large" wrap>
          <HomeOutlined style={{ fontSize: 24, color: "#1677ff" }} />
          <HomeOutlined style={{ fontSize: 24, color: "#52c41a" }} />
          <HomeOutlined style={{ fontSize: 24, color: "#eb2f96" }} />
          <HomeOutlined style={{ fontSize: 24, color: "#faad14" }} />
          <HomeOutlined style={{ fontSize: 24, color: "#ff4d4f" }} />
          <HomeOutlined style={{ fontSize: 24, color: "#722ed1" }} />
        </Space>
      </Section>

      <Section title="TwoTone (default primary color)">
        <Space size="large" wrap>
          <EditTwoTone style={{ fontSize: 32 }} />
          <HeartTwoTone style={{ fontSize: 32 }} />
          <CheckCircleTwoTone style={{ fontSize: 32 }} />
        </Space>
      </Section>

      <Section title="TwoTone (custom colors)">
        <Space size="large" wrap>
          <EditTwoTone twoToneColor="#eb2f96" style={{ fontSize: 32 }} />
          <HeartTwoTone twoToneColor="#52c41a" style={{ fontSize: 32 }} />
          <CheckCircleTwoTone twoToneColor="#faad14" style={{ fontSize: 32 }} />
        </Space>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => <HomeOutlined {...args} />,
}

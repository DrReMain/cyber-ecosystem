import type { Meta, StoryObj } from "@storybook/react-vite"
import { ConfigProvider, DatePicker, Flex, Input, Select, Space } from "antd"
import { Label, Section } from "../helpers"

const sizes = ["small", "medium", "large"] as const

const selectOptions = [
  { value: "apple", label: "Apple" },
  { value: "banana", label: "Banana" },
  { value: "cherry", label: "Cherry" },
]

const meta: Meta<typeof ConfigProvider> = {
  title: "Antd/General/ConfigProvider",
  component: ConfigProvider,
  parameters: { layout: "padded" },
  argTypes: {
    componentSize: {
      control: "radio",
      options: ["small", "medium", "large"],
    },
    componentDisabled: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof ConfigProvider>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section
        title="componentSize"
        description="Global component size via ConfigProvider"
      >
        <Flex vertical gap={12}>
          {sizes.map((size) => (
            <Flex key={size} vertical gap={4}>
              <Label>{size}</Label>
              <ConfigProvider componentSize={size}>
                <Space wrap>
                  <Input placeholder="Input" style={{ width: 200 }} />
                  <Select
                    placeholder="Select"
                    options={selectOptions}
                    style={{ width: 200 }}
                  />
                  <DatePicker placeholder="DatePicker" />
                </Space>
              </ConfigProvider>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section
        title="componentDisabled"
        description="Disable all form controls via ConfigProvider"
      >
        <Flex vertical gap={8}>
          <Label>Enabled</Label>
          <Space wrap>
            <Input placeholder="Input" style={{ width: 200 }} />
            <Select
              placeholder="Select"
              options={selectOptions}
              style={{ width: 200 }}
            />
            <DatePicker placeholder="DatePicker" />
          </Space>
          <Label>Disabled (componentDisabled=true)</Label>
          <ConfigProvider componentDisabled>
            <Space wrap>
              <Input placeholder="Input" style={{ width: 200 }} />
              <Select
                placeholder="Select"
                options={selectOptions}
                style={{ width: 200 }}
              />
              <DatePicker placeholder="DatePicker" />
            </Space>
          </ConfigProvider>
        </Flex>
      </Section>

      <Section
        title="Theme Override"
        description="Custom primary color via token"
      >
        <Flex vertical gap={8}>
          <Label>Default theme</Label>
          <Space wrap>
            <Input placeholder="Input" style={{ width: 200 }} />
            <Select
              placeholder="Select"
              options={selectOptions}
              style={{ width: 200 }}
            />
          </Space>
          <Label>Custom primary color (#722ed1)</Label>
          <ConfigProvider
            theme={{
              token: {
                colorPrimary: "#722ed1",
                borderRadius: 4,
              },
            }}
          >
            <Space wrap>
              <Input placeholder="Input" style={{ width: 200 }} />
              <Select
                placeholder="Select"
                options={selectOptions}
                style={{ width: 200 }}
              />
            </Space>
          </ConfigProvider>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: {
    componentSize: "medium",
    componentDisabled: false,
  },
  render: (args) => (
    <ConfigProvider {...args}>
      <Space vertical size={12}>
        <Input placeholder="Input" style={{ width: 300 }} />
        <Select
          placeholder="Select"
          options={selectOptions}
          style={{ width: 300 }}
        />
        <DatePicker placeholder="DatePicker" />
      </Space>
    </ConfigProvider>
  ),
}

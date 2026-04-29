import {
  AppstoreOutlined,
  BarsOutlined,
  CloudOutlined,
  DownloadOutlined,
  FileOutlined,
  SettingOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Segmented, Space } from "antd"
import { expect, fn, within } from "storybook/test"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Segmented> = {
  title: "Antd/Data Display/Segmented",
  component: Segmented,
  parameters: { layout: "padded" },
  args: {
    options: ["Daily", "Weekly", "Monthly", "Quarterly", "Yearly"],
    defaultValue: "Weekly",
  },
  argTypes: {
    block: { control: "boolean" },
    size: sizeArg(["small", "medium", "large"]),
    shape: {
      control: "radio",
      options: ["default", "round"],
    },
    disabled: { control: "boolean" },
    orientation: {
      control: "radio",
      options: ["horizontal", "vertical"],
    },
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Segmented>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes" description="small, medium (default), large">
        <Space orientation="vertical" size={12}>
          {(["small", "medium", "large"] as const).map((size) => (
            <div key={size}>
              <Label>{size}</Label>
              <Segmented
                size={size}
                options={["Map", "Transit", "Satellite"]}
                defaultValue="Map"
              />
            </div>
          ))}
        </Space>
      </Section>

      <Section title="Shapes" description="default vs round">
        <Space orientation="vertical" size={12}>
          <div>
            <Label>shape: default</Label>
            <Segmented
              options={["Map", "Transit", "Satellite"]}
              defaultValue="Map"
            />
          </div>
          <div>
            <Label>shape: round</Label>
            <Segmented
              shape="round"
              options={["Map", "Transit", "Satellite"]}
              defaultValue="Map"
            />
          </div>
        </Space>
      </Section>

      <Section title="Block" description="Fit width to parent">
        <Segmented
          block
          options={["Option A", "Option B", "Option C"]}
          defaultValue="Option A"
        />
      </Section>

      <Section title="Disabled States">
        <Space orientation="vertical" size={12}>
          <div>
            <Label>All disabled</Label>
            <Segmented
              disabled
              options={["Active", "Inactive", "Pending"]}
              defaultValue="Active"
            />
          </div>
          <div>
            <Label>Per-option disabled</Label>
            <Segmented
              options={[
                "Active",
                { label: "Disabled", value: "disabled", disabled: true },
                "Active 2",
                { label: "Also Disabled", value: "also", disabled: true },
              ]}
              defaultValue="Active"
            />
          </div>
        </Space>
      </Section>

      <Section title="With Icons">
        <Space orientation="vertical" size={12}>
          <div>
            <Label>Icon + label</Label>
            <Segmented
              options={[
                { label: "List", value: "list", icon: <BarsOutlined /> },
                { label: "Grid", value: "grid", icon: <AppstoreOutlined /> },
              ]}
              defaultValue="list"
            />
          </div>
          <div>
            <Label>Icon only</Label>
            <Segmented
              options={[
                { value: "file", icon: <FileOutlined /> },
                { value: "cloud", icon: <CloudOutlined /> },
                { value: "download", icon: <DownloadOutlined /> },
                { value: "settings", icon: <SettingOutlined /> },
              ]}
              defaultValue="file"
            />
          </div>
        </Space>
      </Section>

      <Section
        title="Vertical (orientation)"
        description="orientation=vertical"
      >
        <Space size={32} align="start">
          <div>
            <Label>horizontal (default)</Label>
            <Segmented
              options={["Option A", "Option B", "Option C"]}
              defaultValue="Option A"
            />
          </div>
          <div>
            <Label>vertical</Label>
            <Segmented
              orientation="vertical"
              options={["Option A", "Option B", "Option C"]}
              defaultValue="Option A"
            />
          </div>
        </Space>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Segmented
            options={["Map", "Transit", "Satellite"]}
            defaultValue="Map"
          />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    options: ["Daily", "Weekly", "Monthly"],
    defaultValue: "Weekly",
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.getByText("Monthly").click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

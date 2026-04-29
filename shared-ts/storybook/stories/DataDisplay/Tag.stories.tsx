import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  MinusCircleOutlined,
  SyncOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Space, Tag } from "antd"
import { useState } from "react"
import { expect, fn, within } from "storybook/test"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Tag> = {
  title: "Antd/Data Display/Tag",
  component: Tag,
  parameters: { layout: "padded" },
  args: { children: "Tag" },
  argTypes: {
    color: {
      control: "select",
      options: [
        "default",
        "success",
        "processing",
        "error",
        "warning",
        "magenta",
        "red",
        "volcano",
        "orange",
        "gold",
        "lime",
        "green",
        "cyan",
        "blue",
        "geekblue",
        "purple",
      ],
    },
    variant: {
      control: "select",
      options: ["filled", "outlined", "solid"],
    },
    closeIcon: { control: "boolean" },
    disabled: { control: "boolean" },
    onClose: { action: "closed" },
  },
}

export default meta
type Story = StoryObj<typeof Tag>

const presetColors = [
  "magenta",
  "red",
  "volcano",
  "orange",
  "gold",
  "lime",
  "green",
  "cyan",
  "blue",
  "geekblue",
  "purple",
]

const statusColors = [
  "success",
  "processing",
  "error",
  "warning",
  "default",
] as const

const variants = ["filled", "outlined", "solid"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const { CheckableTag, CheckableTagGroup } = Tag
    const [checkedTags, setCheckedTags] = useState<Record<string, boolean>>({
      tag1: true,
      tag2: false,
      tag3: false,
    })

    return (
      <Flex vertical gap={24}>
        <Section title="Colors">
          <Flex vertical gap={8}>
            <Space wrap>
              {presetColors.map((color) => (
                <Tag color={color} key={color}>
                  {color}
                </Tag>
              ))}
            </Space>
            <Space wrap>
              <Tag color="#f50">#f50</Tag>
              <Tag color="#2db7f5">#2db7f5</Tag>
              <Tag color="#87d068">#87d068</Tag>
              <Tag color="#108ee9">#108ee9</Tag>
            </Space>
          </Flex>
        </Section>

        <Section title="Variants by Color">
          <Flex vertical gap={12}>
            {variants.map((variant) => (
              <div key={variant}>
                <Label>{variant}</Label>
                <Space wrap>
                  {presetColors.slice(0, 6).map((color) => (
                    <Tag
                      color={color}
                      variant={variant}
                      key={`${color}-${variant}`}
                    >
                      {color}
                    </Tag>
                  ))}
                </Space>
              </div>
            ))}
          </Flex>
        </Section>

        <Section title="Status Presets">
          <Flex vertical gap={8}>
            <Space wrap>
              {statusColors.map((color) => (
                <Tag color={color} key={color}>
                  {color}
                </Tag>
              ))}
            </Space>
            <Space wrap>
              <Tag icon={<SyncOutlined spin />} color="processing">
                Processing
              </Tag>
              <Tag icon={<CheckCircleOutlined />} color="success">
                Success
              </Tag>
              <Tag icon={<CloseCircleOutlined />} color="error">
                Error
              </Tag>
              <Tag icon={<ExclamationCircleOutlined />} color="warning">
                Warning
              </Tag>
              <Tag icon={<ClockCircleOutlined />} color="default">
                Waiting
              </Tag>
              <Tag icon={<MinusCircleOutlined />} color="default">
                Stopped
              </Tag>
            </Space>
          </Flex>
        </Section>

        <Section title="Close Icon">
          <Space wrap>
            <Tag closeIcon>Default</Tag>
            {presetColors.slice(0, 5).map((color) => (
              <Tag closeIcon color={color} key={`closable-${color}`}>
                {color}
              </Tag>
            ))}
          </Space>
        </Section>

        <Section title="Outlined Variant (borderless)">
          <Flex vertical gap={12}>
            <div>
              <Label>outlined (no fill, border only)</Label>
              <Space wrap>
                {presetColors.slice(0, 6).map((color) => (
                  <Tag
                    color={color}
                    variant="outlined"
                    key={`outlined-${color}`}
                  >
                    {color}
                  </Tag>
                ))}
              </Space>
            </div>
          </Flex>
        </Section>

        <Section title="CheckableTag">
          <Space wrap>
            {(["tag1", "tag2", "tag3"] as const).map((key) => (
              <CheckableTag
                key={key}
                checked={checkedTags[key]}
                onChange={(checked) =>
                  setCheckedTags((prev) => ({ ...prev, [key]: checked }))
                }
              >
                {key} {checkedTags[key] ? "(checked)" : "(unchecked)"}
              </CheckableTag>
            ))}
          </Space>
        </Section>

        <Section title="CheckableTagGroup (single)">
          <CheckableTagGroup
            options={[
              { label: "Option A", value: "a" },
              { label: "Option B", value: "b" },
              { label: "Option C", value: "c" },
            ]}
            defaultValue="a"
          />
        </Section>

        <Section title="States">
          <PseudoStates>
            <Tag>Tag</Tag>
          </PseudoStates>
        </Section>
      </Flex>
    )
  },
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    closeIcon: true,
    children: "Closable Tag",
    onClose: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const closeBtn = canvas.getByRole("img", { name: /close/i })
    await closeBtn.click()
    await expect(args.onClose).toHaveBeenCalled()
  },
}

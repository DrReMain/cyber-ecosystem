import {
  CustomerServiceOutlined,
  FileTextOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
  SyncOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, FloatButton } from "antd"
import { expect, fn, within } from "storybook/test"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof FloatButton> = {
  title: "Antd/General/FloatButton",
  component: FloatButton,
  parameters: { layout: "padded" },
  args: {
    type: "default",
    shape: "circle",
  },
  argTypes: {
    type: {
      control: "radio",
      options: ["default", "primary"],
    },
    shape: {
      control: "radio",
      options: ["circle", "square"],
    },
    tooltip: { control: "text" },
    href: { control: "text" },
    target: { control: "text" },
    onClick: { action: "clicked" },
  },
}

export default meta
type Story = StoryObj<typeof FloatButton>

const types = ["default", "primary"] as const
const shapes = ["circle", "square"] as const

/** Spread FloatButtons horizontally so they don't overlap at bottom-right. */
const at = (i: number, gap = 64) => ({ style: { right: 32 + i * gap } })

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <>
      <Section title="Type x Shape">
        <Flex vertical gap={8}>
          {types.map((type) => (
            <Flex key={type} vertical gap={4}>
              <Label>{type}</Label>
              <Label>{shapes.join(" / ")}</Label>
            </Flex>
          ))}
        </Flex>
      </Section>
      {types.flatMap((type, ti) =>
        shapes.map((shape, si) => (
          <FloatButton
            key={`${type}-${shape}`}
            type={type}
            shape={shape}
            icon={<CustomerServiceOutlined />}
            tooltip={`${type} ${shape}`}
            {...at(ti * 2 + si)}
          />
        )),
      )}

      <Section title="Badge Variants">
        <Label>count: 5 / dot / count: 99</Label>
      </Section>
      <FloatButton
        icon={<CustomerServiceOutlined />}
        badge={{ count: 5 }}
        {...at(0)}
      />
      <FloatButton
        icon={<CustomerServiceOutlined />}
        badge={{ dot: true }}
        {...at(1)}
      />
      <FloatButton
        type="primary"
        icon={<CustomerServiceOutlined />}
        badge={{ count: 99 }}
        {...at(2)}
      />

      <Section title="Groups (circle)">
        <Label>Left: hover trigger -- Right: click trigger</Label>
      </Section>
      <FloatButton.Group
        trigger="hover"
        type="primary"
        icon={<CustomerServiceOutlined />}
        {...at(0, 120)}
      >
        <FloatButton icon={<QuestionCircleOutlined />} tooltip="Help" />
        <FloatButton icon={<FileTextOutlined />} tooltip="Docs" />
        <FloatButton icon={<SettingOutlined />} tooltip="Settings" />
      </FloatButton.Group>
      <FloatButton.Group
        trigger="click"
        icon={<CustomerServiceOutlined />}
        {...at(1, 120)}
      >
        <FloatButton icon={<QuestionCircleOutlined />} tooltip="Help" />
        <FloatButton icon={<FileTextOutlined />} tooltip="Docs" />
        <FloatButton icon={<SettingOutlined />} tooltip="Settings" />
      </FloatButton.Group>

      <Section title="Groups (square)">
        <Label>Left: default type -- Right: primary type</Label>
      </Section>
      <FloatButton.Group
        shape="square"
        trigger="hover"
        icon={<CustomerServiceOutlined />}
        {...at(0, 120)}
      >
        <FloatButton icon={<QuestionCircleOutlined />} tooltip="Help" />
        <FloatButton icon={<FileTextOutlined />} tooltip="Docs" />
      </FloatButton.Group>
      <FloatButton.Group
        shape="square"
        type="primary"
        trigger="hover"
        icon={<CustomerServiceOutlined />}
        {...at(1, 120)}
      >
        <FloatButton icon={<QuestionCircleOutlined />} tooltip="Help" />
        <FloatButton icon={<FileTextOutlined />} tooltip="Docs" />
      </FloatButton.Group>
    </>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <FloatButton {...args} icon={args.icon ?? <CustomerServiceOutlined />} />
  ),
}

export const StaticOpen: Story = {
  parameters: { layout: "fullscreen" },
  render: () => (
    <>
      <div style={{ padding: "60px 24px", minHeight: "100vh" }}>
        <PageContent>
          <Label>
            The FloatButton.Group is fixed at the bottom-right. Scroll to verify
            positioning.
          </Label>
          <div style={{ height: 1200 }} />
        </PageContent>
      </div>
      <FloatButton.Group
        open
        trigger="click"
        type="primary"
        icon={<CustomerServiceOutlined />}
      >
        <FloatButton icon={<QuestionCircleOutlined />} tooltip="Help" />
        <FloatButton icon={<FileTextOutlined />} tooltip="Docs" />
        <FloatButton icon={<SettingOutlined />} tooltip="Settings" />
      </FloatButton.Group>
    </>
  ),
}

export const Interactive: Story = {
  args: {
    type: "primary",
    shape: "circle",
    icon: <SyncOutlined />,
    tooltip: "Click me",
    onClick: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement.ownerDocument.body)
    const btn = canvas.getByRole("button")
    await btn.click()
    await expect(args.onClick).toHaveBeenCalledTimes(1)
  },
}

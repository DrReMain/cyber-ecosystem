import { DownOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import type { MenuProps } from "antd"
import { Button, Dropdown, Flex, Space } from "antd"
import { useRef } from "react"
import { expect, fn, within } from "storybook/test"
import { disabledArg } from "../argTypes"
import { PageContent, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const menuItems: MenuProps["items"] = [
  { key: "1", label: "Item 1" },
  { key: "2", label: "Item 2" },
  { key: "3", label: "Item 3" },
  { type: "divider" },
  { key: "4", label: "Disabled", disabled: true },
]

const meta: Meta<typeof Dropdown> = {
  title: "Antd/Navigation/Dropdown",
  component: Dropdown,
  parameters: { layout: "padded" },
  args: {
    menu: { items: menuItems },
  },
  argTypes: {
    trigger: {
      control: "check",
      options: ["hover", "click", "contextMenu"],
    },
    placement: {
      control: "select",
      options: [
        "topLeft",
        "top",
        "topRight",
        "bottomLeft",
        "bottom",
        "bottomRight",
      ],
    },
    arrow: { control: "boolean" },
    disabled: disabledArg,
  },
}

export default meta
type Story = StoryObj<typeof Dropdown>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Dropdown menu={{ items: menuItems }}>
          <Button>
            Hover me <DownOutlined />
          </Button>
        </Dropdown>
      </Section>
      <Section title="With Arrow">
        <Space>
          <Dropdown menu={{ items: menuItems }} arrow>
            <Button>With Arrow</Button>
          </Dropdown>
          <Dropdown menu={{ items: menuItems }} arrow={{ pointAtCenter: true }}>
            <Button>Arrow at Center</Button>
          </Dropdown>
        </Space>
      </Section>
      <Section title="Disabled">
        <Dropdown menu={{ items: menuItems }} disabled>
          <Button>Disabled Dropdown</Button>
        </Dropdown>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Dropdown
            menu={{
              items: [
                { key: "1", label: "Item 1" },
                { key: "2", label: "Item 2" },
              ],
            }}
          >
            <Button>Hover me</Button>
          </Dropdown>
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Dropdown {...args}>
      <Button>
        Hover me <DownOutlined />
      </Button>
    </Dropdown>
  ),
}

// ── StaticOpen ────────────────────────────────────────────────

export const StaticOpen: Story = {
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div ref={containerRef} style={{ position: "relative", minHeight: 220 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <Dropdown
            open
            menu={{
              items: [
                { key: "1", label: "Regular Item 1" },
                { key: "2", label: "Regular Item 2" },
                { key: "3", label: "Regular Item 3" },
                { type: "divider" },
                { key: "4", label: "Disabled Item", disabled: true },
                { type: "divider" },
                {
                  key: "sub",
                  label: "Submenu",
                  children: [
                    { key: "sub-1", label: "Sub Item 1" },
                    { key: "sub-2", label: "Sub Item 2" },
                    { key: "sub-3", label: "Sub Item 3" },
                  ],
                },
                { key: "5", label: "Last Item" },
              ],
            }}
            getPopupContainer={() => containerRef.current!}
          >
            <Button>
              Static Open <DownOutlined />
            </Button>
          </Dropdown>
        </div>
      </div>
    )
  },
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    menu: { items: menuItems },
    trigger: ["click"],
    onOpenChange: fn(),
  },
  render: (args) => (
    <Dropdown {...args}>
      <Button>
        Click me <DownOutlined />
      </Button>
    </Dropdown>
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const trigger = canvas.getByRole("button")
    await trigger.click()
    await expect(args.onOpenChange).toHaveBeenCalled()
  },
}

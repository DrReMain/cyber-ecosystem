import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, TreeSelect } from "antd"
import { useRef } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { disabledArg, sizeArg, statusArg, variantArg } from "../argTypes"
import type { TreeNodeType } from "../fixtures"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof TreeSelect> = {
  title: "Antd/Data Entry/TreeSelect",
  component: TreeSelect,
  parameters: { layout: "padded" },
  args: {
    style: { width: 280 },
    placeholder: "Select a node",
  },
  argTypes: {
    multiple: { control: "boolean" },
    showSearch: { control: "boolean" },
    disabled: disabledArg,
    treeCheckable: { control: "boolean" },
    status: statusArg,
    size: sizeArg(["small", "medium", "large"]),
    variant: variantArg(),
    onChange: { action: "changed" },
    onOpenChange: { action: "openChanged" },
  },
}

export default meta
type Story = StoryObj<typeof TreeSelect>

const treeData: TreeNodeType[] = [
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
      {
        title: "Sub Node 3-3",
        value: "3-3",
        children: [
          { title: "Leaf 3-3-1", value: "3-3-1" },
          { title: "Leaf 3-3-2", value: "3-3-2" },
        ],
      },
    ],
  },
]

const disabledTreeData = [
  { title: "Available", value: "available" },
  { title: "Disabled node", value: "disabled", disabled: true },
  { title: "Also available", value: "also-available" },
]

const w = { width: 280 }

const sizes = ["small", "medium", "large"] as const
const variants = ["outlined", "filled", "borderless", "underlined"] as const
const statuses = ["error", "warning"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Sizes">
        <Flex gap={16} align="center">
          {sizes.map((size) => (
            <TreeSelect
              key={size}
              treeData={treeData}
              size={size}
              placeholder={size}
              style={w}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Variants">
        <Flex gap={16} align="center">
          {variants.map((variant) => (
            <TreeSelect
              key={variant}
              treeData={treeData}
              variant={variant}
              placeholder={variant}
              style={w}
            />
          ))}
        </Flex>
      </Section>

      <Section title="Status">
        <Flex vertical gap={12}>
          {statuses.map((status) => (
            <Flex key={status} gap={16} align="center">
              <Label>{status}</Label>
              <TreeSelect
                treeData={treeData}
                status={status}
                placeholder={status}
                style={w}
              />
              <TreeSelect
                treeData={treeData}
                status={status}
                multiple
                placeholder={`${status} multiple`}
                style={{ width: 350 }}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex vertical gap={12}>
          <Flex gap={16} align="center">
            <Label>Disabled select</Label>
            <TreeSelect treeData={treeData} disabled style={w} />
          </Flex>
          <Flex gap={16} align="center">
            <Label>Disabled option</Label>
            <TreeSelect treeData={disabledTreeData} style={w} />
          </Flex>
        </Flex>
      </Section>

      <Section title="Multiple">
        <Flex vertical gap={12}>
          <TreeSelect
            treeData={treeData}
            multiple
            defaultValue={["1-1", "2-1"]}
            style={{ width: "100%" }}
            placeholder="Select multiple nodes"
            treeDefaultExpandAll
          />
        </Flex>
      </Section>

      <Section title="treeCheckable">
        <Flex vertical gap={12}>
          <Flex vertical gap={4}>
            <Label>Checkable (default SHOW_CHILD)</Label>
            <TreeSelect
              treeData={treeData}
              treeCheckable
              treeDefaultExpandAll
              style={{ width: "100%" }}
              placeholder="Check nodes"
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>showCheckedStrategy=SHOW_ALL</Label>
            <TreeSelect
              treeData={treeData}
              treeCheckable
              treeDefaultExpandAll
              showCheckedStrategy={TreeSelect.SHOW_ALL}
              style={{ width: "100%" }}
              placeholder="Show all checked"
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>showCheckedStrategy=SHOW_PARENT</Label>
            <TreeSelect
              treeData={treeData}
              treeCheckable
              treeDefaultExpandAll
              showCheckedStrategy={TreeSelect.SHOW_PARENT}
              style={{ width: "100%" }}
              placeholder="Show parent only"
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>treeCheckStrictly</Label>
            <TreeSelect
              treeData={treeData}
              treeCheckable
              treeCheckStrictly
              treeDefaultExpandAll
              style={{ width: "100%" }}
              placeholder="Strictly check (no parent-child relation)"
            />
          </Flex>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { treeData },
}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const containerRef = useRef<HTMLDivElement>(null)
    return (
      <div ref={containerRef} style={{ position: "relative", minHeight: 400 }}>
        <PageContent />
        <div style={{ marginTop: 24 }}>
          <TreeSelect
            open
            treeData={treeData}
            treeDefaultExpandAll
            placeholder="Static open tree select"
            style={w}
            getPopupContainer={() => containerRef.current ?? document.body}
          />
        </div>
      </div>
    )
  },
}

export const Interactive: Story = {
  args: {
    treeData,
    defaultValue: "1-1",
    onChange: fn(),
    onOpenChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const trigger = canvas.getByRole("combobox")
    await userEvent.click(trigger)
    await expect(args.onOpenChange).toHaveBeenCalled()
  },
}

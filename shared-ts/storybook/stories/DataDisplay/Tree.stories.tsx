import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Tree } from "antd"
import { expect, fn, within } from "storybook/test"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Tree> = {
  title: "Antd/Data Display/Tree",
  component: Tree,
  parameters: { layout: "padded" },
  args: {
    defaultExpandAll: false,
    checkable: false,
    showLine: false,
    showIcon: false,
    blockNode: false,
    draggable: false,
    selectable: true,
    multiple: false,
    disabled: false,
  },
  argTypes: {
    checkable: { control: "boolean" },
    defaultExpandAll: { control: "boolean" },
    draggable: { control: "boolean" },
    showLine: { control: "boolean" },
    showIcon: { control: "boolean" },
    blockNode: { control: "boolean" },
    selectable: { control: "boolean" },
    multiple: { control: "boolean" },
    disabled: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Tree>

const treeData = [
  {
    title: "Parent 1",
    key: "0-0",
    value: "0-0",
    children: [
      {
        title: "Child 1-1",
        key: "0-0-0",
        value: "0-0-0",
        children: [
          { title: "Leaf 1-1-1", key: "0-0-0-0", value: "0-0-0-0" },
          { title: "Leaf 1-1-2", key: "0-0-0-1", value: "0-0-0-1" },
        ],
      },
      {
        title: "Child 1-2",
        key: "0-0-1",
        value: "0-0-1",
        children: [{ title: "Leaf 1-2-1", key: "0-0-1-0", value: "0-0-1-0" }],
      },
    ],
  },
  {
    title: "Parent 2",
    key: "0-1",
    value: "0-1",
    children: [
      { title: "Child 2-1", key: "0-1-0", value: "0-1-0" },
      { title: "Child 2-2", key: "0-1-1", value: "0-1-1" },
    ],
  },
]

const disabledData = [
  {
    title: "Parent 1",
    key: "0-0",
    value: "0-0",
    children: [
      { title: "Disabled Node", key: "0-0-0", value: "0-0-0", disabled: true },
      { title: "Normal Node", key: "0-0-1", value: "0-0-1" },
      { title: "Disable Checkbox", key: "0-0-2", value: "0-0-2", disableCheckbox: true },
    ],
  },
]

const directoryData = [
  {
    title: "src",
    key: "src",
    value: "src",
    children: [
      {
        title: "components",
        key: "src/components",
        value: "src/components",
        children: [
          {
            title: "Button.tsx",
            key: "src/components/Button.tsx",
            value: "src/components/Button.tsx",
            isLeaf: true,
          },
          { title: "Input.tsx", key: "src/components/Input.tsx", value: "src/components/Input.tsx", isLeaf: true },
        ],
      },
      {
        title: "pages",
        key: "src/pages",
        value: "src/pages",
        children: [
          { title: "index.tsx", key: "src/pages/index.tsx", value: "src/pages/index.tsx", isLeaf: true },
          { title: "about.tsx", key: "src/pages/about.tsx", value: "src/pages/about.tsx", isLeaf: true },
        ],
      },
      { title: "App.tsx", key: "src/App.tsx", value: "src/App.tsx", isLeaf: true },
    ],
  },
  {
    title: "public",
    key: "public",
    value: "public",
    children: [{ title: "index.html", key: "public/index.html", value: "public/index.html", isLeaf: true }],
  },
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Default Expand All">
        <Tree treeData={treeData} defaultExpandAll />
      </Section>

      <Section title="Checkable">
        <Tree
          treeData={treeData}
          checkable
          defaultCheckedKeys={["0-0-0-0"]}
          defaultExpandAll
        />
      </Section>

      <Section title="Directory Tree">
        <Tree.DirectoryTree treeData={directoryData} defaultExpandAll />
      </Section>

      <Section title="Show Line">
        <Flex vertical gap={12}>
          <div>
            <Label>showLine: true</Label>
            <Tree treeData={treeData} showLine defaultExpandAll />
          </div>
          <div>
            <Label>showLine with showLeafIcon: false</Label>
            <Tree
              treeData={treeData}
              showLine={{ showLeafIcon: false }}
              defaultExpandAll
            />
          </div>
        </Flex>
      </Section>

      <Section title="Draggable">
        <Tree treeData={treeData} draggable defaultExpandAll />
      </Section>

      <Section title="Disabled Nodes">
        <Tree treeData={disabledData} defaultExpandAll />
      </Section>

      <Section title="Virtual Scroll (height)">
        <Tree
          treeData={Array.from({ length: 100 }, (_, i) => ({
            title: `Node ${i + 1}`,
            key: `node-${i}`,
            children: [
              { title: `Child ${i + 1}-1`, key: `node-${i}-1` },
              { title: `Child ${i + 1}-2`, key: `node-${i}-2` },
            ],
          }))}
          height={200}
          defaultExpandedKeys={["node-0", "node-1", "node-2"]}
        />
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => <Tree {...args} treeData={treeData} />,
}

export const Interactive: Story = {
  args: {
    treeData,
    defaultExpandAll: true,
    checkable: true,
    onCheck: fn(),
    onSelect: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const checkboxes = canvas.getAllByRole("checkbox")
    await checkboxes[0].click()
    await expect(args.onCheck).toHaveBeenCalled()
  },
}

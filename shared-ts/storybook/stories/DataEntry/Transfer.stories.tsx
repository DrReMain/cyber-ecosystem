import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Transfer } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { disabledArg, statusArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Transfer> = {
  title: "Antd/Data Entry/Transfer",
  component: Transfer,
  parameters: { layout: "padded" },
  args: {
    render: (item) => item.title ?? "",
  },
  argTypes: {
    disabled: disabledArg,
    showSearch: { control: "boolean" },
    oneWay: { control: "boolean" },
    pagination: { control: "object" },
    status: statusArg,
    onChange: { action: "changed" },
    onSelectChange: { action: "selectedChanged" },
  },
}

export default meta
type Story = StoryObj<typeof Transfer>

interface RecordType {
  key: string
  title: string
  description: string
}

const dataSource: RecordType[] = Array.from({ length: 20 }).map((_, i) => ({
  key: String(i + 1),
  title: `Item ${i + 1}`,
  description: `Description for item ${i + 1}`,
}))

const smallDataSource: RecordType[] = Array.from({ length: 10 }).map(
  (_, i) => ({
    key: String(i + 1),
    title: `Item ${i + 1}`,
    description: `Desc ${i + 1}`,
  }),
)

const basicTargetKeys = ["3", "5", "7", "10"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Transfer
          dataSource={dataSource}
          targetKeys={basicTargetKeys}
          render={(item) => item.title}
        />
      </Section>

      <Section title="With Search">
        <Transfer
          dataSource={dataSource}
          targetKeys={basicTargetKeys}
          showSearch
          render={(item) => item.title}
        />
      </Section>

      <Section title="Disabled">
        <Transfer
          dataSource={dataSource}
          targetKeys={basicTargetKeys}
          disabled
          render={(item) => item.title}
        />
      </Section>

      <Section title="Status">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>Error</Label>
            <Transfer
              dataSource={smallDataSource}
              targetKeys={["3", "5"]}
              status="error"
              render={(item) => item.title}
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>Warning</Label>
            <Transfer
              dataSource={smallDataSource}
              targetKeys={["3", "5"]}
              status="warning"
              render={(item) => item.title}
            />
          </Flex>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: {
    dataSource: smallDataSource,
    targetKeys: [],
  },
}

export const Interactive: Story = {
  args: {
    dataSource: smallDataSource.slice(0, 5),
    targetKeys: [],
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const checkboxes = canvas.getAllByRole("checkbox")
    if (checkboxes.length > 0) {
      await userEvent.click(checkboxes[0])
    }
    const buttons = canvas.getAllByRole("button")
    const moveRightBtn = buttons.find((b) =>
      b.querySelector('[aria-label="right"]'),
    )
    if (moveRightBtn) {
      await userEvent.click(moveRightBtn)
    }
    await expect(args.onChange).toHaveBeenCalled()
  },
}

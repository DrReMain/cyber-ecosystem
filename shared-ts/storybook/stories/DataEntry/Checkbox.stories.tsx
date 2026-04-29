import type { Meta, StoryObj } from "@storybook/react-vite"
import { Checkbox, Flex } from "antd"
import { useState } from "react"
import { expect, fn, within } from "storybook/test"
import { callbackArgs, disabledArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Checkbox> = {
  title: "Antd/Data Entry/Checkbox",
  component: Checkbox,
  parameters: { layout: "padded" },
  args: { children: "Checkbox" },
  argTypes: {
    disabled: disabledArg,
    ...callbackArgs,
  },
}

export default meta
type Story = StoryObj<typeof Checkbox>

const states = [
  { label: "Unchecked", checked: false, disabled: false },
  { label: "Checked", checked: true, disabled: false },
  { label: "Disabled Unchecked", checked: false, disabled: true },
  { label: "Disabled Checked", checked: true, disabled: true },
] as const

const groupOptionTypes: {
  label: string
  options: (string | { label: string; value: string; disabled?: boolean })[]
  defaultValue: string[]
}[] = [
  {
    label: "String options",
    options: ["Option A", "Option B", "Option C", "Option D"],
    defaultValue: ["Option A"],
  },
  {
    label: "Object options with disabled",
    options: [
      { label: "Apple", value: "apple" },
      { label: "Pear", value: "pear" },
      { label: "Orange", value: "orange", disabled: true },
    ],
    defaultValue: ["pear"],
  },
]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [indeterminate, setIndeterminate] = useState(true)
    const [checkAll, setCheckAll] = useState(false)
    const checkAllOptions = ["Apple", "Pear", "Orange"]
    const [checkedList, setCheckedList] = useState<string[]>(["Apple"])

    const onCheckAllChange = (e: { target: { checked: boolean } }) => {
      setCheckedList(e.target.checked ? checkAllOptions : [])
      setIndeterminate(false)
      setCheckAll(e.target.checked)
    }

    return (
      <Flex vertical gap={24}>
        <Section title="Basic States">
          <Flex gap={24} wrap>
            {states.map((s) => (
              <Checkbox
                key={s.label}
                defaultChecked={s.checked}
                disabled={s.disabled}
              >
                {s.label}
              </Checkbox>
            ))}
          </Flex>
        </Section>

        <Section title="Indeterminate">
          <Flex vertical gap={12}>
            <Checkbox
              indeterminate={indeterminate}
              checked={checkAll}
              onChange={onCheckAllChange}
            >
              Check all
            </Checkbox>
            <Checkbox.Group
              options={checkAllOptions}
              value={checkedList}
              onChange={(list) => {
                setCheckedList(list)
                setIndeterminate(
                  !!list.length && list.length < checkAllOptions.length,
                )
                setCheckAll(list.length === checkAllOptions.length)
              }}
            />
          </Flex>
        </Section>

        <Section title="Checkbox Group">
          <Flex vertical gap={16}>
            {groupOptionTypes.map((group) => (
              <Flex key={group.label} vertical gap={8}>
                <Label>{group.label}</Label>
                <Checkbox.Group
                  options={group.options}
                  defaultValue={group.defaultValue}
                />
              </Flex>
            ))}
          </Flex>
        </Section>

        <Section title="States">
          <PseudoStates>
            <Checkbox>Checkbox</Checkbox>
          </PseudoStates>
        </Section>
      </Flex>
    )
  },
}

export const Playground: Story = {
  args: { children: "Checkbox" },
}

export const Interactive: Story = {
  args: {
    children: "Toggle me",
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const checkbox = canvas.getByRole("checkbox")
    await checkbox.click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

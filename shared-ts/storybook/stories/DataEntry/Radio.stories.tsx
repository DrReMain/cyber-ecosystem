import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Radio } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Radio> = {
  title: "Antd/Data Entry/Radio",
  component: Radio,
  parameters: { layout: "padded" },
  args: { children: "Radio" },
  argTypes: {
    disabled: disabledArg,
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Radio>

const fruits = ["Apple", "Pear", "Orange"]
const fruitOptions = [
  { label: "Apple", value: "apple" },
  { label: "Pear", value: "pear" },
  { label: "Orange", value: "orange", disabled: true },
]
const sizes = ["large", "medium", "small"] as const
const buttonStyles = ["outline", "solid"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Flex vertical gap={8}>
          <Radio>Unchecked</Radio>
          <Radio defaultChecked>Checked</Radio>
          <Radio disabled>Disabled</Radio>
          <Radio disabled defaultChecked>
            Disabled Checked
          </Radio>
        </Flex>
      </Section>

      <Section title="Button Style">
        <Flex vertical gap={12}>
          {buttonStyles.map((bs) => (
            <Flex key={bs} align="center" gap={8}>
              <Label>{bs}</Label>
              <Radio.Group
                optionType="button"
                buttonStyle={bs}
                options={fruits}
                defaultValue="Apple"
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Sizes (Button Style)">
        <Flex vertical gap={12}>
          {sizes.map((size) => (
            <Flex key={size} align="center" gap={8}>
              <Label>{size}</Label>
              <Radio.Group
                size={size}
                optionType="button"
                options={fruits}
                defaultValue="Apple"
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Sizes (Default Style)">
        <Flex vertical gap={12}>
          {sizes.map((size) => (
            <Flex key={size} align="center" gap={8}>
              <Label>{size}</Label>
              <Radio.Group
                size={size}
                optionType="default"
                options={fruits}
                defaultValue="Apple"
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled Group">
        <Flex vertical gap={12}>
          <Label>Default style</Label>
          <Radio.Group disabled options={fruitOptions} defaultValue="apple" />
          <Label>Button style</Label>
          <Radio.Group
            disabled
            optionType="button"
            options={fruits}
            defaultValue="Apple"
          />
        </Flex>
      </Section>

      <Section
        title="Button Style x Size"
        description="Radio.Button in different style and size combinations"
      >
        <Flex vertical gap={12}>
          {(["outline", "solid"] as const).map((buttonStyle) => (
            <Flex key={buttonStyle} vertical gap={8}>
              <Label>{buttonStyle}</Label>
              <Flex gap={8}>
                {(["small", "medium", "large"] as const).map((size) => (
                  <Radio.Group
                    key={size}
                    buttonStyle={buttonStyle}
                    size={size}
                    optionType="button"
                    options={[
                      { value: "a", label: "A" },
                      { value: "b", label: "B" },
                      { value: "c", label: "C" },
                    ]}
                    defaultValue="a"
                  />
                ))}
              </Flex>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Radio>Radio</Radio>
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { children: "Radio" },
}

export const Interactive: Story = {
  args: {
    onChange: fn(),
  },
  render: (args) => (
    <Radio.Group
      options={fruits}
      defaultValue="Apple"
      onChange={args.onChange}
    />
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const radios = canvas.getAllByRole("radio")
    await radios[1].click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

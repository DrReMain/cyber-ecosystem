import type { Meta, StoryObj } from "@storybook/react-vite"
import { ConfigProvider, Flex, Slider } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Slider> = {
  title: "Antd/Data Entry/Slider",
  component: Slider,
  parameters: { layout: "padded" },
  args: { defaultValue: 30, min: 0, max: 100, step: 1 },
  argTypes: {
    range: { control: "boolean" },
    min: { control: { type: "number", min: -1000, max: 1000 } },
    max: { control: { type: "number", min: 0, max: 10000 } },
    step: { control: { type: "number", min: 0.1, max: 100 } },
    disabled: disabledArg,
    marks: { control: "object" },
    vertical: { control: "boolean" },
    tooltip: { control: "object" },
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Slider>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Flex vertical gap={12} style={{ width: 400 }}>
          <Label>Default (0)</Label>
          <Slider defaultValue={0} />
          <Label>WithValue (30)</Label>
          <Slider defaultValue={30} />
          <Label>WithValue (70)</Label>
          <Slider defaultValue={70} />
        </Flex>
      </Section>

      <Section title="Range">
        <Flex vertical gap={12} style={{ width: 400 }}>
          <Label>Range [20, 50]</Label>
          <Slider range defaultValue={[20, 50]} />
          <Label>Range [0, 100] full</Label>
          <Slider range defaultValue={[0, 100]} />
        </Flex>
      </Section>

      <Section title="Marks">
        <Flex vertical gap={12} style={{ width: 400 }}>
          <Label>With marks</Label>
          <Slider
            marks={{ 0: "0°C", 26: "26°C", 37: "37°C", 100: "100°C" }}
            defaultValue={37}
          />
          <Label>Marks + included=false</Label>
          <Slider
            included={false}
            marks={{ 0: "0", 25: "25", 50: "50", 75: "75", 100: "100" }}
            defaultValue={50}
          />
          <Label>Marks + dots (snap to marks only)</Label>
          <Slider
            dots
            marks={{
              0: "0",
              20: "20",
              40: "40",
              60: "60",
              80: "80",
              100: "100",
            }}
            defaultValue={40}
          />
          <Label>Marks + step=null (only marks are valid)</Label>
          <Slider
            step={null}
            marks={{ 0: "0", 26: "26", 37: "37°C", 100: "100" }}
            defaultValue={26}
          />
        </Flex>
      </Section>

      <Section title="Sizes (via ConfigProvider)">
        <Flex vertical gap={12} style={{ width: 400 }}>
          {(["small", "medium", "large"] as const).map((size) => (
            <Flex key={size} vertical gap={4}>
              <Label>{size}</Label>
              <ConfigProvider componentSize={size}>
                <Slider defaultValue={50} />
              </ConfigProvider>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Slider defaultValue={30} style={{ width: 200 }} />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { defaultValue: 30 },
}

export const Interactive: Story = {
  args: {
    defaultValue: 30,
    onChange: fn(),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const slider = canvas.getByRole("slider")
    slider.focus()
    await expect(slider).toHaveAttribute("aria-valuenow", "30")
  },
}

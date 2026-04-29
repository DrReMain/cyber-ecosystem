import type { Meta, StoryObj } from "@storybook/react-vite"
import { Card, Flex, Switch, Typography } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg, loadingArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Switch> = {
  title: "Antd/Data Entry/Switch",
  component: Switch,
  parameters: { layout: "padded" },
  args: { defaultChecked: false },
  argTypes: {
    checked: { control: "boolean" },
    disabled: disabledArg,
    loading: loadingArg,
    size: sizeArg(["small", "medium"]),
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Switch>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Flex gap={16} align="center">
          <Switch />
          <Switch defaultChecked />
        </Flex>
      </Section>

      <Section title="Sizes">
        <Flex vertical gap={12}>
          {(["small", "medium"] as const).map((size) => (
            <Flex gap={16} align="center" key={size}>
              <Label>{size}</Label>
              <Switch size={size} />
              <Switch size={size} defaultChecked />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex gap={16} align="center">
          <Switch disabled />
          <Switch disabled defaultChecked />
        </Flex>
      </Section>

      <Section title="Loading">
        <Flex gap={16} align="center">
          <Switch loading />
          <Switch loading defaultChecked />
        </Flex>
      </Section>

      <Section title="With Text (checkedChildren / unCheckedChildren)">
        <Flex vertical gap={12}>
          <Flex gap={16} align="center">
            <Label>On/Off</Label>
            <Switch
              checkedChildren="On"
              unCheckedChildren="Off"
              defaultChecked
            />
            <Switch checkedChildren="On" unCheckedChildren="Off" />
          </Flex>
          <Flex gap={16} align="center">
            <Label>1/0</Label>
            <Switch checkedChildren="1" unCheckedChildren="0" defaultChecked />
            <Switch checkedChildren="1" unCheckedChildren="0" />
          </Flex>
          <Flex gap={16} align="center">
            <Label>Icons</Label>
            <Switch checkedChildren="✓" unCheckedChildren="✗" defaultChecked />
            <Switch checkedChildren="✓" unCheckedChildren="✗" />
          </Flex>
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Switch />
        </PseudoStates>
      </Section>

      <Section
        title="Settings Page"
        description="Switches in a realistic preferences panel layout"
      >
        <Card style={{ maxWidth: 500 }}>
          <Flex vertical gap={20}>
            {[
              {
                label: "Enable notifications",
                desc: "Receive push notifications for important updates",
                defaultChecked: true,
              },
              {
                label: "Dark mode",
                desc: "Use dark color scheme across the interface",
                defaultChecked: false,
              },
              {
                label: "Auto-save drafts",
                desc: "Automatically save form progress every 30 seconds",
                defaultChecked: true,
              },
              {
                label: "Marketing emails",
                desc: "Receive product updates and promotional offers",
                defaultChecked: false,
              },
              {
                label: "Two-factor authentication",
                desc: "Add an extra layer of security to your account",
                defaultChecked: true,
              },
            ].map((item) => (
              <Flex key={item.label} justify="space-between" align="center">
                <Flex vertical gap={2}>
                  <Typography.Text style={{ fontSize: 14 }}>
                    {item.label}
                  </Typography.Text>
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {item.desc}
                  </Typography.Text>
                </Flex>
                <Switch defaultChecked={item.defaultChecked} />
              </Flex>
            ))}
          </Flex>
        </Card>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  args: { defaultChecked: false },
}

export const Interactive: Story = {
  args: {
    defaultChecked: false,
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const switchEl = canvas.getByRole("switch")
    await switchEl.click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

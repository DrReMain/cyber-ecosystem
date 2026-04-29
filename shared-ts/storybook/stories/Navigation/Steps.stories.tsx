import {
  SmileOutlined,
  SolutionOutlined,
  UserOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Steps } from "antd"
import { expect, fn, within } from "storybook/test"
import { sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const basicItems = [
  { title: "Finished", content: "This is a description." },
  { title: "In Progress", content: "This is a description." },
  { title: "Waiting", content: "This is a description." },
]

const types = ["default", "dot", "inline", "navigation"] as const

const meta: Meta<typeof Steps> = {
  title: "Antd/Navigation/Steps",
  component: Steps,
  parameters: { layout: "padded" },
  args: {
    current: 1,
    items: basicItems,
  },
  argTypes: {
    type: {
      control: "select",
      options: ["default", "dot", "inline", "navigation"],
    },
    status: {
      control: "radio",
      options: ["process", "error", "wait", "finish"],
    },
    size: sizeArg(["small", "medium"]),
    variant: {
      control: "radio",
      options: ["filled", "outlined"],
    },
    orientation: {
      control: "radio",
      options: ["horizontal", "vertical"],
    },
    percent: { control: { type: "range", min: 0, max: 100 } },
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Steps>

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      {/* 1. Basic */}
      <Section title="Basic">
        <Steps current={1} items={basicItems} />
      </Section>

      {/* 2. Size */}
      <Section title="Sizes">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>medium (default)</Label>
            <Steps current={1} items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>small</Label>
            <Steps current={1} size="small" items={basicItems} />
          </Flex>
        </Flex>
      </Section>

      {/* 3. With Icon */}
      <Section title="With Icon">
        <Steps
          current={1}
          items={[
            { title: "Login", icon: <UserOutlined /> },
            { title: "Verification", icon: <SolutionOutlined /> },
            { title: "Done", icon: <SmileOutlined /> },
          ]}
        />
      </Section>

      {/* 4. Status */}
      <Section title="Status">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>process (default)</Label>
            <Steps current={1} status="process" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>error</Label>
            <Steps current={1} status="error" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>wait</Label>
            <Steps current={1} status="wait" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>finish</Label>
            <Steps current={1} status="finish" items={basicItems} />
          </Flex>
        </Flex>
      </Section>

      {/* 5. Vertical */}
      <Section title="Vertical">
        <Steps
          current={1}
          orientation="vertical"
          items={basicItems}
          style={{ width: 300 }}
        />
      </Section>

      {/* 6. Types */}
      <Section title="Types">
        <Flex vertical gap={24}>
          {types.map((t) => (
            <Flex key={t} vertical gap={4}>
              <Label>{t}</Label>
              <Steps
                current={1}
                type={t}
                items={
                  t === "navigation"
                    ? [
                        { title: "Login", subTitle: "00:00:01" },
                        { title: "Verification", subTitle: "00:00:02" },
                        { title: "Payment", subTitle: "00:00:03" },
                      ]
                    : basicItems
                }
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      {/* 7. Variant */}
      <Section title="Variant (filled vs outlined)">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>filled (default)</Label>
            <Steps current={1} variant="filled" items={basicItems} />
          </Flex>
          <Flex vertical gap={4}>
            <Label>outlined</Label>
            <Steps current={1} variant="outlined" items={basicItems} />
          </Flex>
        </Flex>
      </Section>

      <Section title="Progress Dot">
        <Flex vertical gap={16}>
          <Flex vertical gap={4}>
            <Label>percent = 40</Label>
            <Steps
              current={1}
              percent={40}
              items={[
                { title: "Finished", description: "Completed" },
                { title: "In Progress", description: "40% done" },
                { title: "Waiting", description: "Pending" },
              ]}
            />
          </Flex>
          <Flex vertical gap={4}>
            <Label>percent = 75</Label>
            <Steps
              current={1}
              percent={75}
              items={[
                { title: "Finished", description: "Completed" },
                { title: "In Progress", description: "75% done" },
                { title: "Waiting", description: "Pending" },
              ]}
            />
          </Flex>
        </Flex>
      </Section>

      <Section title="Vertical + Dot">
        <Steps
          current={1}
          type="dot"
          orientation="vertical"
          items={basicItems}
          style={{ width: 300 }}
        />
      </Section>

      <Section title="Vertical + Status">
        <Flex gap={24} wrap>
          {(["process", "error", "wait", "finish"] as const).map((status) => (
            <Flex key={status} vertical gap={4} style={{ width: 280 }}>
              <Label>{status}</Label>
              <Steps
                current={1}
                status={status}
                orientation="vertical"
                items={basicItems}
              />
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Steps
            current={1}
            items={[
              { title: "Step 1" },
              { title: "Step 2" },
              { title: "Step 3" },
            ]}
          />
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    current: 1,
    items: [
      { title: "Step 1", content: "First" },
      { title: "Step 2", content: "Second" },
      { title: "Step 3", content: "Third" },
    ],
    onChange: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const step3 = canvas.getByText("Step 3")
    await step3.click()
    await expect(args.onChange).toHaveBeenCalled()
  },
}

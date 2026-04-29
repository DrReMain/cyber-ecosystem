import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Progress } from "antd"
import { useState } from "react"
import { expect, userEvent, within } from "storybook/test"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Progress> = {
  title: "Antd/Feedback/Progress",
  component: Progress,
  parameters: { layout: "padded" },
  args: {
    type: "line",
    percent: 60,
    showInfo: true,
  },
  argTypes: {
    type: {
      control: "select",
      options: ["line", "circle", "dashboard"],
    },
    percent: { control: { type: "range", min: 0, max: 100, step: 1 } },
    status: {
      control: "select",
      options: ["normal", "success", "exception", "active"],
    },
    strokeColor: { control: "color" },
    showInfo: { control: "boolean" },
    strokeLinecap: {
      control: "radio",
      options: ["round", "butt", "square"],
    },
    size: {
      control: "select",
      options: ["small", "medium"],
    },
  },
}

export default meta
type Story = StoryObj<typeof Progress>

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const lineStatuses = [
      { percent: 100, status: undefined, label: "100% (auto success)" },
      { percent: 80, status: "active" as const, label: "80% active" },
      { percent: 50, status: "exception" as const, label: "50% exception" },
      { percent: 30, status: undefined, label: "30% normal" },
      { percent: 70, status: "success" as const, label: "70% success" },
    ]

    return (
      <Flex vertical gap={24}>
        <Section title="Line - Statuses">
          <Flex vertical gap={8}>
            {lineStatuses.map((s) => (
              <div key={s.label}>
                <Label>{s.label}</Label>
                <Progress percent={s.percent} status={s.status} />
              </div>
            ))}
          </Flex>
        </Section>

        <Section title="Line - Success Segment">
          <Flex vertical gap={8}>
            <Progress percent={60} success={{ percent: 30 }} />
            <Progress
              percent={80}
              success={{ percent: 50, strokeColor: "#52c41a" }}
            />
          </Flex>
        </Section>

        <Section title="Circle">
          <Flex gap={24} wrap align="center">
            <Progress type="circle" percent={75} />
            <Progress type="circle" percent={100} />
            <Progress type="circle" percent={50} status="exception" />
            <Progress type="circle" percent={30} status="active" />
          </Flex>
        </Section>

        <Section title="Dashboard">
          <Flex gap={24} wrap align="center">
            <Progress type="dashboard" percent={75} />
            <Progress type="dashboard" percent={100} />
            <Progress type="dashboard" percent={50} status="exception" />
          </Flex>
        </Section>

        <Section title="Gradient Stroke">
          <Flex gap={24} wrap align="center">
            <Progress
              type="circle"
              percent={90}
              strokeColor={{ "0%": "#108ee9", "100%": "#87d068" }}
            />
            <Progress
              type="circle"
              percent={100}
              strokeColor={{
                "0%": "#ffc53d",
                "50%": "#ff4d4f",
                "100%": "#cf1322",
              }}
            />
            <Progress
              type="dashboard"
              percent={80}
              strokeColor={{ "0%": "#36cfc9", "100%": "#9254de" }}
            />
            <Progress
              percent={80}
              strokeColor={{ from: "#108ee9", to: "#87d068" }}
            />
          </Flex>
        </Section>

        <Section title="Sizes">
          <Flex vertical gap={16}>
            <Label>small</Label>
            <Progress percent={60} size="small" />
            <Label>medium (default)</Label>
            <Progress percent={60} size="medium" />
            <Label>numeric size (circle/dashboard)</Label>
            <Flex gap={24} wrap align="center">
              <Progress type="circle" percent={60} size={80} />
              <Progress type="circle" percent={60} size={120} />
              <Progress type="dashboard" percent={60} size={80} />
              <Progress type="dashboard" percent={60} size={120} />
            </Flex>
            <Label>object size (width x height for line)</Label>
            <Progress percent={60} size={{ width: 400, height: 20 }} />
          </Flex>
        </Section>
      </Flex>
    )
  },
}

// ── Interactive ───────────────────────────────────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [percent, setPercent] = useState(30)
    return (
      <Flex vertical gap={16}>
        <Progress percent={percent} />
        <Button
          data-testid="toggle-progress"
          onClick={() => setPercent((p) => (p === 30 ? 100 : 30))}
        >
          Toggle Progress
        </Button>
      </Flex>
    )
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByTestId("toggle-progress")
    await userEvent.click(button)
    const bar = canvas.getByRole("progressbar")
    await expect(bar).toHaveAttribute("aria-valuenow", "100")
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Badge, Calendar, Flex, theme } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta<typeof Calendar> = {
  title: "Antd/Data Display/Calendar",
  component: Calendar,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    fullscreen: { control: "boolean" },
    showWeek: { control: "boolean" },
    onSelect: { action: "selected" },
  },
}

export default meta
type Story = StoryObj<typeof Calendar>

// ── Args-driven stories (render-driven for complex composite) ──

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const { token } = theme.useToken()

    return (
      <Flex vertical gap={24}>
        <Section title="Basic" description="Fullscreen month view (default)">
          <Calendar />
        </Section>

        <Section
          title="Card Mode"
          description="fullscreen={false} for compact panel"
        >
          <Calendar fullscreen={false} />
        </Section>

        <Section
          title="Custom Cell Render"
          description="cellRender to add status badges on specific dates"
        >
          <Calendar
            fullscreen={false}
            cellRender={(current, info) => {
              if (info.type !== "date") return info.originNode
              const day = current.date()
              const badgeMap: Record<number, "success" | "warning" | "error"> =
                {
                  8: "success",
                  15: "success",
                  22: "success",
                  10: "warning",
                  20: "warning",
                  25: "error",
                }
              const status = badgeMap[day]
              if (status) {
                return (
                  <div style={{ position: "relative" }}>
                    {day}
                    <Badge
                      status={status}
                      style={{ position: "absolute", top: -2, right: -8 }}
                    />
                  </div>
                )
              }
              return info.originNode
            }}
          />
        </Section>

        <Section
          title="Custom Cell Render (Fullscreen)"
          description="Fullscreen calendar with theme-aware highlighted cells"
        >
          <Calendar
            cellRender={(current, info) => {
              if (info.type !== "date") return info.originNode
              const day = current.date()
              const highlighted = [5, 12, 19, 26]
              if (highlighted.includes(day)) {
                return (
                  <div
                    style={{
                      background: token.colorPrimaryBg,
                      borderRadius: token.borderRadiusSM,
                      padding: "2px 4px",
                    }}
                  >
                    {day}
                  </div>
                )
              }
              return info.originNode
            }}
          />
        </Section>

        <Section
          title="Disabled Dates"
          description="disabledDate prevents selecting weekends"
        >
          <Calendar
            fullscreen={false}
            disabledDate={(current) => {
              const day = current.day()
              return day === 0 || day === 6
            }}
          />
        </Section>
      </Flex>
    )
  },
}

export const Playground: Story = {
  render: (args) => <Calendar {...args} />,
}

export const Interactive: Story = {
  args: {
    fullscreen: false,
    onSelect: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const cells = canvas.getAllByRole("gridcell")
    if (cells.length > 0) {
      await userEvent.click(cells[0])
    }
    await expect(args.onSelect).toHaveBeenCalled()
  },
}

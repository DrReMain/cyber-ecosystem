import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Skeleton, Space } from "antd"
import { useState } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Skeleton> = {
  title: "Antd/Feedback/Skeleton",
  component: Skeleton,
  parameters: { layout: "padded" },
  args: {
    active: true,
    loading: true,
    avatar: false,
    paragraph: { rows: 3 },
    title: true,
    round: false,
  },
  argTypes: {
    active: { control: "boolean" },
    loading: { control: "boolean" },
    avatar: { control: "boolean" },
    round: { control: "boolean" },
    title: { control: "boolean" },
    paragraph: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Skeleton>

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const avatarShapes = ["circle", "square"] as const
    const avatarSizes = ["small", "medium", "large"] as const
    const buttonShapes = ["default", "round", "circle", "square"] as const
    const buttonSizes = ["small", "medium", "large"] as const

    return (
      <Flex vertical gap={24}>
        <Section title="Basic">
          <Skeleton />
        </Section>

        <Section title="Active (Animated)">
          <Skeleton active />
        </Section>

        <Section title="Avatar Shapes & Sizes">
          <Flex vertical gap={12}>
            {avatarSizes.map((size) => (
              <div key={size}>
                <Label>{size}</Label>
                <Space>
                  {avatarShapes.map((shape) => (
                    <Skeleton.Avatar
                      key={shape}
                      active
                      shape={shape}
                      size={size}
                    />
                  ))}
                </Space>
              </div>
            ))}
          </Flex>
        </Section>

        <Section title="Button Shapes & Sizes">
          <Flex vertical gap={12}>
            {buttonSizes.map((size) => (
              <div key={size}>
                <Label>{size}</Label>
                <Space>
                  {buttonShapes.map((shape) => (
                    <Skeleton.Button
                      key={shape}
                      active
                      shape={shape}
                      size={size}
                    />
                  ))}
                </Space>
              </div>
            ))}
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
    const [loading, setLoading] = useState(true)
    const onClick = fn()
    return (
      <Flex vertical gap={16}>
        <Skeleton active loading={loading} paragraph={{ rows: 3 }}>
          <div data-testid="skeleton-content">Content loaded!</div>
        </Skeleton>
        <Button
          data-testid="toggle-loading"
          onClick={() => {
            onClick()
            setLoading((l) => !l)
          }}
        >
          Toggle Loading
        </Button>
      </Flex>
    )
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByTestId("toggle-loading")
    await userEvent.click(button)
    await expect(canvas.getByTestId("skeleton-content")).toBeInTheDocument()
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {}

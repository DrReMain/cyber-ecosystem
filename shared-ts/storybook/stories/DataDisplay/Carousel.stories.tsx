import { LeftOutlined, RightOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Carousel, Flex, Space, theme } from "antd"
import type { CarouselRef } from "antd/es/carousel"
import { useRef } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Carousel> = {
  title: "Antd/Data Display/Carousel",
  component: Carousel,
  parameters: { layout: "padded" },
  args: {},
  argTypes: {
    effect: {
      control: "radio",
      options: ["scrollx", "fade"],
    },
    autoplay: { control: "boolean" },
    dots: { control: "boolean" },
    dotPlacement: {
      control: "radio",
      options: ["top", "bottom"],
    },
    draggable: { control: "boolean" },
    adaptiveHeight: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Carousel>

const slideStyle = (
  color: string,
  textColor?: string,
): React.CSSProperties => ({
  height: 160,
  color: textColor ?? "var(--ant-color-bg-layout)",
  lineHeight: "160px",
  textAlign: "center",
  background: color,
  fontSize: 24,
})

const useTokenSlides = () => {
  const { token } = theme.useToken()
  return [
    { color: token.colorPrimary, label: "1" },
    { color: token.colorSuccess, label: "2" },
    { color: token.colorWarning, label: "3" },
    { color: token.colorError, label: "4" },
  ]
}

const effects: Array<"scrollx" | "fade"> = ["scrollx", "fade"]
const dotPlacements: Array<
  React.ComponentProps<typeof Carousel>["dotPlacement"]
> = ["top", "bottom"]

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const arrowRef = useRef<CarouselRef>(null)
    const { token } = theme.useToken()
    const slides = useTokenSlides()
    return (
      <Flex vertical gap={24}>
        <Section
          title="Effect Variants"
          description="scrollx (default) vs fade"
        >
          <Flex vertical gap={16}>
            {effects.map((effect) => (
              <Space key={effect} orientation="vertical" size={4}>
                <Label>{effect}</Label>
                <Carousel effect={effect} autoplay>
                  {slides.map((s) => (
                    <div key={s.label}>
                      <h3 style={slideStyle(s.color)}>
                        {effect} {s.label}
                      </h3>
                    </div>
                  ))}
                </Carousel>
              </Space>
            ))}
          </Flex>
        </Section>

        <Section title="Dot Placement" description="Position of indicator dots">
          <Flex vertical gap={16}>
            {dotPlacements.map((placement) => (
              <Space key={placement} orientation="vertical" size={4}>
                <Label>{placement}</Label>
                <Carousel autoplay dotPlacement={placement}>
                  {slides.map((s) => (
                    <div key={s.label}>
                      <h3 style={slideStyle(s.color)}>
                        {placement} {s.label}
                      </h3>
                    </div>
                  ))}
                </Carousel>
              </Space>
            ))}
          </Flex>
        </Section>

        <Section
          title="Autoplay"
          description="Auto-scrolling with default 3s interval"
        >
          <Carousel autoplay>
            {slides.map((s) => (
              <div key={s.label}>
                <h3 style={slideStyle(s.color)}>Autoplay {s.label}</h3>
              </div>
            ))}
          </Carousel>
        </Section>

        <Section
          title="Autoplay with Progress"
          description="autoplay={{ dotDuration: true }} shows progress bar"
        >
          <Carousel autoplay={{ dotDuration: true }}>
            {slides.map((s) => (
              <div key={s.label}>
                <h3 style={slideStyle(s.color)}>Progress {s.label}</h3>
              </div>
            ))}
          </Carousel>
        </Section>

        <Section
          title="With Arrows"
          description="Built-in arrow navigation via ref control"
        >
          <Flex vertical gap={12}>
            <Space>
              <Button
                icon={<LeftOutlined />}
                onClick={() => arrowRef.current?.prev()}
              />
              <Button
                icon={<RightOutlined />}
                onClick={() => arrowRef.current?.next()}
              />
            </Space>
            <Carousel ref={arrowRef}>
              {slides.map((s) => (
                <div key={s.label}>
                  <h3 style={slideStyle(s.color)}>Arrow {s.label}</h3>
                </div>
              ))}
            </Carousel>
          </Flex>
        </Section>

        <Section
          title="Draggable"
          description="Enable mouse/touch drag to change slides"
        >
          <Carousel draggable>
            {slides.map((s) => (
              <div key={s.label}>
                <h3 style={slideStyle(s.color)}>Drag {s.label}</h3>
              </div>
            ))}
          </Carousel>
        </Section>

        <Section
          title="Single Slide"
          description="Verify dot behavior with minimal content"
        >
          <Carousel autoplay>
            <div>
              <h3 style={slideStyle(token.colorPrimary)}>Only slide</h3>
            </div>
          </Carousel>
        </Section>

        <Section title="No Dots" description="dots={false} hides indicator">
          <Carousel dots={false} autoplay>
            {slides.map((s) => (
              <div key={s.label}>
                <h3 style={slideStyle(s.color)}>No dots {s.label}</h3>
              </div>
            ))}
          </Carousel>
        </Section>

        <Section
          title="Adaptive Height"
          description="Height adjusts to current slide content"
        >
          <Carousel adaptiveHeight>
            {slides.map((s, i) => (
              <div key={s.label}>
                <h3
                  style={{
                    ...slideStyle(s.color),
                    height: 80 + i * 40,
                    lineHeight: `${80 + i * 40}px`,
                  }}
                >
                  Height {80 + i * 40}px
                </h3>
              </div>
            ))}
          </Carousel>
        </Section>
      </Flex>
    )
  },
}

export const Playground: Story = {
  render: (args) => {
    const slides = useTokenSlides()
    return (
      <Carousel {...args}>
        {slides.map((s) => (
          <div key={s.label}>
            <h3 style={slideStyle(s.color)}>Slide {s.label}</h3>
          </div>
        ))}
      </Carousel>
    )
  },
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    afterChange: fn(),
  },
  render: (args) => <ArrowCarousel afterChange={args.afterChange} />,
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const nextBtn = canvas.getByRole("button", { name: /next/i })
    await userEvent.click(nextBtn)
    await expect(args.afterChange).toHaveBeenCalled()
  },
}

function ArrowCarousel({
  afterChange,
}: {
  afterChange?: (current: number) => void
}) {
  const ref = useRef<CarouselRef>(null)
  const slides = useTokenSlides()
  return (
    <Flex vertical gap={12}>
      <Space>
        <Button
          icon={<LeftOutlined />}
          onClick={() => ref.current?.prev()}
          aria-label="Previous slide"
        >
          Prev
        </Button>
        <Button
          icon={<RightOutlined />}
          onClick={() => ref.current?.next()}
          aria-label="Next slide"
        >
          Next
        </Button>
      </Space>
      <Carousel ref={ref} afterChange={afterChange}>
        {slides.map((s) => (
          <div key={s.label}>
            <h3 style={slideStyle(s.color)}>Slide {s.label}</h3>
          </div>
        ))}
      </Carousel>
    </Flex>
  )
}

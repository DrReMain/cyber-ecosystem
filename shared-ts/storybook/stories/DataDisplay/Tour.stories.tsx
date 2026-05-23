import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Space, Tour } from "antd"
import { useRef, useState } from "react"
import { expect, fn, within } from "storybook/test"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof Tour> = {
  title: "Antd/Data Display/Tour",
  component: Tour,
  parameters: { layout: "padded" },
  args: { open: false },
  argTypes: {
    type: {
      control: "select",
      options: ["default", "primary"],
    },
    arrow: { control: "boolean" },
    disabledInteraction: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Tour>

// ── Standalone function components (used by named stories & Gallery) ──

function BasicTour() {
  const ref1 = useRef<HTMLButtonElement>(null)
  const ref2 = useRef<HTMLButtonElement>(null)
  const ref3 = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  const steps = [
    {
      title: "Upload File",
      description: "Put your files here.",
      target: () => ref1.current as HTMLElement,
    },
    {
      title: "Save",
      description: "Save your changes.",
      target: () => ref2.current as HTMLElement,
    },
    {
      title: "Other Actions",
      description: "Click to see other actions.",
      target: () => ref3.current as HTMLElement,
    },
  ]

  return (
    <>
      <Space>
        <Button onClick={() => setOpen(true)}>Start Tour</Button>
        <Button type="primary" ref={ref1}>
          Upload
        </Button>
        <Button ref={ref2}>Save</Button>
        <Button ref={ref3}>More</Button>
      </Space>
      <Tour open={open} onClose={() => setOpen(false)} steps={steps} />
    </>
  )
}

function PrimaryTypeTour() {
  const ref = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  return (
    <>
      <Space>
        <Button onClick={() => setOpen(true)}>Primary Type</Button>
        <Button ref={ref}>Target</Button>
      </Space>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        type="primary"
        steps={[
          {
            title: "Primary Tour",
            description: "This tour uses primary type styling.",
            target: () => ref.current as HTMLElement,
          },
          {
            title: "Step 2",
            description: "Still primary type.",
            target: () => ref.current as HTMLElement,
            placement: "top",
          },
        ]}
      />
    </>
  )
}

function NoMaskTour() {
  const ref = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  return (
    <>
      <Space>
        <Button onClick={() => setOpen(true)}>No Mask</Button>
        <Button ref={ref}>Target</Button>
      </Space>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        mask={false}
        steps={[
          {
            title: "No Mask",
            description: "This tour has no mask overlay.",
            target: () => ref.current as HTMLElement,
          },
          {
            title: "Step 2",
            description: "Still no mask.",
            target: () => ref.current as HTMLElement,
            placement: "top",
          },
        ]}
      />
    </>
  )
}

function CustomMaskTour() {
  const ref = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  return (
    <>
      <Space>
        <Button onClick={() => setOpen(true)}>Custom Mask</Button>
        <Button ref={ref}>Target</Button>
      </Space>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        mask={{ style: { backgroundColor: "rgba(0, 0, 0, 0.6)" } }}
        steps={[
          {
            title: "Custom Mask",
            description: "The mask has a darker overlay color.",
            target: () => ref.current as HTMLElement,
          },
        ]}
      />
    </>
  )
}

function NoArrowTour() {
  const ref = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  return (
    <>
      <Space>
        <Button onClick={() => setOpen(true)}>No Arrow</Button>
        <Button ref={ref}>Target</Button>
      </Space>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        arrow={false}
        steps={[
          {
            title: "No Arrow",
            description: "This tour has no arrow pointing to the target.",
            target: () => ref.current as HTMLElement,
          },
        ]}
      />
    </>
  )
}

function CenterTour() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button onClick={() => setOpen(true)}>Center Tour (no target)</Button>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        steps={[
          {
            title: "Welcome!",
            description:
              "This tour appears in the center of the screen without a target element.",
            placement: "center",
          },
          {
            title: "Get Started",
            description: "Click next to continue the tour.",
            placement: "center",
          },
        ]}
      />
    </>
  )
}

function StaticOpenTour() {
  const ref = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(true)

  return (
    <>
      <Space>
        <Button type="primary" ref={ref}>
          Target Element
        </Button>
        <Button onClick={() => setOpen(true)} disabled={open}>
          Reopen Tour
        </Button>
      </Space>
      <Tour
        open={open}
        onClose={() => setOpen(false)}
        steps={[
          {
            title: "Welcome to the Tour",
            description:
              "This tour is statically open so skin devs can see the card, mask, arrow, and step indicators without interaction.",
            target: () => ref.current as HTMLElement,
          },
          {
            title: "Second Step",
            description:
              "Navigate using the buttons below to see different step states.",
            target: () => ref.current as HTMLElement,
            placement: "right",
          },
        ]}
      />
    </>
  )
}

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic Tour">
        <BasicTour />
      </Section>

      <Section title="Type: primary">
        <PrimaryTypeTour />
      </Section>

      <Section title="Mask Options">
        <Flex vertical gap={12}>
          <div>
            <Label>No mask</Label>
            <NoMaskTour />
          </div>
          <div>
            <Label>Custom mask style</Label>
            <CustomMaskTour />
          </div>
        </Flex>
      </Section>

      <Section title="Arrow: hidden">
        <NoArrowTour />
      </Section>

      <Section title="Center Placement (no target)">
        <CenterTour />
      </Section>
    </Flex>
  ),
}

export const StaticOpen: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <>
      <PageContent>
        <div style={{ marginTop: 24 }}>
          <StaticOpenTour />
        </div>
      </PageContent>
    </>
  ),
}

export const Playground: Story = {
  render: () => <BasicTour />,
}

// ── Interactive story with play function ──────────────────────

export const Interactive: Story = {
  args: {
    onClose: fn(),
  },
  render: (args) => {
    const ref = useRef<HTMLButtonElement>(null)
    const [open, setOpen] = useState(false)

    return (
      <>
        <Space>
          <Button onClick={() => setOpen(true)}>Start Tour</Button>
          <Button ref={ref}>Target</Button>
        </Space>
        <Tour
          open={open}
          onClose={() => {
            setOpen(false)
            args.onClose?.(
              {} as Parameters<NonNullable<typeof args.onClose>>[0],
            )
          }}
          steps={[
            {
              title: "Interactive Tour",
              description:
                "This story verifies the tour opens on click and renders correctly.",
              target: () => ref.current as HTMLElement,
            },
          ]}
        />
      </>
    )
  },
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByText("Start Tour")
    await button.click()
    const tour = await within(document.body).findByRole("dialog")
    await expect(tour).toBeInTheDocument()
    const closeBtn = within(tour).getByRole("button", { name: /close/i })
    await closeBtn.click()
    await expect(args.onClose).toHaveBeenCalled()
  },
}

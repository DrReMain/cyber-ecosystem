import type { Meta, StoryObj } from "@storybook/react-vite"
import { Flex, Image, Space } from "antd"
import { useState } from "react"
import { Label, PageContent, Section } from "../helpers"

const meta: Meta<typeof Image> = {
  title: "Antd/Data Display/Image",
  component: Image,
  parameters: { layout: "padded" },
  args: {
    width: 200,
    src: "https://picsum.photos/seed/basic/200/150",
    alt: "Image",
  },
  argTypes: {
    width: { control: { type: "number", min: 50, max: 600 } },
    height: { control: { type: "number", min: 50, max: 600 } },
    preview: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Image>

const img = (seed: string, w = 200, h = 150) =>
  `https://picsum.photos/seed/${seed}/${w}/${h}`

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Image width={200} src={img("basic")} alt="Basic image" />
      </Section>

      <Section title="Preview Disabled">
        <Image
          width={200}
          src={img("no-preview")}
          preview={false}
          alt="No preview"
        />
      </Section>

      <Section
        title="Preview Group"
        description="Image.PreviewGroup for gallery/lightbox browsing"
      >
        <Image.PreviewGroup>
          <Flex gap={8} wrap>
            {[
              "https://picsum.photos/seed/antd1/200/150",
              "https://picsum.photos/seed/antd2/200/150",
              "https://picsum.photos/seed/antd3/200/150",
              "https://picsum.photos/seed/antd4/200/150",
              "https://picsum.photos/seed/antd5/200/150",
              "https://picsum.photos/seed/antd6/200/150",
            ].map((src) => (
              <Image
                key={src}
                width={120}
                height={90}
                src={src}
                style={{ borderRadius: 4, objectFit: "cover" }}
              />
            ))}
          </Flex>
        </Image.PreviewGroup>
      </Section>

      <Section title="Image Grid" description="Responsive image grid layout">
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fill, minmax(150px, 1fr))",
            gap: 8,
          }}
        >
          {[
            "https://picsum.photos/seed/grid1/300/200",
            "https://picsum.photos/seed/grid2/300/200",
            "https://picsum.photos/seed/grid3/300/200",
            "https://picsum.photos/seed/grid4/300/200",
            "https://picsum.photos/seed/grid5/300/200",
            "https://picsum.photos/seed/grid6/300/200",
            "https://picsum.photos/seed/grid7/300/200",
            "https://picsum.photos/seed/grid8/300/200",
          ].map((src) => (
            <Image
              key={src}
              src={src}
              style={{ borderRadius: 4, objectFit: "cover", width: "100%" }}
            />
          ))}
        </div>
      </Section>

      <Section
        title="Fallback"
        description="Fallback image when the source fails to load"
      >
        <Flex gap={16} align="center">
          <div style={{ textAlign: "center" }}>
            <Label>Broken URL with fallback</Label>
            <Image
              width={120}
              height={90}
              src="https://invalid-url.example.com/broken.jpg"
              fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTIwIiBoZWlnaHQ9IjkwIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxyZWN0IHdpZHRoPSIxMjAiIGhlaWdodD0iOTAiIGZpbGw9IiNmNWY1ZjUiLz48dGV4dCB4PSI2MCIgeT0iNDUiIGZvbnQtc2l6ZT0iMTIiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiNiZmJmYmYiPkltYWdlIEVycm9yPC90ZXh0Pjwvc3ZnPg=="
              style={{ borderRadius: 4 }}
            />
          </div>
          <div style={{ textAlign: "center" }}>
            <Label>Placeholder image</Label>
            <Image
              width={120}
              height={90}
              src="https://picsum.photos/seed/placeholder/120/90"
              placeholder={
                <div
                  style={{
                    background: "var(--ant-color-fill-quaternary)",
                    width: 120,
                    height: 90,
                    borderRadius: 4,
                  }}
                />
              }
              style={{ borderRadius: 4 }}
            />
          </div>
        </Flex>
      </Section>
    </Flex>
  ),
}

function PreviewVisibleDemo() {
  const [open, setOpen] = useState(true)
  return (
    <Space orientation="vertical">
      <Label>Preview opens on mount; close it to dismiss</Label>
      <Image
        width={200}
        src={img("preview-visible", 800, 600)}
        preview={{ open, onOpenChange: (vis) => setOpen(vis) }}
        alt="Preview visible demo"
      />
    </Space>
  )
}

export const StaticOpen: Story = {
  render: () => (
    <>
      <PageContent />
      <PreviewVisibleDemo />
    </>
  ),
}

export const Playground: Story = {}

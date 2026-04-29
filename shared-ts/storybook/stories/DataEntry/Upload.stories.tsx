import { PlusOutlined, UploadOutlined } from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Upload } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg } from "../argTypes"
import { Label, Section } from "../helpers"

const meta: Meta<typeof Upload> = {
  title: "Antd/Data Entry/Upload",
  component: Upload,
  parameters: { layout: "padded" },
  args: {
    action: "https://run.mocky.io/v3/435e08b3-884a-4d68-8044-e5e73350a4a8",
  },
  argTypes: {
    disabled: disabledArg,
    multiple: { control: "boolean" },
    listType: {
      control: "radio",
      options: ["text", "picture", "picture-card", "picture-circle"],
    },
    maxCount: { control: { type: "number", min: 1, max: 20 } },
    accept: { control: "text" },
    onChange: { action: "changed" },
  },
}

export default meta
type Story = StoryObj<typeof Upload>

const mockAction =
  "https://run.mocky.io/v3/435e08b3-884a-4d68-8044-e5e73350a4a8"

const defaultFileList = [
  {
    uid: "1",
    name: "example.png",
    status: "done" as const,
    url: "https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png",
  },
  {
    uid: "2",
    name: "document.pdf",
    status: "done" as const,
  },
]

const uploadingFileList = [
  {
    uid: "3",
    name: "uploading.jpg",
    status: "uploading" as const,
    percent: 55,
  },
]

const errorFileList = [
  {
    uid: "4",
    name: "failed.txt",
    status: "error" as const,
    response: "Server Error 500",
  },
]

const listTypes = ["text", "picture", "picture-card", "picture-circle"] as const

// ── Args-driven stories ───────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="listType variants">
        <Flex vertical gap={16}>
          {listTypes.map((listType) => (
            <Flex vertical gap={4} key={listType}>
              <Label>{listType}</Label>
              <Upload
                action={mockAction}
                listType={listType}
                defaultFileList={defaultFileList}
              >
                {listType === "picture-card" ||
                listType === "picture-circle" ? (
                  <div>
                    <PlusOutlined /> Upload
                  </div>
                ) : (
                  <Button icon={<UploadOutlined />}>Upload ({listType})</Button>
                )}
              </Upload>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Disabled">
        <Flex vertical gap={8}>
          <Upload disabled>
            <Button icon={<UploadOutlined />} disabled>
              Click to Upload
            </Button>
          </Upload>
          <Upload disabled defaultFileList={defaultFileList}>
            <Button icon={<UploadOutlined />} disabled>
              Upload
            </Button>
          </Upload>
        </Flex>
      </Section>

      <Section title="Dragger">
        <Flex gap={24} wrap>
          <Upload.Dragger action={mockAction} style={{ width: 300 }}>
            <p className="ant-upload-drag-icon">
              <UploadOutlined />
            </p>
            <p className="ant-upload-text">
              Click or drag file to this area to upload
            </p>
            <p className="ant-upload-hint">
              Support for a single or bulk upload.
            </p>
          </Upload.Dragger>
          <Upload.Dragger action={mockAction} disabled style={{ width: 300 }}>
            <p className="ant-upload-drag-icon">
              <UploadOutlined />
            </p>
            <p className="ant-upload-text">Disabled dragger</p>
          </Upload.Dragger>
        </Flex>
      </Section>

      <Section title="File Statuses">
        <Flex vertical gap={16}>
          <Label>done / uploading / error states</Label>
          <Upload
            action={mockAction}
            defaultFileList={[
              ...defaultFileList,
              ...uploadingFileList,
              ...errorFileList,
            ]}
          >
            <Button icon={<UploadOutlined />}>Upload</Button>
          </Upload>
        </Flex>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Upload {...args}>
      <Button icon={<UploadOutlined />}>Click to Upload</Button>
    </Upload>
  ),
}

export const Interactive: Story = {
  args: {
    onChange: fn(),
  },
  render: (args) => (
    <Upload {...args}>
      <Button icon={<UploadOutlined />}>Click to Upload</Button>
    </Upload>
  ),
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole("button", { name: /click to upload/i })
    await button.click()
    await expect(args.onChange).not.toHaveBeenCalled()
  },
}

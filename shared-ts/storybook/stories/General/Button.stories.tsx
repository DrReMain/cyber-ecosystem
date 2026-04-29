import {
  CloudDownloadOutlined,
  LoadingOutlined,
  SearchOutlined,
} from "@ant-design/icons"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Space } from "antd"
import { expect, fn, within } from "storybook/test"
import { disabledArg, loadingArg, sizeArg } from "../argTypes"
import { Label, Section } from "../helpers"
import { PseudoStates } from "../pseudo-states"

const meta: Meta<typeof Button> = {
  title: "Antd/General/Button",
  component: Button,
  parameters: { layout: "padded" },
  args: { children: "Button" },
  argTypes: {
    variant: {
      control: "select",
      options: ["outlined", "dashed", "solid", "filled", "text", "link"],
    },
    color: {
      control: "select",
      options: [
        "default",
        "primary",
        "danger",
        "blue",
        "purple",
        "cyan",
        "green",
        "magenta",
        "pink",
        "red",
        "orange",
        "yellow",
        "volcano",
        "geekblue",
        "lime",
        "gold",
      ],
    },
    size: sizeArg(["large", "medium", "small"]),
    shape: {
      control: "radio",
      options: ["default", "circle", "round"],
    },
    disabled: disabledArg,
    loading: loadingArg,
    ghost: { control: "boolean" },
    block: { control: "boolean" },
    danger: { control: "boolean" },
    onClick: { action: "clicked" },
  },
}

export default meta
type Story = StoryObj<typeof Button>

const variants = [
  "outlined",
  "dashed",
  "solid",
  "filled",
  "text",
  "link",
] as const
const sizes = ["large", "medium", "small"] as const
const colors = ["default", "primary", "danger"] as const
const presetColors = [
  "blue",
  "green",
  "volcano",
  "orange",
  "gold",
  "lime",
  "cyan",
  "geekblue",
  "purple",
  "magenta",
  "red",
  "pink",
  "yellow",
] as const

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Variants x Colors">
        <Flex vertical gap={12}>
          {colors.map((color) => (
            <Flex key={color} vertical gap={4}>
              <Label>{color}</Label>
              <Space wrap>
                {variants.map((v) => (
                  <Button key={v} variant={v} color={color}>
                    {v}
                  </Button>
                ))}
              </Space>
            </Flex>
          ))}
        </Flex>
      </Section>

      <Section title="Preset Colors (solid variant)">
        <Space wrap>
          {presetColors.map((c) => (
            <Button key={c} variant="solid" color={c}>
              {c}
            </Button>
          ))}
        </Space>
      </Section>
      <Section title="Preset Colors (outlined variant)">
        <Space wrap>
          {presetColors.map((c) => (
            <Button key={c} variant="outlined" color={c}>
              {c}
            </Button>
          ))}
        </Space>
      </Section>

      <Section title="Sizes">
        <Flex vertical gap={8}>
          {sizes.map((size) => (
            <Space key={size} wrap>
              <Button color="primary" variant="solid" size={size}>
                Primary
              </Button>
              <Button variant="outlined" size={size}>
                Default
              </Button>
              <Button variant="dashed" size={size}>
                Dashed
              </Button>
              <Button variant="filled" size={size}>
                Filled
              </Button>
              <Button variant="text" size={size}>
                Text
              </Button>
              <Button variant="link" size={size}>
                Link
              </Button>
            </Space>
          ))}
        </Flex>
      </Section>

      <Section title="Shapes">
        <Flex vertical gap={8}>
          <Label>default</Label>
          <Space wrap>
            <Button color="primary" variant="solid">
              Primary
            </Button>
            <Button variant="outlined">Default</Button>
            <Button variant="dashed">Dashed</Button>
          </Space>
          <Label>circle</Label>
          <Space wrap>
            <Button color="primary" variant="solid" shape="circle">
              A
            </Button>
            <Button shape="circle" icon={<SearchOutlined />} />
            <Button variant="dashed" shape="circle" icon={<SearchOutlined />} />
            {sizes.map((s) => (
              <Button
                key={s}
                color="primary"
                variant="solid"
                shape="circle"
                size={s}
                icon={<SearchOutlined />}
              />
            ))}
          </Space>
          <Label>round</Label>
          <Space wrap>
            <Button color="primary" variant="solid" shape="round">
              Primary
            </Button>
            <Button shape="round">Default</Button>
            <Button variant="dashed" shape="round">
              Dashed
            </Button>
            <Button variant="filled" shape="round">
              Filled
            </Button>
            <Button
              color="primary"
              variant="solid"
              shape="round"
              icon={<SearchOutlined />}
            >
              Search
            </Button>
          </Space>
        </Flex>
      </Section>
      <Section title="Icons">
        <Flex vertical gap={8}>
          <Label>icon + text</Label>
          <Space wrap>
            <Button
              color="primary"
              variant="solid"
              icon={<SearchOutlined />}
              iconPlacement="start"
            >
              Search
            </Button>
            <Button
              color="primary"
              variant="solid"
              icon={<SearchOutlined />}
              iconPlacement="end"
            >
              Search
            </Button>
            <Button variant="dashed" icon={<CloudDownloadOutlined />}>
              Download
            </Button>
            <Button variant="filled" icon={<SearchOutlined />}>
              Search
            </Button>
            <Button variant="text" icon={<SearchOutlined />}>
              Search
            </Button>
            <Button variant="link" icon={<SearchOutlined />}>
              Search
            </Button>
          </Space>
          <Label>icon only</Label>
          <Space wrap>
            <Button color="primary" variant="solid" icon={<SearchOutlined />} />
            <Button icon={<SearchOutlined />} />
            <Button variant="dashed" icon={<SearchOutlined />} />
            <Button variant="filled" icon={<SearchOutlined />} />
            <Button variant="text" icon={<SearchOutlined />} />
            <Button variant="link" icon={<SearchOutlined />} />
          </Space>
        </Flex>
      </Section>

      <Section title="Disabled">
        <Space wrap>
          {variants.map((v) => (
            <Button key={v} variant={v} color="primary" disabled>
              {v}
            </Button>
          ))}
        </Space>
      </Section>

      <Section title="Loading">
        <Flex vertical gap={8}>
          <Space wrap>
            <Button color="primary" variant="solid" loading>
              Loading
            </Button>
            <Button loading>Loading</Button>
            <Button variant="dashed" loading>
              Dashed
            </Button>
            <Button variant="filled" loading>
              Filled
            </Button>
            <Button variant="text" loading>
              Text
            </Button>
            <Button variant="link" loading>
              Link
            </Button>
          </Space>
          <Space wrap>
            <Button color="primary" variant="solid" loading shape="circle" />
            <Button loading shape="circle" />
            <Button
              color="primary"
              variant="solid"
              loading={{ icon: <LoadingOutlined /> }}
            >
              Custom Icon
            </Button>
          </Space>
        </Flex>
      </Section>

      <Section title="Ghost">
        <div
          style={{
            background:
              "color-mix(in srgb, var(--ant-color-text) 60%, transparent)",
            padding: 24,
            borderRadius: 8,
          }}
        >
          <Space wrap>
            {variants
              .filter((v) => v !== "text" && v !== "link")
              .map((v) => (
                <Button key={v} variant={v} color="primary" ghost>
                  {v}
                </Button>
              ))}
            <Button color="danger" variant="solid" ghost>
              Danger
            </Button>
            <Button variant="outlined" ghost disabled>
              Disabled
            </Button>
          </Space>
        </div>
      </Section>

      <Section title="Danger">
        <Space wrap>
          <Button variant="solid" color="primary" danger>
            Primary
          </Button>
          <Button variant="outlined" danger>
            Default
          </Button>
          <Button variant="dashed" danger>
            Dashed
          </Button>
          <Button variant="text" danger>
            Text
          </Button>
          <Button variant="link" danger>
            Link
          </Button>
        </Space>
      </Section>

      <Section title="States">
        <PseudoStates>
          <Button color="primary" variant="solid">
            Primary
          </Button>
        </PseudoStates>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {}

export const Interactive: Story = {
  args: {
    variant: "solid",
    color: "primary",
    children: "Click Me",
    onClick: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.getByRole("button").click()
    await expect(args.onClick).toHaveBeenCalledTimes(1)
  },
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Col, Flex, Row, Tag } from "antd"
import { Section } from "../helpers"

const meta: Meta<typeof Row> = {
  title: "Antd/Layout/Grid",
  component: Row,
  parameters: { layout: "padded" },
  args: { gutter: 8 },
  argTypes: {
    justify: {
      control: "radio",
      options: [
        "start",
        "center",
        "end",
        "space-around",
        "space-between",
        "space-evenly",
      ],
    },
    align: {
      control: "radio",
      options: ["top", "middle", "bottom", "stretch"],
    },
    gutter: {
      control: "select",
      options: [0, 8, 16, 24],
    },
    wrap: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof Row>

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="Basic">
        <Flex vertical gap={8}>
          <Row>
            <Col span={24}>
              <Tag
                color="blue"
                style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
              >
                col-24
              </Tag>
            </Col>
          </Row>
          <Row gutter={8}>
            <Col span={12}>
              <Tag
                color="blue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-12
              </Tag>
            </Col>
            <Col span={12}>
              <Tag
                color="geekblue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-12
              </Tag>
            </Col>
          </Row>
          <Row gutter={8}>
            <Col span={8}>
              <Tag
                color="blue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-8
              </Tag>
            </Col>
            <Col span={8}>
              <Tag
                color="geekblue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-8
              </Tag>
            </Col>
            <Col span={8}>
              <Tag
                color="blue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-8
              </Tag>
            </Col>
          </Row>
          <Row gutter={8}>
            <Col span={6}>
              <Tag
                color="blue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-6
              </Tag>
            </Col>
            <Col span={6}>
              <Tag
                color="geekblue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-6
              </Tag>
            </Col>
            <Col span={6}>
              <Tag
                color="blue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-6
              </Tag>
            </Col>
            <Col span={6}>
              <Tag
                color="geekblue"
                style={{
                  width: "100%",
                  padding: "8px 0",
                  textAlign: "center",
                }}
              >
                col-6
              </Tag>
            </Col>
          </Row>
        </Flex>
      </Section>
      <Section title="Horizontal gutter: 16">
        <Row gutter={16}>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
        </Row>
      </Section>
      <Section title="Horizontal & vertical gutter: [16, 16]">
        <Row gutter={[16, 16]}>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="geekblue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
          <Col span={6}>
            <Tag
              color="blue"
              style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
            >
              col-6
            </Tag>
          </Col>
        </Row>
      </Section>
    </Flex>
  ),
}

export const Playground: Story = {
  render: (args) => (
    <Row {...args}>
      <Col span={12}>
        <Tag
          color="blue"
          style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
        >
          col-12
        </Tag>
      </Col>
      <Col span={12}>
        <Tag
          color="geekblue"
          style={{ width: "100%", padding: "8px 0", textAlign: "center" }}
        >
          col-12
        </Tag>
      </Col>
    </Row>
  ),
}

import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button, Flex, Result, Space } from "antd"
import { expect, fn, userEvent, within } from "storybook/test"
import { Section } from "../helpers"

const meta: Meta<typeof Result> = {
  title: "Antd/Feedback/Result",
  component: Result,
  parameters: { layout: "padded" },
  args: {
    status: "info",
    title: "Your operation has been executed",
    subTitle: "The system is processing, please wait.",
  },
  argTypes: {
    status: {
      control: "select",
      options: ["success", "error", "info", "warning", "403", "404", "500"],
    },
    title: { control: "text" },
    subTitle: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof Result>

// ── Gallery ──────────────────────────────────────────────────────

export const Gallery: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <Flex vertical gap={24}>
      <Section title="All Statuses">
        <Flex vertical gap={16}>
          <Result
            status="success"
            title="Successfully Purchased Cloud Server ECS!"
            subTitle="Order number: 2017182818828182881. Cloud server configuration takes 1-5 minutes."
          />
          <Result
            status="error"
            title="Submission Failed"
            subTitle="Please check and modify the following information before resubmitting."
          />
          <Result
            status="info"
            title="Your operation has been executed"
            subTitle="The system is processing, please wait."
          />
          <Result
            status="warning"
            title="There are some issues"
            subTitle="Please check the details before proceeding."
          />
          <Result
            status="404"
            title="404"
            subTitle="Sorry, the page you visited does not exist."
          />
          <Result
            status="403"
            title="403"
            subTitle="Sorry, you are not authorized to access this page."
          />
          <Result
            status="500"
            title="500"
            subTitle="Sorry, something went wrong."
          />
        </Flex>
      </Section>

      <Section title="With Extra Actions">
        <Flex vertical gap={16}>
          <Result
            status="success"
            title="Successfully Purchased Cloud Server ECS!"
            subTitle="Order number: 2017182818828182881. Cloud server configuration takes 1-5 minutes."
            extra={[
              <Button type="primary" key="console">
                Go Console
              </Button>,
              <Button key="buy">Buy Again</Button>,
            ]}
          />
          <Result
            status="error"
            title="Submission Failed"
            subTitle="Please check and modify the following information before resubmitting."
            extra={
              <Space>
                <Button type="primary">Edit</Button>
                <Button>Cancel</Button>
              </Space>
            }
          />
          <Result
            status="404"
            title="404"
            subTitle="Sorry, the page you visited does not exist."
            extra={<Button type="primary">Back Home</Button>}
          />
        </Flex>
      </Section>
    </Flex>
  ),
}

// ── Interactive ───────────────────────────────────────────────────

export const Interactive: Story = {
  parameters: { controls: { disable: true } },
  args: {
    status: "success",
    title: "Operation completed",
    subTitle: "Your changes have been saved.",
    extra: undefined,
  },
  render: (args) => {
    const onClick = fn()
    return (
      <Result
        {...args}
        extra={
          <Button type="primary" data-testid="action-btn" onClick={onClick}>
            View Details
          </Button>
        }
      />
    )
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByTestId("action-btn")
    await userEvent.click(button)
    await expect(button).toHaveTextContent("View Details")
  },
}

// ── Playground ───────────────────────────────────────────────────

export const Playground: Story = {}

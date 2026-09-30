import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
  Slider,
  Space,
  Switch,
} from "antd";
import { Section } from "../helpers";

const meta: Meta = {
  title: "Combinations/Form Page",
};

export default meta;

type Story = StoryObj;

function FormPage() {
  return (
    <div style={{ maxWidth: 760 }}>
      <Section
        description="Mixed controls in one vertical form: the skin's line/label/focus vocabulary."
        title="Mission scheduling"
      >
        <Card>
          <Form labelCol={{ span: 6 }} onFinish={() => {}} wrapperCol={{ span: 16 }}>
            <Form.Item
              label="Mission name"
              name="name"
              rules={[{ required: true, message: "Name is required" }]}
            >
              <Input placeholder="River crossing sweep" />
            </Form.Item>
            <Form.Item label="Robot fleet" name="fleet">
              <Select
                mode="multiple"
                options={[
                  { value: "north", label: "North substation" },
                  { value: "east", label: "East corridor" },
                  { value: "river", label: "River crossing" },
                ]}
                placeholder="Select fleets"
              />
            </Form.Item>
            <Form.Item label="Window" name="window">
              <DatePicker.RangePicker showTime style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item label="Battery floor" name="floor" wrapperCol={{ span: 10 }}>
              <Slider defaultValue={30} tooltip={{ formatter: (v) => `${v}%` }} />
            </Form.Item>
            <Form.Item label="Repeat" name="repeat">
              <Radio.Group defaultValue="once">
                <Radio value="once">Once</Radio>
                <Radio value="daily">Daily</Radio>
                <Radio value="custom">Custom</Radio>
              </Radio.Group>
            </Form.Item>
            <Form.Item label="Retries" name="retries" wrapperCol={{ span: 8 }}>
              <InputNumber max={5} min={0} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item label="Notify on failure" name="notify" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item label="Notes" name="notes">
              <Input.TextArea placeholder="Operator notes" rows={3} />
            </Form.Item>
            <Form.Item wrapperCol={{ offset: 6, span: 16 }}>
              <Space>
                <Button htmlType="submit" type="primary">
                  Schedule
                </Button>
                <Button htmlType="reset">Reset</Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <FormPage />,
};

export const TwoColumn: Story = {
  render: () => (
    <div style={{ maxWidth: 900 }}>
      <Section title="Two-column layout">
        <Card>
          <Form onFinish={() => {}}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label="Field A" name="a">
                  <Input />
                </Form.Item>
                <Form.Item label="Field C" name="c">
                  <Select options={[{ value: 1, label: "One" }]} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Field B" name="b">
                  <Input />
                </Form.Item>
                <Form.Item label="Field D" name="d">
                  <DatePicker style={{ width: "100%" }} />
                </Form.Item>
              </Col>
            </Row>
            <Button htmlType="submit" type="primary">
              Submit
            </Button>
          </Form>
        </Card>
      </Section>
    </div>
  ),
};

import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Checkbox,
  DatePicker,
  Input,
  InputNumber,
  Radio,
  Segmented,
  Select,
  Slider,
  Space,
  Switch,
} from "antd";
import { useState } from "react";
import { Label, Section } from "../helpers";

const meta: Meta<typeof Input> = {
  title: "Components/Inputs",
  component: Input,
};

export default meta;

type Story = StoryObj<typeof Input>;

function InputsPage() {
  const [seg, setSeg] = useState("map");
  return (
    <div style={{ maxWidth: 760 }}>
      <Section title="Text inputs">
        <Space vertical size={12} style={{ display: "flex", maxWidth: 380 }}>
          <Input placeholder="Basic input" />
          <Input allowClear placeholder="With clear" prefix="@" />
          <Input disabled placeholder="Disabled" />
          <Input status="error" placeholder="Error status" />
          <Input status="warning" placeholder="Warning status" />
          <Input.Password placeholder="Password" />
          <Input.TextArea placeholder="Multiline" rows={3} />
          <Input.Search placeholder="Search" />
          <InputNumber placeholder="Number" style={{ width: 180 }} />
        </Space>
      </Section>

      <Section title="Variant filled">
        <Space vertical size={12} style={{ display: "flex", maxWidth: 380 }}>
          <Input placeholder="Filled variant" variant="filled" />
          <Select
            options={[
              { value: 1, label: "Option A" },
              { value: 2, label: "Option B" },
            ]}
            placeholder="Filled select"
            style={{ width: 220 }}
            variant="filled"
          />
        </Space>
      </Section>

      <Section title="Select & date">
        <Space wrap>
          <Select
            defaultValue="a"
            options={[
              { value: "a", label: "Alpha" },
              { value: "b", label: "Beta" },
              { value: "c", label: "Gamma", disabled: true },
            ]}
            style={{ width: 180 }}
          />
          <Select
            defaultValue={["a"]}
            labelRender={({ label }) => label}
            maxTagCount="responsive"
            mode="multiple"
            options={[
              { value: "a", label: "Alpha" },
              { value: "b", label: "Beta" },
              { value: "c", label: "Gamma" },
            ]}
            style={{ minWidth: 220 }}
          />
          <DatePicker />
          <DatePicker.RangePicker />
        </Space>
      </Section>

      <Section title="Boolean controls">
        <Space vertical size={12}>
          <Space size={20}>
            <Switch defaultChecked />
            <Switch />
            <Switch checkedChildren="ON" defaultChecked unCheckedChildren="OFF" />
            <Switch disabled />
          </Space>
          <Space size={20}>
            <Checkbox defaultChecked>Checked</Checkbox>
            <Checkbox>Unchecked</Checkbox>
            <Checkbox disabled>Disabled</Checkbox>
            <Checkbox indeterminate>Indeterminate</Checkbox>
          </Space>
          <Radio.Group defaultValue="a">
            <Radio value="a">Alpha</Radio>
            <Radio value="b">Beta</Radio>
            <Radio disabled value="c">
              Gamma
            </Radio>
          </Radio.Group>
        </Space>
      </Section>

      <Section title="Slider & segmented">
        <Space vertical size={16} style={{ display: "flex", maxWidth: 420 }}>
          <div>
            <Label>slider</Label>
            <Slider defaultValue={42} />
          </div>
          <div>
            <Label>segmented</Label>
            <Segmented
              onChange={(v) => setSeg(v as string)}
              options={[
                { value: "list", label: "List" },
                { value: "map", label: "Map" },
                { value: "grid", label: "Grid" },
              ]}
              value={seg}
            />
          </div>
        </Space>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <InputsPage />,
};

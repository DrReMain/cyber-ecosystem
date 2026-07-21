import { createFileRoute } from "@tanstack/react-router";
import { Button, DatePicker, Space } from "antd";

export const Route = createFileRoute("/_app/antd")({ component: AntdDemo });

function AntdDemo() {
  return (
    <div className="space-y-8">
      <section className="space-y-3">
        <h2 className="font-semibold text-lg">Button</h2>
        <Space wrap>
          <Button type="primary">Primary</Button>
          <Button>Default</Button>
          <Button type="dashed">Dashed</Button>
          <Button type="text">Text</Button>
          <Button type="link">Link</Button>
          <Button disabled>Disabled</Button>
          <Button loading>Loading</Button>
        </Space>
      </section>

      <section className="space-y-3">
        <h2 className="font-semibold text-lg">DatePicker</h2>
        <Space orientation="vertical">
          <DatePicker />
          <DatePicker.RangePicker />
        </Space>
      </section>
    </div>
  );
}

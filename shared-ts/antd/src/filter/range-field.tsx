import { DatePicker, Form, InputNumber } from "antd";
import type { NamePath } from "antd/es/form/interface";
import { MoveRight } from "lucide-react";
import { rangeTupleFromEvent, rangeTupleToProps } from "./date-utils";
import type { RangeFilterOption } from "./types";

const Item = Form.Item;

export function RangeField({
  name,
  placeholder,
  type,
}: Pick<RangeFilterOption, "name" | "placeholder" | "type">) {
  if (type === "range-number") {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          width: "100%",
          gap: "8px",
        }}
      >
        <Item style={{ flex: 1, marginBottom: 0 }} name={name[0] as NamePath} noStyle>
          <InputNumber style={{ width: "100%" }} placeholder={placeholder?.[0]} />
        </Item>
        <MoveRight
          size={14}
          style={{ flexShrink: 0, color: "var(--ant-color-text-quaternary, inherit)" }}
        />
        <Item style={{ flex: 1, marginBottom: 0 }} name={name[1] as NamePath} noStyle>
          <InputNumber style={{ width: "100%" }} placeholder={placeholder?.[1]} />
        </Item>
      </div>
    );
  }

  // One form key carries the pair; Filter translates tuple ↔ name[0]/name[1]
  // at the initialValues/onFinish boundaries, so callers keep flat keys.
  // allowEmpty + needConfirm make a one-sided range a first-class commit.
  return (
    <Item
      getValueFromEvent={rangeTupleFromEvent}
      getValueProps={rangeTupleToProps}
      name={name[0] as NamePath}
      noStyle
    >
      <DatePicker.RangePicker
        allowEmpty={[true, true]}
        needConfirm={false}
        placeholder={placeholder}
        showTime={type === "range-datetime" || undefined}
        style={{ width: "100%" }}
      />
    </Item>
  );
}

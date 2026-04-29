import { DatePicker, Form, InputNumber } from "antd"
import type { NamePath } from "antd/es/form/interface"
import { MoveRight } from "lucide-react"
import { dayjsFromEvent, dayjsToProps } from "./date-utils"
import type { RangeFilterOption } from "./types"

const Item = Form.Item

export function RangeField({
  name,
  placeholder,
  type,
}: Pick<RangeFilterOption, "name" | "placeholder" | "type">) {
  const isNumber = type === "range-number"
  const showTime = type === "range-datetime"
  const extraProps = isNumber
    ? {}
    : { getValueProps: dayjsToProps, getValueFromEvent: dayjsFromEvent }

  const renderInput = (idx: 0 | 1) =>
    isNumber ? (
      <InputNumber style={{ width: "100%" }} placeholder={placeholder?.[idx]} />
    ) : (
      <DatePicker
        style={{ width: "100%" }}
        showTime={showTime || undefined}
        placeholder={placeholder?.[idx]}
      />
    )

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        width: "100%",
        gap: "8px",
      }}
    >
      <Item
        style={{ flex: 1, marginBottom: 0 }}
        name={name[0] as NamePath}
        noStyle
        {...extraProps}
      >
        {renderInput(0)}
      </Item>
      <MoveRight
        size={14}
        style={{ flexShrink: 0, color: "var(--antd-color-text-quaternary)" }}
      />
      <Item
        style={{ flex: 1, marginBottom: 0 }}
        name={name[1] as NamePath}
        noStyle
        {...extraProps}
      >
        {renderInput(1)}
      </Item>
    </div>
  )
}

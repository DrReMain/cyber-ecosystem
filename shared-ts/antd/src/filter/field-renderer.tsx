import { DatePicker, Input, InputNumber, Select } from "antd"
import type {
  DateFilterOption,
  InputFilterOption,
  NumberFilterOption,
  SelectFilterOption,
} from "./types"

type SingleFieldOption =
  | InputFilterOption
  | SelectFilterOption
  | DateFilterOption
  | NumberFilterOption

// biome-ignore lint/suspicious/noExplicitAny: antd Form.Item injects typed props at runtime
type FormControlProps = { value?: any; onChange?: (...args: any[]) => void }

export function FieldRenderer({
  option,
  ...control
}: { option: SingleFieldOption } & FormControlProps) {
  switch (option.type) {
    case "select":
      return (
        <Select
          allowClear
          options={option.options}
          placeholder={option.placeholder}
          mode={option.mode}
          style={{ width: "100%" }}
          {...control}
        />
      )
    case "date":
      return (
        <DatePicker
          allowClear
          picker={option.picker}
          showTime={option.showTime}
          placeholder={option.placeholder}
          style={{ width: "100%" }}
          {...control}
        />
      )
    case "number":
      return (
        <InputNumber
          placeholder={option.placeholder}
          min={option.min}
          max={option.max}
          style={{ width: "100%" }}
          {...control}
        />
      )
    case "input":
    default:
      return <Input allowClear placeholder={option.placeholder} {...control} />
  }
}

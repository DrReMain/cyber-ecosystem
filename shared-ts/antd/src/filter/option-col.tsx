import { Col, Form } from "antd"
import { dayjsFromEvent, dayjsToProps } from "./date-utils"
import { FieldRenderer } from "./field-renderer"
import { RangeField } from "./range-field"
import type {
  ElementFilterOption,
  FilterOption,
  RangeFilterOption,
} from "./types"

const Item = Form.Item

function isRangeOption(opt: FilterOption): opt is RangeFilterOption {
  return Array.isArray(opt.name)
}

function isElementOption(opt: FilterOption): opt is ElementFilterOption {
  return "element" in opt && opt.element !== undefined
}

export function OptionCol({
  option,
  hidden,
  spanProps,
}: {
  option: FilterOption
  hidden: boolean
  spanProps: Record<string, number>
}) {
  if (isRangeOption(option)) {
    return (
      <Col {...spanProps}>
        <Item hidden={hidden} label={option.label} style={{ marginBottom: 0 }}>
          <RangeField
            name={option.name}
            placeholder={option.placeholder}
            type={option.type}
          />
        </Item>
      </Col>
    )
  }

  if (isElementOption(option)) {
    return (
      <Col {...spanProps}>
        <Item
          hidden={hidden}
          label={option.label}
          name={option.name}
          style={{ marginBottom: 0 }}
        >
          {option.element}
        </Item>
      </Col>
    )
  }

  const dateExtra =
    option.type === "date"
      ? { getValueProps: dayjsToProps, getValueFromEvent: dayjsFromEvent }
      : {}

  return (
    <Col {...spanProps}>
      <Item
        hidden={hidden}
        label={option.label}
        name={option.name}
        style={{ marginBottom: 0 }}
        {...dateExtra}
      >
        <FieldRenderer option={option} />
      </Item>
    </Col>
  )
}

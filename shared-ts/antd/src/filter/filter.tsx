import { Button, Col, Form, Grid, Row, Space } from "antd"
import { ChevronDown, ChevronUp, Search, Undo2 } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { OptionCol } from "./option-col"
import type { FilterProps } from "./types"

const Item = Form.Item
const useBreakpoint = Grid.useBreakpoint

const DEFAULT_COL_SPAN = {
  xs: 24,
  sm: 12,
  md: 12,
  lg: 8,
  xl: 6,
  xxl: 6,
  xxxl: 4,
}
const HIDDEN_SPAN = { xs: 0, sm: 0, md: 0, lg: 0, xl: 0, xxl: 0, xxxl: 0 }

function isRangeOption(opt: {
  name: string | [string, string]
}): opt is { name: [string, string] } {
  return Array.isArray(opt.name)
}

export function Filter<
  T extends Record<string, unknown> = Record<string, unknown>,
>({
  options,
  onFilter,
  onReset,
  disabled,
  initialValues,
  columns,
  size,
  labels,
  buttonVariant = "icon-text",
}: FilterProps<T>) {
  const [form] = Form.useForm<T>()
  const screens = useBreakpoint()
  const [expanded, setExpanded] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    if (Object.keys(screens).length > 0) setMounted(true)
  }, [screens])

  // Sync form with URL-derived initialValues
  // setFieldsValue is a merge — we must explicitly clear fields removed from the URL
  useEffect(() => {
    if (initialValues) {
      const iv = initialValues as Record<string, unknown>
      const current = form.getFieldsValue() as Record<string, unknown>
      const keysToReset = Object.keys(current).filter((k) => !(k in iv))
      if (keysToReset.length > 0) form.resetFields(keysToReset as never)
      form.setFieldsValue(iv as Parameters<typeof form.setFieldsValue>[0])
    }
  }, [form, initialValues])

  const colSpan = useMemo(() => {
    if (!columns) return DEFAULT_COL_SPAN
    const span = 24 / columns
    return { ...DEFAULT_COL_SPAN, lg: span, xl: span, xxl: span, xxxl: span }
  }, [columns])

  const visibleOptions = useMemo(
    () => options.filter((o) => o.show !== false),
    [options],
  )

  const colsPerRow = useMemo(() => {
    if (screens.xxxl) return 24 / colSpan.xxxl
    if (screens.xxl) return 24 / colSpan.xxl
    if (screens.xl) return 24 / colSpan.xl
    if (screens.lg) return 24 / colSpan.lg
    if (screens.md) return 24 / colSpan.md
    if (screens.sm) return 24 / colSpan.sm
    return 1
  }, [screens, colSpan])

  const total = visibleOptions.length
  const showToggle = total >= colsPerRow

  const isHidden = (index: number): boolean => {
    if (colsPerRow === 1) return !expanded && index >= 1
    return !expanded && index >= colsPerRow - 1
  }

  const buttonOffset = useMemo(() => {
    if (colsPerRow === 1) return 0
    const displayed = expanded ? total : Math.min(total, colsPerRow - 1)
    const remainder = displayed % colsPerRow
    const gap = colsPerRow - 1 - remainder
    return (gap * 24) / colsPerRow
  }, [expanded, total, colsPerRow])

  const handleFinish = (values: T) => {
    onFilter?.(values)
  }

  const handleReset = () => {
    if (onReset) onReset()
    else handleFinish(form.getFieldsValue() as T)
  }

  const {
    search: searchLabel = "Search",
    reset: resetLabel = "Reset",
    fold: foldLabel = "Fold",
    expand: expandLabel = "Expand",
  } = labels ?? {}

  const showIcon = buttonVariant !== "text"
  const showText = buttonVariant !== "icon"

  return (
    <Form
      form={form}
      initialValues={initialValues}
      layout="vertical"
      size={size}
      disabled={disabled}
      onFinish={handleFinish}
      onReset={handleReset}
      requiredMark={(label) => label}
    >
      {!mounted && (
        <>
          <style>{`@keyframes filter-pulse{0%,100%{opacity:1}50%{opacity:.4}}`}</style>
          <div
            style={{
              height: 60,
              borderRadius: 6,
              background: "var(--ant-color-fill-secondary, rgba(0,0,0,.06))",
              animation: "filter-pulse 1.5s ease-in-out infinite",
            }}
          />
        </>
      )}
      {mounted && (
        <Row gutter={[8, 8]} wrap>
          {visibleOptions.map((opt, i) => {
            const hidden = isHidden(i)
            return (
              <OptionCol
                key={
                  isRangeOption(opt)
                    ? `${opt.name[0]}-${opt.name[1]}`
                    : opt.name
                }
                option={opt}
                hidden={hidden}
                spanProps={hidden ? HIDDEN_SPAN : colSpan}
              />
            )
          })}

          <Col {...colSpan} offset={buttonOffset}>
            <Item label=" " style={{ marginBottom: 0 }}>
              <Space>
                <Button
                  color="primary"
                  htmlType="submit"
                  icon={showIcon ? <Search size={14} /> : undefined}
                  variant="filled"
                >
                  {showText ? searchLabel : undefined}
                </Button>
                <Button
                  color="default"
                  htmlType="reset"
                  icon={showIcon ? <Undo2 size={14} /> : undefined}
                  variant="filled"
                >
                  {showText ? resetLabel : undefined}
                </Button>
                {showToggle && (
                  <Button
                    size="small"
                    color="primary"
                    icon={
                      expanded ? (
                        <ChevronUp size={14} />
                      ) : (
                        <ChevronDown size={14} />
                      )
                    }
                    variant="link"
                    onClick={() => setExpanded((v) => !v)}
                  >
                    {expanded ? foldLabel : expandLabel}
                  </Button>
                )}
              </Space>
            </Item>
          </Col>
        </Row>
      )}
    </Form>
  )
}

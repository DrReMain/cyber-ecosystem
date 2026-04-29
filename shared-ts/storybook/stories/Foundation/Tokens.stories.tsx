import type { Meta, StoryObj } from "@storybook/react-vite"
import { Card, Flex, Typography, theme } from "antd"
import { Label, Section } from "../helpers"

const meta: Meta = {
  title: "Antd/Foundation/Tokens",
  parameters: { layout: "padded", controls: { disable: true } },
}

export default meta
type Story = StoryObj

const { Title, Text, Paragraph } = Typography

function ColorSwatch({ color, label }: { color: string; label: string }) {
  return (
    <Flex vertical align="center" gap={4}>
      <div
        style={{
          width: 64,
          height: 64,
          borderRadius: 8,
          background: color,
          border:
            "1px solid color-mix(in srgb, var(--ant-color-text) 15%, transparent)",
        }}
      />
      <Text style={{ fontSize: 11, textAlign: "center" }}>{label}</Text>
      <Text type="secondary" style={{ fontSize: 10 }}>
        {color}
      </Text>
    </Flex>
  )
}

function TokenRow({
  label,
  value,
  preview,
}: {
  label: string
  value: string | number
  preview?: React.ReactNode
}) {
  return (
    <Flex justify="space-between" align="center" style={{ padding: "4px 0" }}>
      <Text code style={{ fontSize: 12 }}>
        {label}
      </Text>
      <Flex align="center" gap={8}>
        {preview}
        <Text type="secondary" style={{ fontSize: 11 }}>
          {String(value)}
        </Text>
      </Flex>
    </Flex>
  )
}

export const Gallery: Story = {
  render: () => {
    const { token } = theme.useToken()

    const brandColors = [
      { key: "colorPrimary", val: token.colorPrimary },
      { key: "colorInfo", val: token.colorInfo },
      { key: "colorSuccess", val: token.colorSuccess },
      { key: "colorWarning", val: token.colorWarning },
      { key: "colorError", val: token.colorError },
    ]

    const bgColors = [
      { key: "colorBgBase", val: token.colorBgBase },
      { key: "colorBgLayout", val: token.colorBgLayout },
      { key: "colorBgContainer", val: token.colorBgContainer },
      { key: "colorBgElevated", val: token.colorBgElevated },
    ]

    const textColors = [
      { key: "colorTextBase", val: token.colorTextBase },
      { key: "colorText", val: token.colorText },
      { key: "colorTextSecondary", val: token.colorTextSecondary },
      { key: "colorTextTertiary", val: token.colorTextTertiary },
      { key: "colorTextQuaternary", val: token.colorTextQuaternary },
      { key: "colorLink", val: token.colorLink },
    ]

    const borderColors = [
      { key: "colorBorder", val: token.colorBorder },
      { key: "colorBorderSecondary", val: token.colorBorderSecondary },
    ]

    const fillColors = [
      { key: "colorFill", val: token.colorFill },
      { key: "colorFillSecondary", val: token.colorFillSecondary },
      { key: "colorFillTertiary", val: token.colorFillTertiary },
      { key: "colorFillQuaternary", val: token.colorFillQuaternary },
    ]

    const spacings = [
      { key: "paddingXXS", val: token.paddingXXS },
      { key: "paddingXS", val: token.paddingXS },
      { key: "paddingSM", val: token.paddingSM },
      { key: "padding", val: token.padding },
      { key: "paddingMD", val: token.paddingMD },
      { key: "paddingLG", val: token.paddingLG },
      { key: "paddingXL", val: token.paddingXL },
      { key: "marginXXS", val: token.marginXXS },
      { key: "marginXS", val: token.marginXS },
      { key: "marginSM", val: token.marginSM },
      { key: "margin", val: token.margin },
      { key: "marginMD", val: token.marginMD },
      { key: "marginLG", val: token.marginLG },
      { key: "marginXL", val: token.marginXL },
    ]

    const radii = [
      { key: "borderRadiusXS", val: token.borderRadiusXS },
      { key: "borderRadiusSM", val: token.borderRadiusSM },
      { key: "borderRadius", val: token.borderRadius },
      { key: "borderRadiusLG", val: token.borderRadiusLG },
    ]

    const fonts = [
      { key: "fontSizeSM", val: token.fontSizeSM },
      { key: "fontSize", val: token.fontSize },
      { key: "fontSizeLG", val: token.fontSizeLG },
      { key: "fontSizeXL", val: token.fontSizeXL },
      { key: "fontSizeHeading1", val: token.fontSizeHeading1 },
      { key: "fontSizeHeading2", val: token.fontSizeHeading2 },
      { key: "fontSizeHeading3", val: token.fontSizeHeading3 },
      { key: "fontSizeHeading4", val: token.fontSizeHeading4 },
      { key: "fontSizeHeading5", val: token.fontSizeHeading5 },
    ]

    const lineHeights = [
      { key: "lineHeight", val: token.lineHeight },
      { key: "lineHeightLG", val: token.lineHeightLG },
      { key: "lineHeightHeading1", val: token.lineHeightHeading1 },
      { key: "lineHeightHeading2", val: token.lineHeightHeading2 },
      { key: "lineHeightHeading3", val: token.lineHeightHeading3 },
      { key: "lineHeightHeading4", val: token.lineHeightHeading4 },
      { key: "lineHeightHeading5", val: token.lineHeightHeading5 },
    ]

    return (
      <Flex vertical gap={32} style={{ maxWidth: 900 }}>
        <Section title="Color Palette">
          <Flex vertical gap={20}>
            {[
              { label: "Brand & Semantic", colors: brandColors },
              { label: "Background", colors: bgColors },
              { label: "Text", colors: textColors },
              { label: "Border", colors: borderColors },
              { label: "Fill", colors: fillColors },
            ].map((group) => (
              <div key={group.label}>
                <Label>{group.label}</Label>
                <Flex wrap gap={12}>
                  {group.colors.map((c) => (
                    <ColorSwatch key={c.key} color={c.val} label={c.key} />
                  ))}
                </Flex>
              </div>
            ))}
          </Flex>
        </Section>

        <Section title="Typography Scale">
          <Card size="small">
            <Flex vertical gap={4}>
              {fonts.map((f) => (
                <TokenRow
                  key={f.key}
                  label={f.key}
                  value={f.val}
                  preview={
                    <span style={{ fontSize: f.val, lineHeight: 1 }}>Aa</span>
                  }
                />
              ))}
            </Flex>
          </Card>
          <div style={{ marginTop: 16 }}>
            <Title level={1}>Heading 1</Title>
            <Title level={2}>Heading 2</Title>
            <Title level={3}>Heading 3</Title>
            <Title level={4}>Heading 4</Title>
            <Title level={5}>Heading 5</Title>
            <Paragraph>Body text — default paragraph</Paragraph>
            <Text type="secondary">Secondary text</Text>
          </div>
        </Section>

        <Section title="Line Heights">
          <Card size="small">
            <Flex vertical gap={4}>
              {lineHeights.map((lh) => (
                <TokenRow key={lh.key} label={lh.key} value={lh.val} />
              ))}
            </Flex>
          </Card>
        </Section>

        <Section title="Spacing Scale">
          <Card size="small">
            <Flex vertical gap={4}>
              {spacings.map((s) => (
                <TokenRow
                  key={s.key}
                  label={s.key}
                  value={s.val}
                  preview={
                    <div
                      style={{
                        width: Number(s.val) || 0,
                        height: 8,
                        background: token.colorPrimary,
                        borderRadius: 2,
                        opacity: 0.7,
                      }}
                    />
                  }
                />
              ))}
            </Flex>
          </Card>
        </Section>

        <Section title="Border Radius">
          <Flex wrap gap={16} align="end">
            {radii.map((r) => (
              <Flex key={r.key} vertical align="center" gap={4}>
                <div
                  style={{
                    width: 56,
                    height: 56,
                    borderRadius: r.val,
                    background: token.colorPrimary,
                    opacity: 0.6,
                  }}
                />
                <Text style={{ fontSize: 11 }}>{r.key}</Text>
                <Text type="secondary" style={{ fontSize: 10 }}>
                  {r.val}px
                </Text>
              </Flex>
            ))}
          </Flex>
        </Section>

        <Section title="Shadows">
          <Flex wrap gap={16}>
            {[
              { key: "boxShadow", val: token.boxShadow },
              { key: "boxShadowSecondary", val: token.boxShadowSecondary },
            ].map((s) => (
              <Card
                key={s.key}
                size="small"
                style={{ width: 200, boxShadow: s.val }}
              >
                <Text strong>{s.key}</Text>
                <br />
                <Text
                  type="secondary"
                  style={{ fontSize: 10, wordBreak: "break-all" }}
                >
                  {String(s.val)}
                </Text>
              </Card>
            ))}
          </Flex>
        </Section>

        <Section title="Font Family & Weight">
          <Card size="small">
            <TokenRow label="fontFamily" value={token.fontFamily} />
            <TokenRow label="fontWeightStrong" value={token.fontWeightStrong} />
          </Card>
        </Section>

        <Section title="Control Sizes">
          <Card size="small">
            <Flex vertical gap={4}>
              <TokenRow label="controlHeightXS" value={token.controlHeightXS} />
              <TokenRow label="controlHeightSM" value={token.controlHeightSM} />
              <TokenRow label="controlHeight" value={token.controlHeight} />
              <TokenRow label="controlHeightLG" value={token.controlHeightLG} />
            </Flex>
          </Card>
        </Section>

        <Section title="Motion / Animation">
          <Card size="small">
            <Flex vertical gap={8}>
              {[
                { key: "motionDurationSlow", val: token.motionDurationSlow },
                { key: "motionDurationMid", val: token.motionDurationMid },
                { key: "motionDurationFast", val: token.motionDurationFast },
              ].map((d) => (
                <TokenRow
                  key={d.key}
                  label={d.key}
                  value={d.val}
                  preview={
                    <div
                      style={{
                        width: 64,
                        height: 8,
                        borderRadius: 2,
                        background: token.colorPrimary,
                        opacity: 0.6,
                        animation: `token-motion-bar ${d.val} ease-in-out infinite alternate`,
                      }}
                    />
                  }
                />
              ))}
              <style>{`@keyframes token-motion-bar { from { width: 12px; } to { width: 64px; } }`}</style>
              {[
                { key: "motionEaseInOut", val: token.motionEaseInOut },
                { key: "motionEaseOut", val: token.motionEaseOut },
                { key: "motionEaseInBack", val: token.motionEaseInBack },
              ].map((e) => (
                <TokenRow key={e.key} label={e.key} value={String(e.val)} />
              ))}
            </Flex>
          </Card>
        </Section>

        <Section title="Screen Breakpoints">
          <Card size="small">
            <Flex vertical gap={8}>
              {[
                { key: "screenXS", val: token.screenXS },
                { key: "screenSM", val: token.screenSM },
                { key: "screenMD", val: token.screenMD },
                { key: "screenLG", val: token.screenLG },
                { key: "screenXL", val: token.screenXL },
                { key: "screenXXL", val: token.screenXXL },
              ].map((bp) => {
                const maxVal = Number(token.screenXXL) || 1
                const width = Math.max(
                  ((Number(bp.val) || 0) / maxVal) * 100,
                  4,
                )
                return (
                  <TokenRow
                    key={bp.key}
                    label={bp.key}
                    value={bp.val}
                    preview={
                      <div
                        style={{
                          width: `${width}%`,
                          height: 8,
                          background: token.colorPrimary,
                          borderRadius: 2,
                          opacity: 0.6,
                          maxWidth: 200,
                        }}
                      />
                    }
                  />
                )
              })}
            </Flex>
          </Card>
        </Section>

        <Section title="Z-Index">
          <Card size="small">
            <Flex vertical gap={4}>
              {[
                { key: "zIndexBase", val: token.zIndexBase },
                { key: "zIndexPopupBase", val: token.zIndexPopupBase },
              ].map((z) => {
                const maxVal =
                  Math.max(token.zIndexBase, token.zIndexPopupBase) || 1
                const width = ((Number(z.val) || 0) / maxVal) * 100
                return (
                  <TokenRow
                    key={z.key}
                    label={z.key}
                    value={z.val}
                    preview={
                      <div
                        style={{
                          width: `${width}%`,
                          height: 8,
                          background: token.colorPrimary,
                          borderRadius: 2,
                          opacity: 0.6,
                          maxWidth: 200,
                        }}
                      />
                    }
                  />
                )
              })}
            </Flex>
          </Card>
        </Section>

        <Section title="Line Width">
          <Card size="small">
            <Flex vertical gap={4}>
              <TokenRow
                label="lineWidth"
                value={token.lineWidth}
                preview={
                  <div
                    style={{
                      width: 64,
                      height: 0,
                      borderTop: `${token.lineWidth}px solid ${token.colorPrimary}`,
                    }}
                  />
                }
              />
              <TokenRow
                label="lineWidthFocus"
                value={token.lineWidthFocus}
                preview={
                  <div
                    style={{
                      width: 64,
                      height: 0,
                      borderTop: `${token.lineWidthFocus}px solid ${token.colorPrimary}`,
                    }}
                  />
                }
              />
              <TokenRow label="lineType" value={token.lineType} />
            </Flex>
          </Card>
        </Section>

        <Section title="Interactive Color Scale">
          <Card size="small">
            <Flex vertical gap={4}>
              {[
                { key: "colorPrimaryBg", val: token.colorPrimaryBg },
                {
                  key: "colorPrimaryBgHover",
                  val: token.colorPrimaryBgHover,
                },
                {
                  key: "colorPrimaryBorder",
                  val: token.colorPrimaryBorder,
                },
                {
                  key: "colorPrimaryBorderHover",
                  val: token.colorPrimaryBorderHover,
                },
                { key: "colorPrimaryHover", val: token.colorPrimaryHover },
                { key: "colorPrimary", val: token.colorPrimary },
                { key: "colorPrimaryActive", val: token.colorPrimaryActive },
              ].map((c) => (
                <TokenRow
                  key={c.key}
                  label={c.key}
                  value={String(c.val)}
                  preview={
                    <div
                      style={{
                        width: 48,
                        height: 16,
                        borderRadius: 4,
                        background: c.val,
                      }}
                    />
                  }
                />
              ))}
            </Flex>
            <div
              style={{
                height: 16,
                borderRadius: 4,
                marginTop: 8,
                background: `linear-gradient(to right, ${token.colorPrimaryBg}, ${token.colorPrimaryBgHover}, ${token.colorPrimaryBorder}, ${token.colorPrimaryBorderHover}, ${token.colorPrimaryHover}, ${token.colorPrimary}, ${token.colorPrimaryActive})`,
              }}
            />
          </Card>
        </Section>

        <Section title="Neutral Text Scale">
          <Card size="small">
            <Flex vertical gap={8}>
              {[
                {
                  key: "colorTextQuaternary",
                  val: token.colorTextQuaternary,
                },
                { key: "colorTextTertiary", val: token.colorTextTertiary },
                { key: "colorTextSecondary", val: token.colorTextSecondary },
                { key: "colorText", val: token.colorText },
              ].map((c) => (
                <TokenRow
                  key={c.key}
                  label={c.key}
                  value={String(c.val)}
                  preview={
                    <Text style={{ color: c.val, fontSize: 14 }}>
                      Sample text
                    </Text>
                  }
                />
              ))}
            </Flex>
          </Card>
        </Section>
      </Flex>
    )
  },
}

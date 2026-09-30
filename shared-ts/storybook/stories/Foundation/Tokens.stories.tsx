import type { Meta, StoryObj } from "@storybook/react-vite";
import { theme } from "antd";
import { Section, Swatch } from "../helpers";

const meta: Meta = {
  title: "Foundation/Tokens",
};

export default meta;

type Story = StoryObj;

function SwatchRow({ entries, bordered }: { entries: [string, string][]; bordered?: boolean }) {
  return (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 12 }}>
      {entries.map(([name, value]) => (
        <Swatch bordered={bordered} key={name} name={name} value={value} />
      ))}
    </div>
  );
}

function TokensPage() {
  const { token } = theme.useToken();

  return (
    <div style={{ maxWidth: 1100 }}>
      <Section
        description="Seeds and their hover/active derivatives as resolved by the active skin."
        title="Brand & status"
      >
        <SwatchRow
          entries={[
            ["colorPrimary", token.colorPrimary],
            ["colorPrimaryHover", token.colorPrimaryHover],
            ["colorPrimaryActive", token.colorPrimaryActive],
            ["colorInfo", token.colorInfo],
            ["colorSuccess", token.colorSuccess],
            ["colorWarning", token.colorWarning],
            ["colorError", token.colorError],
            ["colorLink", token.colorLink],
          ]}
        />
      </Section>

      <Section description="The surface ladder every layer of the UI stands on." title="Surfaces">
        <SwatchRow
          bordered
          entries={[
            ["colorBgBase", token.colorBgBase],
            ["colorBgLayout", token.colorBgLayout],
            ["colorBgContainer", token.colorBgContainer],
            ["colorBgElevated", token.colorBgElevated],
          ]}
        />
      </Section>

      <Section description="Text hierarchy from heading to placeholder." title="Ink scale">
        <SwatchRow
          bordered
          entries={[
            ["colorText", token.colorText],
            ["colorTextSecondary", token.colorTextSecondary],
            ["colorTextTertiary", token.colorTextTertiary],
            ["colorTextQuaternary", token.colorTextQuaternary],
            ["colorTextHeading", token.colorTextHeading],
            ["colorTextLabel", token.colorTextLabel],
            ["colorTextDescription", token.colorTextDescription],
            ["colorTextDisabled", token.colorTextDisabled],
            ["colorTextLightSolid", token.colorTextLightSolid],
          ]}
        />
      </Section>

      <Section description="Hairlines and hover fills." title="Lines & fills">
        <SwatchRow
          bordered
          entries={[
            ["colorBorder", token.colorBorder],
            ["colorBorderSecondary", token.colorBorderSecondary],
            ["colorFill", token.colorFill],
            ["colorFillSecondary", token.colorFillSecondary],
            ["colorFillTertiary", token.colorFillTertiary],
            ["colorFillQuaternary", token.colorFillQuaternary],
          ]}
        />
      </Section>

      <Section description="Radius scale and the three elevation shadows." title="Radius & shadows">
        <div style={{ display: "flex", gap: 12, marginBottom: 16 }}>
          {(
            [
              ["borderRadiusXS", token.borderRadiusXS],
              ["borderRadiusSM", token.borderRadiusSM],
              ["borderRadius", token.borderRadius],
              ["borderRadiusLG", token.borderRadiusLG],
            ] as [string, number][]
          ).map(([name, r]) => (
            <div key={name} style={{ textAlign: "center" }}>
              <div
                style={{
                  background: token.colorFillTertiary,
                  borderRadius: r,
                  height: 44,
                  width: 110,
                }}
              />
              <div style={{ fontSize: 12, marginTop: 4 }}>{name}</div>
              <div style={{ color: token.colorTextTertiary, fontSize: 11 }}>
                {typeof r === "number" ? `${r}px` : r}
              </div>
            </div>
          ))}
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          {(
            [
              ["boxShadowTertiary", token.boxShadowTertiary],
              ["boxShadowSecondary", token.boxShadowSecondary],
              ["boxShadow", token.boxShadow],
            ] as [string, string][]
          ).map(([name, s]) => (
            <div key={name} style={{ textAlign: "center" }}>
              <div
                style={{
                  background: token.colorBgContainer,
                  borderRadius: token.borderRadiusLG,
                  boxShadow: s,
                  height: 44,
                  width: 150,
                }}
              />
              <div style={{ fontSize: 12, marginTop: 8 }}>{name}</div>
            </div>
          ))}
        </div>
      </Section>

      <Section description="Geometry and motion the skin pinned." title="Typography & motion facts">
        <div style={{ fontSize: 12, lineHeight: 2 }}>
          <div>
            fontSize: {token.fontSize}px / lineHeight: {token.lineHeight}
          </div>
          <div>
            fontWeightStrong: {token.fontWeightStrong} / controlHeight: {token.controlHeight}px
          </div>
          <div>
            motionDuration: {token.motionDurationFast} / {token.motionDurationMid} /{" "}
            {token.motionDurationSlow}
          </div>
          <div style={{ fontFamily: token.fontFamily }}>
            fontFamily: {token.fontFamily.slice(0, 72)}
            {token.fontFamily.length > 72 ? "…" : ""}
          </div>
        </div>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <TokensPage />,
};

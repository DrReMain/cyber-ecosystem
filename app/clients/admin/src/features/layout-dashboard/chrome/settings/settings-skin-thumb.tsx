import { theme } from "antd";

export function SettingsSkinThumb() {
  const { token } = theme.useToken();

  return (
    <div
      className="flex overflow-hidden rtl:-scale-x-100"
      style={{
        aspectRatio: "16 / 9",
        background: token.colorBgLayout,
        border: `1px solid ${token.colorBorderSecondary}`,
        borderRadius: token.borderRadiusSM,
      }}
    >
      <div
        className="flex w-1/4 flex-col gap-1 p-1.5"
        style={{ background: token.colorBgContainer }}
      >
        <div
          style={{
            background: token.colorPrimary,
            borderRadius: token.borderRadiusSM,
            height: 5,
            width: "70%",
          }}
        />
        <div
          style={{
            background: token.colorFillSecondary,
            borderRadius: 2,
            height: 4,
            boxShadow: `inset 2px 0 0 ${token.colorPrimary}`,
          }}
        />
        <div style={{ background: token.colorFillSecondary, borderRadius: 2, height: 4 }} />
        <div style={{ background: token.colorFillSecondary, borderRadius: 2, height: 4 }} />
      </div>
      <div className="flex flex-1 flex-col">
        <div
          className="flex items-center gap-1 px-1.5"
          style={{
            background: token.colorBgContainer,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            height: 10,
          }}
        >
          <span className="size-1 rounded-full" style={{ background: token.colorFill }} />
          <span className="size-1 rounded-full" style={{ background: token.colorFill }} />
        </div>
        <div className="flex flex-1 flex-col gap-1 p-1.5">
          <div
            className="flex-1"
            style={{
              background: token.colorBgContainer,
              border: `1px solid ${token.colorBorderSecondary}`,
              borderRadius: token.borderRadiusSM,
            }}
          />
          <div
            style={{
              background: token.colorPrimary,
              borderRadius: token.borderRadiusSM,
              width: 26,
              height: 6,
            }}
          />
        </div>
      </div>
    </div>
  );
}

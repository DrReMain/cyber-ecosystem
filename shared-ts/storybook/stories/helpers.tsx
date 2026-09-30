import { Typography } from "antd";
import type { ReactNode } from "react";

export function Label({ children }: { children: ReactNode }) {
  return (
    <Typography.Text style={{ fontSize: 12 }} type="secondary">
      {children}
    </Typography.Text>
  );
}

export function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <div style={{ marginBottom: 28 }}>
      <Typography.Text
        strong
        style={{
          display: "block",
          fontSize: 13,
          textTransform: "uppercase",
          letterSpacing: "0.05em",
          marginBottom: 4,
        }}
      >
        {title}
      </Typography.Text>
      {description ? (
        <Typography.Text
          type="secondary"
          style={{ display: "block", fontSize: 12, marginBottom: 12 }}
        >
          {description}
        </Typography.Text>
      ) : null}
      {children}
    </div>
  );
}

/** One resolved token as a color chip. The value shown is what the active
 * skin's ConfigProvider actually produced — the specimen proves itself. */
export function Swatch({
  name,
  value,
  bordered,
}: {
  name: string;
  value: string;
  bordered?: boolean;
}) {
  return (
    <div style={{ width: 150 }}>
      <div
        style={{
          background: value,
          border: bordered ? "1px solid rgba(128,128,128,0.35)" : undefined,
          borderRadius: 8,
          height: 44,
        }}
      />
      <div style={{ marginTop: 4, fontSize: 12 }}>{name}</div>
      <Typography.Text type="secondary" style={{ fontSize: 11 }}>
        {value}
      </Typography.Text>
    </div>
  );
}

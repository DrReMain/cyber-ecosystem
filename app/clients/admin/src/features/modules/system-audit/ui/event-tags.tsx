import { Tag, Tooltip } from "antd";
import { m } from "#/paraglide/messages";

function statusColor(status: number): string | undefined {
  if (status >= 500) return "error";
  if (status === 403 || status === 401) return "error";
  if (status >= 400) return "warning";
  if (status >= 200) return "success";
  return undefined;
}

function denyColor(deny: string): string | undefined {
  if (deny === "NO_GRANT") return "purple";
  if (deny === "ABAC_CONSTRAINT") return "cyan";
  return undefined;
}

export function denyLabel(deny: string): string {
  if (deny === "NO_GRANT") return m.system_audit_deny_label_no_grant();
  if (deny === "ABAC_CONSTRAINT") return m.system_audit_deny_label_abac();
  return deny;
}

export function denyHint(deny: string | undefined): string | undefined {
  if (deny === "NO_GRANT") return m.system_audit_deny_no_grant();
  if (deny === "ABAC_CONSTRAINT") return m.system_audit_deny_abac();
  return undefined;
}

export function StatusTag({ status }: Readonly<{ status: number }>) {
  return (
    <Tag className="font-mono" color={statusColor(status)}>
      {status}
    </Tag>
  );
}

export function DenyTag({ deny }: Readonly<{ deny: string | undefined }>) {
  const value = deny ?? "";
  if (value === "") return <span>-</span>;
  const hint = denyHint(value);
  return (
    <Tooltip title={hint === undefined ? value : `${value} · ${hint}`}>
      <Tag color={denyColor(value)}>{denyLabel(value)}</Tag>
    </Tooltip>
  );
}

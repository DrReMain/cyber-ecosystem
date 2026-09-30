import type { AuditLog } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit_pb";
import { dateLikeMs, useColumns } from "@cyber-ecosystem/shared-antd/table";
import { Button, Descriptions, Drawer, Skeleton, Typography } from "antd";
import { toast } from "sonner";
import { UserDeletedTag, userRefView } from "#/features/app/user-ref";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { denyHint, denyLabel, StatusTag } from "./event-tags";
import { Latency } from "./latency";

interface AuditDrawerProps {
  event: AuditLog | null;
  open: boolean;
  actorName: (id: string) => string | undefined;
  actorKnown: boolean;
  afterOpenChange: (open: boolean) => void;
  onClose: () => void;
}

const isoOf = (v: AuditLog["createdAt"]): string => {
  // createdAt may arrive as an RFC3339 string (JSON transport), not a Timestamp.
  const ms = dateLikeMs(v);
  return ms == null ? "" : new Date(ms).toISOString();
};

function eventJson(event: AuditLog): string {
  return JSON.stringify(
    {
      id: event.id ?? "",
      createdAt: isoOf(event.createdAt),
      tenant: event.tenant ?? "",
      actor: event.actor ?? "",
      principalType: event.principalType ?? "",
      operation: event.operation ?? "",
      httpMethod: event.httpMethod ?? "",
      httpPath: event.httpPath ?? "",
      status: Number(event.status ?? 0),
      latencyMs: Number(event.latencyMs ?? 0),
      ip: event.ip ?? "",
      userAgent: event.userAgent ?? "",
      denyReason: event.denyReason ?? "",
    },
    null,
    2,
  );
}

const dash = (v: string | undefined) => (v === undefined || v === "" ? "-" : v);

function ActorCell({
  actor,
  actorName,
  actorKnown,
}: Readonly<{
  actor: string | undefined;
  actorName: (id: string) => string | undefined;
  actorKnown: boolean;
}>) {
  const view = userRefView(actor, actorName, actorKnown);
  if (view.kind === "empty") return m.system_audit_anonymous();
  if (view.kind === "raw") {
    return (
      <span className="flex items-center gap-1.5">
        <span className="font-mono text-[12px]">{view.id}</span>
        {view.deleted ? <UserDeletedTag /> : null}
      </span>
    );
  }
  return (
    <span className="flex flex-col items-start">
      <span>{view.email}</span>
      <Typography.Text className="font-mono text-[12px]" type="secondary">
        {view.id}
      </Typography.Text>
    </span>
  );
}

function DenyCell({ deny }: Readonly<{ deny: string | undefined }>) {
  if (deny === undefined || deny === "") return "-";
  const hint = denyHint(deny);
  return (
    <span className="flex flex-col items-start">
      <span className="flex items-baseline gap-1.5">
        {denyLabel(deny)}
        <span className="font-mono text-[12px] text-black/45 dark:text-white/45">{deny}</span>
      </span>
      {hint !== undefined && (
        <Typography.Text className="text-[12px]" type="secondary">
          {hint}
        </Typography.Text>
      )}
    </span>
  );
}

export function AuditDrawer({
  event,
  open,
  actorName,
  actorKnown,
  afterOpenChange,
  onClose,
}: Readonly<AuditDrawerProps>) {
  const { formatTime } = useColumns();

  return (
    <Drawer
      afterOpenChange={afterOpenChange}
      footer={
        event ? (
          <Button
            block
            color="default"
            onClick={() => {
              navigator.clipboard
                .writeText(eventJson(event))
                .then(() => toast.success(m.system_audit_copied()))
                .catch(() => toast.error(m.system_audit_copy_failed()));
            }}
            variant="filled"
          >
            {m.system_audit_copy_json()}
          </Button>
        ) : null
      }
      onClose={onClose}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      size={560}
    >
      {event === null ? (
        <Skeleton active paragraph={{ rows: 10 }} />
      ) : (
        <Descriptions
          bordered
          column={1}
          size="small"
          styles={{
            label: {
              whiteSpace: "nowrap",
            },
          }}
        >
          <Descriptions.Item label={m.system_audit_field_time()}>
            {formatTime(event.createdAt, { time: "second" })}
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_col_status()}>
            <StatusTag status={Number(event.status ?? 0)} />
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_col_actor()}>
            <ActorCell actor={event.actor} actorKnown={actorKnown} actorName={actorName} />
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_principal()}>
            {dash(event.principalType)}
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_operation()}>
            <span className="break-all font-mono text-[12px]">{event.operation}</span>
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_col_http()}>
            <span className="break-all font-mono text-[12px]">
              {`${event.httpMethod ?? ""} ${event.httpPath ?? ""}`}
            </span>
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_col_latency()}>
            <Latency ms={Number(event.latencyMs ?? 0)} />
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_col_ip()}>
            <span className="font-mono text-[12px]">{dash(event.ip)}</span>
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_deny()}>
            <DenyCell deny={event.denyReason} />
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_tenant()}>
            {dash(event.tenant)}
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_ua()}>
            <span className="break-all font-mono text-[12px]">{dash(event.userAgent)}</span>
          </Descriptions.Item>
          <Descriptions.Item label={m.system_audit_field_id()}>
            <Typography.Text copyable={{ tooltips: false }} style={{ fontFamily: "monospace" }}>
              {event.id}
            </Typography.Text>
          </Descriptions.Item>
        </Descriptions>
      )}
    </Drawer>
  );
}

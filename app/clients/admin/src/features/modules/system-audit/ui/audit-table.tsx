import type { AuditLog } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Button, Tag } from "antd";
import { UserDeletedTag, userRefView } from "#/features/app/user-ref";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { DenyTag, StatusTag } from "./event-tags";
import { Latency } from "./latency";

interface AuditTableProps {
  events: AuditLog[];
  loading: boolean;
  pagination: TablePaginationConfig;
  emptyDescription: string;
  actorName: (id: string) => string | undefined;
  actorKnown: boolean;
  onRefresh: () => void;
  onDetail: (event: AuditLog) => void;
}

export function AuditTable({
  events,
  loading,
  pagination,
  emptyDescription,
  actorName,
  actorKnown,
  onRefresh,
  onDetail,
}: Readonly<AuditTableProps>) {
  const { fieldTimestamp, fieldAction } = useColumns();

  const columns: TableProps<AuditLog>["columns"] = [
    {
      ...fieldTimestamp<AuditLog>("createdAt", {
        title: m.system_audit_col_time(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      title: () => m.system_audit_col_actor(),
      dataIndex: "actor",
      width: 200,
      render: (_, event) => {
        const view = userRefView(event.actor, actorName, actorKnown);
        if (view.kind === "empty") {
          return (
            <span className="text-black/40 dark:text-white/40">{m.system_audit_anonymous()}</span>
          );
        }
        if (view.kind === "raw") {
          return (
            <span className="inline-flex min-w-0 items-center gap-1">
              <span className="truncate font-mono text-[12px]">{view.id}</span>
              {view.deleted ? <UserDeletedTag /> : null}
            </span>
          );
        }
        return <span>{view.email}</span>;
      },
    },
    {
      title: () => m.system_audit_col_operation(),
      dataIndex: "operation",
      render: (_, event) => {
        const sp = event.operation?.split("/");
        const rpc = sp?.pop();
        const namespace = sp?.pop();
        return (
          <span className="inline-flex items-baseline gap-0.5 text-[12px]">
            <span className="text-black/45 dark:text-white/45">{`/${namespace}/`}</span>
            <span className="font-black">{rpc}</span>
          </span>
        );
      },
    },
    {
      title: () => m.system_audit_col_status(),
      dataIndex: "status",
      width: 110,
      render: (_, event) => <StatusTag status={Number(event.status ?? 0)} />,
    },
    {
      title: () => m.system_audit_col_deny(),
      dataIndex: "denyReason",
      width: 190,
      render: (_, event) => <DenyTag deny={event.denyReason} />,
    },
    {
      title: () => m.system_audit_col_http(),
      dataIndex: "httpPath",
      render: (_, event) => {
        const method = event.httpMethod ?? "";
        const path = event.httpPath ?? "";
        if (method === "") {
          return <span className="font-mono text-[12px]">{path}</span>;
        }
        return (
          <span className="inline-flex min-w-0 items-center gap-1.5">
            <Tag className="m-0 font-mono" color={method === "DELETE" ? "error" : undefined}>
              {method}
            </Tag>
            <span className="truncate font-mono text-[12px]">{path}</span>
          </span>
        );
      },
    },
    {
      align: getTextDirection() === "rtl" ? "left" : "right",
      title: () => m.system_audit_col_latency(),
      dataIndex: "latencyMs",
      width: 110,
      render: (_, event) => <Latency ms={Number(event.latencyMs ?? 0)} />,
    },
    {
      title: () => m.system_audit_col_ip(),
      dataIndex: "ip",
      width: 150,
      render: (_, event) => (
        <span className="font-mono text-[12px]">{event.ip === "" ? "-" : event.ip}</span>
      ),
    },
    fieldAction<AuditLog>(
      (_, event) => (
        <Button color="primary" onClick={() => onDetail(event)} size="small" variant="text">
          {m.system_audit_act_detail()}
        </Button>
      ),
      { title: m.system_audit_col_action() },
    ),
  ];

  return (
    <DataTable<AuditLog>
      columnSettings={{
        storageKey: "cyber.columns.system-audit",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={events}
      emptyDescription={emptyDescription}
      loading={loading}
      pagination={pagination}
      rowClassName={(event) =>
        (event.denyReason ?? "") !== "" ? "bg-red-500/5 dark:bg-red-400/10" : ""
      }
      rowKey="id"
      toolbar={{
        labels: {
          density: m.common_toolbar_density(),
          densityLarge: m.common_toolbar_density_large(),
          densityMiddle: m.common_toolbar_density_middle(),
          densitySmall: m.common_toolbar_density_small(),
          refresh: m.common_toolbar_refresh(),
        },
        loading,
        onRefresh,
      }}
    />
  );
}

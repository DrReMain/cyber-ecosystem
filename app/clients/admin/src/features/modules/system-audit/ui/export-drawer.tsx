import { useQuery } from "@connectrpc/connect-query";
import type { AuditExport } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit_pb";
import { listAuditExports } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit-AuditService_connectquery";
import { useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Button, Drawer, Table, Tag, Typography } from "antd";
import { useMemo } from "react";
import { useFileUrls } from "#/features/app/file/use-file-url";
import { UserDeletedTag } from "#/features/app/user/user-deleted-tag";
import { userRefView } from "#/features/app/user/user-ref";
import { formatSize } from "#/libs";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

type ExportStatus = "processing" | "confirmed" | "failed" | "deleted";

const POLL_MS = 2000;

function exportStatus(e: AuditExport): ExportStatus {
  if (!e.file) return "deleted";
  switch (String(e.file.status)) {
    case "FILE_STATUS_PROCESSING":
      return "processing";
    case "FILE_STATUS_FAILED":
      return "failed";
    default:
      return "confirmed";
  }
}

function hasProcessing(list: readonly AuditExport[] | undefined): boolean {
  return (list ?? []).some((e) => exportStatus(e) === "processing");
}

interface ExportDrawerProps {
  open: boolean;
  ownerName: (id: string) => string | undefined;
  ownerKnown: boolean;
  onClose: () => void;
}

export function ExportDrawer({
  open,
  ownerName,
  ownerKnown,
  onClose,
}: Readonly<ExportDrawerProps>) {
  const { fieldTimestamp, fixed } = useColumns();
  const { data, isFetching } = useQuery(
    listAuditExports,
    { orderBy: ["createdAt:desc"], page: { all: true } },
    {
      enabled: open,
      refetchInterval: (query) => (hasProcessing(query.state.data?.list) ? POLL_MS : false),
    },
  );
  const rows = useMemo(() => data?.list ?? [], [data]);
  const urls = useFileUrls(
    open ? rows.filter((e) => exportStatus(e) === "confirmed").map((e) => e.fileId) : [],
  );

  const columns: TableProps<AuditExport>["columns"] = [
    {
      ...fieldTimestamp<AuditExport>("createdAt", {
        title: m.system_audit_export_col_time(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      title: () => m.system_audit_export_col_owner(),
      dataIndex: "ownerId",
      width: 200,
      render: (_, e) => {
        const view = userRefView(e.ownerId, ownerName, ownerKnown);
        if (view.kind === "raw") {
          return (
            <span className="inline-flex min-w-0 items-center gap-1">
              <span className="font-mono text-[12px]">{view.id}</span>
              {view.deleted ? <UserDeletedTag /> : null}
            </span>
          );
        }
        if (view.kind === "known") {
          return <span>{view.email}</span>;
        }
        return null;
      },
    },
    {
      title: () => m.system_audit_export_col_file(),
      render: (_, e) =>
        e.file ? (
          <span className="flex flex-col items-start">
            <span className="max-w-80 truncate font-medium">{e.file.name}</span>
            <Typography.Text className="font-mono text-[12px]" type="secondary">
              {formatSize(Number(e.file.size ?? 0))}
            </Typography.Text>
          </span>
        ) : (
          <span className="text-black/40 dark:text-white/40">
            {m.system_audit_export_file_deleted()}
          </span>
        ),
    },
    {
      title: () => m.system_audit_export_col_status(),
      render: (_, e) => {
        switch (exportStatus(e)) {
          case "processing":
            return <Tag color="processing">{m.system_audit_export_status_processing()}</Tag>;
          case "failed":
            return <Tag color="error">{m.system_audit_export_status_failed()}</Tag>;
          case "deleted":
            return <Tag>{m.system_audit_export_file_deleted()}</Tag>;
          default:
            return <Tag color="success">{m.system_audit_export_status_confirmed()}</Tag>;
        }
      },
    },
    {
      key: "download",
      title: () => m.system_audit_export_col_action(),
      fixed: fixed("right"),
      render: (_, e) => {
        if (exportStatus(e) !== "confirmed") return null;
        const url = urls.get(e.fileId);
        return (
          <Button
            color="primary"
            disabled={!url}
            href={url}
            size="small"
            target="_blank"
            variant="text"
          >
            {m.system_audit_act_download()}
          </Button>
        );
      },
    },
  ];

  return (
    <Drawer
      onClose={onClose}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      size={560}
      title={m.system_audit_export_title()}
    >
      <Table<AuditExport>
        columns={columns}
        dataSource={rows}
        loading={isFetching}
        locale={{ emptyText: m.system_audit_export_empty() }}
        pagination={false}
        rowKey={(e) => e.fileId}
        scroll={{ x: "max-content" }}
        size="small"
      />
    </Drawer>
  );
}

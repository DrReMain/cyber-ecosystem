import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Button, Popconfirm, Progress, Tag, Tooltip } from "antd";
import { RotateCw, X } from "lucide-react";
import type { ReactNode } from "react";
import { isPreviewable } from "#/features/app/file/is-previewable";
import { useFileUrls } from "#/features/app/file/use-file-url";
import type { UploadProgress } from "#/features/app/file/use-upload";
import { UserDeletedTag } from "#/features/app/user/user-deleted-tag";
import { userRefView } from "#/features/app/user/user-ref";
import { formatSize } from "#/libs";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { UploadProbe } from "./upload-probe";

const enumName = (v: unknown): string => (v == null ? "" : String(v));

type FileStatusView = "uploading" | "confirmed" | "processing" | "failed";

function fileStatus(file: FileView): FileStatusView {
  switch (enumName(file.status)) {
    case "FILE_STATUS_CONFIRMED":
      return "confirmed";
    case "FILE_STATUS_PROCESSING":
      return "processing";
    case "FILE_STATUS_FAILED":
      return "failed";
    default:
      return "uploading";
  }
}

function statusTag(status: FileStatusView) {
  switch (status) {
    case "confirmed":
      return <Tag color="success">{m.files_status_confirmed()}</Tag>;
    case "processing":
      return <Tag color="processing">{m.files_status_processing()}</Tag>;
    case "failed":
      return <Tag color="error">{m.files_status_failed()}</Tag>;
    default:
      return <Tag>{m.files_status_uploading()}</Tag>;
  }
}

const SOURCE_COLORS: Record<string, string> = {
  FILE_SOURCE_CLIENT_UPLOAD: "blue",
  FILE_SOURCE_SERVER_GENERATED: "purple",
};

function sourceLabel(source: string): string {
  if (source === "FILE_SOURCE_SERVER_GENERATED") {
    return m.files_source_server();
  }
  return m.files_source_client();
}

interface FilesTableProps {
  files: FileView[];
  loading: boolean;
  pagination: TablePaginationConfig;
  emptyDescription: string;
  ownerName: (id: string) => string | undefined;
  ownerKnown: boolean;
  pendingDeleteId: string | null;
  pendingAbortId: string | null;
  progressById: Record<string, UploadProgress>;
  toolbarExtra: ReactNode;
  onRename: (file: FileView) => void;
  onDelete: (file: FileView) => void;
  onResume: (file: FileView) => void;
  onAbort: (file: FileView) => void;
  onRefresh: () => void;
}

export function FilesTable({
  files,
  loading,
  pagination,
  emptyDescription,
  ownerName,
  ownerKnown,
  pendingDeleteId,
  pendingAbortId,
  progressById,
  toolbarExtra,
  onRename,
  onDelete,
  onResume,
  onAbort,
  onRefresh,
}: Readonly<FilesTableProps>) {
  const { fieldTimestamp, fieldAction } = useColumns();
  const urls = useFileUrls(
    files
      .filter((f) => fileStatus(f) === "confirmed")
      .map((f) => f.id ?? "")
      .filter(Boolean),
  );

  const renderCancel = (file: FileView) => (
    <Popconfirm
      cancelButtonProps={{ variant: "filled", color: "default" }}
      okButtonProps={{ variant: "filled", color: "danger" }}
      onConfirm={() => onAbort(file)}
      title={m.files_abort_confirm()}
    >
      <Tooltip title={m.common_cancel()}>
        <span className="inline-flex">
          <Button
            aria-label={m.common_cancel()}
            color="danger"
            disabled={pendingAbortId === (file.id ?? "")}
            icon={<X size={14} />}
            size="small"
            variant="text"
          />
        </span>
      </Tooltip>
    </Popconfirm>
  );

  const columns: TableProps<FileView>["columns"] = [
    {
      title: () => m.files_col_name(),
      dataIndex: "name",
      render: (_, file) => <span className="font-medium">{file.name}</span>,
    },
    {
      title: () => m.files_col_type(),
      dataIndex: "contentType",
      width: 170,
      render: (_, file) => <Tag>{file.contentType || "-"}</Tag>,
    },
    {
      title: () => m.files_col_size(),
      dataIndex: "size",
      width: 110,
      align: getTextDirection() === "rtl" ? "left" : "right",
      render: (_, file) => formatSize(Number(file.size ?? 0)),
    },
    {
      title: () => m.files_col_source(),
      dataIndex: "source",
      width: 130,
      render: (_, file) => (
        <Tag color={SOURCE_COLORS[enumName(file.source)] ?? "default"}>
          {sourceLabel(enumName(file.source))}
        </Tag>
      ),
    },
    {
      title: () => m.files_col_status(),
      dataIndex: "status",
      width: 110,
      render: (_, file) => statusTag(fileStatus(file)),
    },
    {
      key: "progress",
      title: () => m.files_col_progress(),
      width: 220,
      render: (_, file) => {
        const status = fileStatus(file);
        if (status === "confirmed") {
          return <Progress percent={100} size="small" status="success" />;
        }
        if (status === "processing") {
          return <Progress percent={100} showInfo={false} size="small" status="active" />;
        }
        if (status === "failed") {
          return <Progress percent={100} showInfo={false} size="small" status="exception" />;
        }
        const p = progressById[file.id ?? ""];
        const pct = p && p.total > 0 ? Math.min(100, Math.floor((p.loaded / p.total) * 100)) : 0;
        return (
          <div className="flex items-center gap-2">
            {p ? <Progress percent={pct} size="small" /> : <UploadProbe file={file} />}
            <Tooltip title={m.files_act_resume()}>
              <span className="inline-flex">
                <Button
                  aria-label={m.files_act_resume()}
                  color="primary"
                  disabled={Boolean(p)}
                  icon={<RotateCw size={14} />}
                  onClick={() => onResume(file)}
                  size="small"
                  variant="text"
                />
              </span>
            </Tooltip>
            {renderCancel(file)}
          </div>
        );
      },
    },
    {
      title: () => m.files_col_owner(),
      dataIndex: "ownerId",
      width: 200,
      render: (_, file) => {
        const view = userRefView(file.ownerId, ownerName, ownerKnown);
        if (view.kind === "known") {
          return <span>{view.email}</span>;
        }
        if (view.kind === "raw") {
          return (
            <span className="inline-flex min-w-0 items-center gap-1">
              <span
                className={
                  view.deleted ? "font-mono text-[12px]" : "font-mono text-[12px] text-ink-tertiary"
                }
              >
                {view.id}
              </span>
              {view.deleted ? <UserDeletedTag /> : null}
            </span>
          );
        }
        return <span className="text-ink-tertiary">-</span>;
      },
    },
    {
      ...fieldTimestamp<FileView>("createdAt", {
        title: m.files_col_created(),
        format: { time: "second" },
      }),
      width: 180,
    },
    fieldAction<FileView>(
      (_, file) => {
        const status = fileStatus(file);
        const confirmed = status === "confirmed";
        const deletable = confirmed || status === "failed";
        const url = confirmed ? urls.get(file.id ?? "") : undefined;
        return (
          <div className="flex items-center gap-1">
            <Button
              color="primary"
              disabled={!url}
              href={url}
              rel="noreferrer"
              size="small"
              target="_blank"
              variant="text"
            >
              {isPreviewable(file.contentType ?? "")
                ? m.files_act_preview()
                : m.files_act_preview()}
            </Button>
            <Button color="primary" onClick={() => onRename(file)} size="small" variant="text">
              {m.files_act_rename()}
            </Button>
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => onDelete(file)}
              placement={getTextDirection() === "rtl" ? "right" : "left"}
              title={m.files_delete_confirm()}
            >
              <Button
                color="danger"
                disabled={!deletable || pendingDeleteId === (file.id ?? "")}
                size="small"
                variant="text"
              >
                {m.files_act_delete()}
              </Button>
            </Popconfirm>
          </div>
        );
      },
      { title: m.files_col_actions() },
    ),
  ];

  return (
    <DataTable<FileView>
      columnSettings={{
        storageKey: "cyber.columns.files",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={files}
      emptyDescription={emptyDescription}
      loading={loading}
      pagination={pagination}
      rowKey="id"
      toolbar={{
        extra: toolbarExtra,
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

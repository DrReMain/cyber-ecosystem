import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Progress, Tag } from "antd";
import { Eye, Pencil, RotateCw, Trash, X } from "lucide-react";
import type { ReactNode } from "react";
import { ActionIconButton } from "#/components/action-icon-button";
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
  meId: string;
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
  meId,
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
    <ActionIconButton
      confirm={{ onConfirm: () => onAbort(file), title: m.files_abort_confirm() }}
      disabled={pendingAbortId === (file.id ?? "")}
      icon={X}
      label={m.common_cancel()}
    />
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
        if (file.ownerId !== meId) {
          return null;
        }
        const p = progressById[file.id ?? ""];
        const pct = p && p.total > 0 ? Math.min(100, Math.floor((p.loaded / p.total) * 100)) : 0;
        return (
          <div className="flex items-center gap-2">
            {p ? <Progress percent={pct} size="small" /> : <UploadProbe file={file} />}
            <ActionIconButton
              disabled={Boolean(p)}
              icon={RotateCw}
              label={m.files_act_resume()}
              onClick={() => onResume(file)}
            />
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
          return <span>{view.account}</span>;
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
            <ActionIconButton
              disabled={!url}
              href={url}
              icon={Eye}
              label={m.files_act_preview()}
              rel="noreferrer"
              target="_blank"
            />
            <ActionIconButton
              icon={Pencil}
              label={m.files_act_rename()}
              onClick={() => onRename(file)}
            />
            <ActionIconButton
              confirm={{ onConfirm: () => onDelete(file), title: m.files_delete_confirm() }}
              disabled={!deletable || pendingDeleteId === (file.id ?? "")}
              icon={Trash}
              label={m.files_act_delete()}
            />
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

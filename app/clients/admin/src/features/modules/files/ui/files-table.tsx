import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Button, Popconfirm, Tag } from "antd";
import { Pencil, Trash2 } from "lucide-react";
import type { ReactNode } from "react";
import { UserDeletedTag, userRefView } from "#/features/app/user-ref";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

const STATUS_CONFIRMED = "FILE_STATUS_CONFIRMED";

const enumName = (v: unknown): string => (v == null ? "" : String(v));

const SOURCE_COLORS: Record<string, string> = {
  FILE_SOURCE_CLIENT_UPLOAD: "blue",
  FILE_SOURCE_SERVER_GENERATED: "purple",
};

function formatSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  const units = ["KB", "MB", "GB", "TB"];
  let value = bytes;
  let unit = -1;
  do {
    value /= 1024;
    unit += 1;
  } while (value >= 1024 && unit < units.length - 1);
  return `${value < 10 ? value.toFixed(1) : Math.round(value)} ${units[unit]}`;
}

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
  toolbarExtra: ReactNode;
  onRename: (file: FileView) => void;
  onDelete: (file: FileView) => void;
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
  toolbarExtra,
  onRename,
  onDelete,
  onRefresh,
}: Readonly<FilesTableProps>) {
  const { fieldTimestamp, fieldAction } = useColumns();

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
      render: (_, file) =>
        enumName(file.status) === STATUS_CONFIRMED ? (
          <Tag color="success">{m.files_status_confirmed()}</Tag>
        ) : (
          <Tag>{m.files_status_uploading()}</Tag>
        ),
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
                  view.deleted
                    ? "font-mono text-[12px]"
                    : "font-mono text-[12px] text-black/40 dark:text-white/40"
                }
              >
                {view.id}
              </span>
              {view.deleted ? <UserDeletedTag /> : null}
            </span>
          );
        }
        return <span className="text-black/40 dark:text-white/40">-</span>;
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
        const renameBtn = (
          <Button
            color="primary"
            icon={<Pencil size={14} />}
            onClick={() => onRename(file)}
            size="small"
            variant="text"
          >
            {m.files_act_rename()}
          </Button>
        );
        // An uploading row is the upload channel's in-flight state — delete
        // refuses it server-side, so only rename is offered here.
        if (enumName(file.status) !== STATUS_CONFIRMED) {
          return renameBtn;
        }
        return (
          <div className="flex items-center gap-1">
            {renameBtn}
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => onDelete(file)}
              placement={getTextDirection() === "rtl" ? "right" : "left"}
              title={m.files_delete_confirm()}
            >
              <Button
                color="danger"
                disabled={pendingDeleteId === (file.id ?? "")}
                icon={<Trash2 size={14} />}
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

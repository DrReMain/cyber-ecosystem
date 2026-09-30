import type { CardProps, TableProps } from "antd";
import { Button, Card, Empty, Flex, Space, Table, Typography } from "antd";
import type { ComponentProps, ReactNode, Ref } from "react";
import { useMemo } from "react";
import type { ColumnSettingsConfig, ColumnSettingsEntry } from "./column-settings";
import { ColumnSettingsPopover, useColumnVisibility } from "./column-settings";
import type { TableToolbarLabels } from "./table-toolbar";
import { TableToolbar } from "./table-toolbar";
import { useTable } from "./use-table";

// DataTable is the list-page paradigm component: Card + toolbar + Table as
// one coherent entity. Opinionated defaults (density follows the global
// preference, no pagination, horizontal max-content scroll) with the full
// antd Table surface left open via spread — rowSelection, expandable,
// virtual, sticky, summary, loading, onChange all pass straight through.
// Wiring points for capabilities that are business decisions, not defaults:
//   - toolbar.tools        — right-cluster slot (export and friends)
//   - toolbar.extra        — left slot (create button, counters…)
//   - selection.actions    — batch-action buttons shown while rows are
//                            selected (the bar itself renders on any
//                            non-empty controlled rowSelection)
//   - columnSettings       — per-table visibility dropdown over column ids
//                            (key ?? dataIndex, arrays joined with "."),
//                            persisted as the hidden list so new columns
//                            default visible
// Text stays in the host: toolbar labels, selection copy, column-settings
// copy and empty copy all arrive as props, never imported here.

export interface DataTableToolbar {
  extra?: ReactNode;
  labels?: TableToolbarLabels;
  loading?: boolean;
  onRefresh?: () => void;
  tools?: ReactNode;
}

export interface DataTableSelection {
  actions?: ReactNode;
  labels?: {
    selected?: (count: number) => string;
    clear?: string;
  };
}

export interface DataTableProps<T extends object> extends Omit<TableProps<T>, "size"> {
  cardProps?: CardProps;
  columnSettings?: ColumnSettingsConfig;
  emptyAction?: ReactNode;
  emptyDescription?: ReactNode;
  ref?: Ref<ComponentProps<typeof Table>["ref"] extends Ref<infer R> ? R : never>;
  selection?: DataTableSelection;
  toolbar?: DataTableToolbar;
}

// Column identity for settings: React keys may be numbers too. Nested-path
// dataIndex arrays join with "." so array and dotted-string forms of the same
// path share one stable storage id.
function columnId(column: { key?: unknown; dataIndex?: unknown }): string | undefined {
  if (typeof column.key === "string" || typeof column.key === "number") {
    return String(column.key);
  }
  if (typeof column.dataIndex === "string") return column.dataIndex;
  if (Array.isArray(column.dataIndex) && column.dataIndex.length > 0) {
    return column.dataIndex.map(String).join(".");
  }
  return undefined;
}

function columnTitle(column: { title?: unknown }): ReactNode {
  if (typeof column.title === "function") {
    return (column.title as (props: unknown) => ReactNode)({});
  }
  return (column.title ?? null) as ReactNode;
}

export function DataTable<T extends object>({
  cardProps,
  columnSettings,
  emptyAction,
  emptyDescription,
  ref,
  rowSelection,
  selection,
  toolbar,
  ...table
}: DataTableProps<T>) {
  const { tableSize, setTableSize } = useTable();
  const columns = table.columns ?? [];
  const ids = useMemo(
    () => columns.map((c) => columnId(c)).filter((id): id is string => id !== undefined),
    [columns],
  );
  const visibleIds = useColumnVisibility(columnSettings?.storageKey, ids);
  const settingsEntries = useMemo<ColumnSettingsEntry[]>(() => {
    if (!columnSettings) return [];
    return columns
      .map((c) => {
        const id = columnId(c);
        return id === undefined
          ? null
          : { id, title: columnTitle(c), locked: columnSettings.locked?.includes(id) ?? false };
      })
      .filter((e): e is ColumnSettingsEntry => e !== null);
  }, [columns, columnSettings]);
  const effectiveColumns = useMemo(
    () =>
      columns.map((c) => {
        const id = columnId(c);
        return id !== undefined && !visibleIds.includes(id) ? { ...c, hidden: true } : c;
      }),
    [columns, visibleIds],
  );
  const hasEmpty = emptyAction !== undefined || emptyDescription !== undefined;
  const locale = hasEmpty
    ? {
        ...table.locale,
        emptyText: (
          <Empty description={emptyDescription} image={Empty.PRESENTED_IMAGE_SIMPLE}>
            {emptyAction}
          </Empty>
        ),
      }
    : table.locale;
  const selectedCount = rowSelection?.selectedRowKeys?.length ?? 0;
  return (
    <Card {...cardProps}>
      <Flex gap={16} vertical>
        <TableToolbar
          extra={toolbar?.extra}
          labels={toolbar?.labels}
          loading={toolbar?.loading}
          onRefresh={toolbar?.onRefresh}
          onSizeChange={setTableSize}
          size={tableSize}
          tools={
            columnSettings ? (
              <>
                <ColumnSettingsPopover
                  config={columnSettings}
                  entries={settingsEntries}
                  visibleIds={visibleIds}
                />
                {toolbar?.tools}
              </>
            ) : (
              toolbar?.tools
            )
          }
        />
        {rowSelection && selectedCount > 0 && (
          <Flex align="center" gap={12} justify="space-between">
            <Space size={12}>
              <Typography.Text type="secondary">
                {(selection?.labels?.selected ?? ((n) => `Selected ${n}`))(selectedCount)}
              </Typography.Text>
              <Button
                onClick={() => rowSelection.onChange?.([], [], { type: "none" })}
                size="small"
                type="link"
              >
                {selection?.labels?.clear ?? "Clear"}
              </Button>
            </Space>
            <Space>{selection?.actions}</Space>
          </Flex>
        )}
        {/* Spread first, defaults after: host props always win over the
            no-ops, while size stays owned by the global density. */}
        <Table<T>
          {...table}
          columns={effectiveColumns}
          locale={locale}
          pagination={table.pagination ?? false}
          ref={ref}
          rowSelection={rowSelection}
          scroll={table.scroll ?? { x: "max-content" }}
          size={tableSize}
        />
      </Flex>
    </Card>
  );
}

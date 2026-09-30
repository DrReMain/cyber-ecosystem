import { Button, Dropdown, Flex, Space, Tooltip } from "antd";
import { RefreshCw, Rows3 } from "lucide-react";
import type { ReactNode } from "react";
import type { TableSize } from "./use-table";

export interface TableToolbarLabels {
  refresh?: string;
  density?: string;
  densityLarge?: string;
  densityMiddle?: string;
  densitySmall?: string;
}

export interface TableToolbarProps {
  labels?: TableToolbarLabels;
  extra?: ReactNode;
  size?: TableSize;
  onSizeChange?: (size: TableSize) => void;
  onRefresh?: () => void;
  loading?: boolean;
  tools?: ReactNode;
}

export function TableToolbar({
  labels,
  extra,
  size,
  onSizeChange,
  onRefresh,
  loading,
  tools,
}: TableToolbarProps) {
  const {
    refresh: refreshLabel = "Refresh",
    density: densityLabel = "Density",
    densityLarge: largeLabel = "Large",
    densityMiddle: middleLabel = "Default",
    densitySmall: smallLabel = "Compact",
  } = labels ?? {};

  return (
    <Flex justify="space-between">
      {extra ?? <span />}
      <Space>
        {tools}
        {size && onSizeChange && (
          <Dropdown
            trigger={["click"]}
            menu={{
              items: [
                { key: "large", label: largeLabel },
                { key: "middle", label: middleLabel },
                { key: "small", label: smallLabel },
              ],
              selectedKeys: [size],
              onClick: ({ key }) => onSizeChange(key as TableSize),
            }}
          >
            <Tooltip title={densityLabel}>
              <Button
                aria-label={densityLabel}
                color="default"
                icon={<Rows3 size={14} />}
                variant="filled"
              />
            </Tooltip>
          </Dropdown>
        )}
        {onRefresh && (
          <Tooltip title={refreshLabel}>
            <span>
              <Button
                aria-label={refreshLabel}
                color="default"
                disabled={loading}
                icon={<RefreshCw size={14} />}
                onClick={onRefresh}
                variant="filled"
              />
            </span>
          </Tooltip>
        )}
      </Space>
    </Flex>
  );
}

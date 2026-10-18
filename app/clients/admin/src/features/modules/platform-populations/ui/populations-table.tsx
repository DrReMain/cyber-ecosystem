import type { Population } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/population_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Avatar, Space, Tag, Typography } from "antd";
import { avatarColor } from "#/libs";
import { m } from "#/paraglide/messages";

// The compile-time archetype vocabulary rendered as its two labels.
function archetypeTag(archetype: string | undefined) {
  if (archetype === "engine")
    return <Tag color="processing">{m.platform_populations_arch_engine()}</Tag>;
  if (archetype === "patron")
    return <Tag color="purple">{m.platform_populations_arch_patron()}</Tag>;
  return <Tag>{archetype}</Tag>;
}

interface PopulationsTableProps {
  populations: Population[];
  loading?: boolean;
  onRefresh: () => void;
}

export function PopulationsTable({
  populations,
  loading,
  onRefresh,
}: Readonly<PopulationsTableProps>) {
  const { fieldTimestamp, fixed } = useColumns();
  const columns: TableProps<Population>["columns"] = [
    {
      title: () => m.platform_populations_col_population(),
      dataIndex: "name",
      width: 240,
      fixed: fixed("left"),
      render: (_, population) => (
        <Space>
          <Avatar style={{ backgroundColor: avatarColor(population.code ?? ""), flexShrink: 0 }}>
            {population.name?.charAt(0)}
          </Avatar>
          <div className="flex flex-col">
            <span>{population.name}</span>
            <Typography.Text className="font-mono" type="secondary">
              {population.code}
            </Typography.Text>
          </div>
        </Space>
      ),
    },
    {
      align: "center",
      title: () => m.platform_populations_col_archetype(),
      width: 140,
      render: (_, population) => archetypeTag(population.archetype),
    },
    fieldTimestamp<Population>("createdAt", { title: m.platform_populations_col_created_at() }),
    fieldTimestamp<Population>("updatedAt", { title: m.platform_populations_col_updated_at() }),
  ];
  return (
    <DataTable<Population>
      columnSettings={{
        storageKey: "cyber.columns.platform-populations",
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={populations}
      emptyDescription={m.platform_populations_empty()}
      loading={loading}
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

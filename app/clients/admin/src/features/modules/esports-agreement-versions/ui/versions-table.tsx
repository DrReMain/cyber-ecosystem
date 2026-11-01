import type { AgreementVersion } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Tag } from "antd";
import { FileText, ListChecks, Star, Trash2 } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { fullDate } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";
import { docOptions } from "../vocab";

interface VersionsTableProps {
  rows: AgreementVersion[];
  loading?: boolean;
  onDelete: (row: AgreementVersion) => void;
  onRefresh: () => void;
  onSetCurrent: (row: AgreementVersion) => void;
  onView: (row: AgreementVersion) => void;
}

export function VersionsTable({
  rows,
  loading,
  onDelete,
  onRefresh,
  onSetCurrent,
  onView,
}: Readonly<VersionsTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<AgreementVersion>["columns"] = [
    {
      title: () => m.esports_agreement_versions_col_doc(),
      dataIndex: "doc",
      width: 110,
      render: (_, row) => (
        <Tag
          className="m-0"
          color={String(row.doc) === "AGREEMENT_DOC_PRIVACY" ? "purple" : "blue"}
        >
          {docOptions.find((o) => o.value === String(row.doc))?.label() ?? String(row.doc)}
        </Tag>
      ),
    },
    {
      title: () => m.esports_agreement_versions_col_no(),
      dataIndex: "no",
      width: 110,
      render: (_, row) => <span className="font-mono">{row.no}</span>,
    },
    {
      title: () => m.esports_agreement_versions_col_title(),
      dataIndex: "title",
      width: 220,
    },
    {
      title: () => m.esports_agreement_versions_col_effective(),
      dataIndex: "effectiveAt",
      width: 130,
      render: (_, row) => (row.effectiveAt ? fullDate(row.effectiveAt) : ""),
    },
    {
      title: () => m.esports_agreement_versions_col_agreed(),
      dataIndex: "agreedCount",
      width: 110,
      render: (_, row) => <span className="tabular-nums">{Number(row.agreedCount ?? 0)}</span>,
    },
    {
      title: () => m.esports_agreement_versions_col_state(),
      dataIndex: "isCurrent",
      width: 100,
      render: (_, row) =>
        row.isCurrent ? (
          <Tag className="m-0" color="success">
            {m.esports_agreement_versions_state_current()}
          </Tag>
        ) : (
          <Tag className="m-0">{m.esports_agreement_versions_state_history()}</Tag>
        ),
    },
    {
      ...fieldAction<AgreementVersion>(
        (_, row) => {
          // Only a historical version can be promoted or deleted — the
          // current one anchors the login gate and must be replaced, not
          // removed; the server refuses the same pair, the table never
          // offers them.
          return (
            <div className="flex gap-1">
              <ActionIconButton
                icon={FileText}
                label={m.esports_agreement_versions_act_view()}
                onClick={() => onView(row)}
              />
              {!row.isCurrent && (
                <ActionIconButton
                  confirm={{
                    description: m.esports_agreement_versions_current_audit(),
                    onConfirm: () => onSetCurrent(row),
                    title: m.esports_agreement_versions_current_confirm({ no: row.no ?? "" }),
                  }}
                  icon={Star}
                  label={m.esports_agreement_versions_act_set_current()}
                />
              )}
              <ActionIconButton
                href={`/dashboard/compliance/agreement-consents?doc=${String(row.doc)}&ver=${encodeURIComponent(row.no ?? "")}&pageNo=1&pageSize=20`}
                icon={ListChecks}
                label={m.esports_agreement_versions_act_consents()}
              />
              {!row.isCurrent && (
                <ActionIconButton
                  color="danger"
                  confirm={{
                    description: m.esports_agreement_versions_delete_audit(),
                    onConfirm: () => onDelete(row),
                    title: m.esports_agreement_versions_delete_confirm({ no: row.no ?? "" }),
                  }}
                  icon={Trash2}
                  label={m.esports_agreement_versions_act_delete()}
                />
              )}
            </div>
          );
        },
        { title: m.esports_agreement_versions_col_actions() },
      ),
      width: 150,
    },
  ];

  return (
    <DataTable<AgreementVersion>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_agreement_versions_empty()}
      loading={loading}
      pagination={false}
      rowKey={(r) => r.id ?? ""}
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

import type { AgreementConsent } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Tag } from "antd";
import { m } from "#/paraglide/messages";
import { docOptions, loginMethodLabel } from "../../esports-agreement-versions/vocab";

interface ConsentsTableProps {
  rows: AgreementConsent[];
  loading?: boolean;
  pagination: TablePaginationConfig;
  onRefresh: () => void;
}

export function ConsentsTable({
  rows,
  loading,
  pagination,
  onRefresh,
}: Readonly<ConsentsTableProps>) {
  const { fieldTimestamp } = useColumns();

  const columns: TableProps<AgreementConsent>["columns"] = [
    {
      title: () => m.esports_agreement_consents_col_user(),
      dataIndex: "nickname",
      width: 150,
      render: (_, row) => row.nickname ?? row.userId ?? "",
    },
    {
      title: () => m.esports_agreement_consents_col_phone(),
      dataIndex: "phoneMasked",
      width: 140,
      render: (_, row) => <span className="tabular-nums">{row.phoneMasked ?? "—"}</span>,
    },
    {
      title: () => m.esports_agreement_consents_col_doc(),
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
      title: () => m.esports_agreement_consents_col_ver(),
      dataIndex: "versionNo",
      width: 110,
      render: (_, row) => <span className="font-mono">{row.versionNo}</span>,
    },
    {
      ...fieldTimestamp<AgreementConsent>("consentedAt", {
        title: m.esports_agreement_consents_col_time(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      title: () => m.esports_agreement_consents_col_method(),
      dataIndex: "loginMethod",
      width: 130,
      render: (_, row) => loginMethodLabel(String(row.loginMethod)),
    },
  ];

  return (
    <DataTable<AgreementConsent>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_agreement_consents_empty()}
      loading={loading}
      pagination={pagination}
      rowKey={(r) => `${r.userId}-${r.doc}-${r.consentedAt?.seconds ?? ""}-${r.loginMethod}`}
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

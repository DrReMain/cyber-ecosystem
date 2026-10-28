import {
  ReportAccessPolicy,
  type ReportPolicyRow,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { ReportType } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/entitlement_pb";

import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Modal, Select, Tag, Typography } from "antd";
import { KeyRound } from "lucide-react";
import { useEffect, useState } from "react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";

export interface PolicyRowKey {
  reportType: string;
  current: string;
}

interface PoliciesTableProps {
  rows: ReportPolicyRow[];
  loading?: boolean;
  pending?: boolean;
  setting: PolicyRowKey | null;
  onRefresh: () => void;
  onSet: (row: PolicyRowKey | null) => void;
  onSubmit: (reportType: ReportType, policy: ReportAccessPolicy) => void;
}

const familyLabels: Record<string, () => string> = {
  REPORT_TYPE_COGNITIVE: () => m.esports_report_type_cognitive(),
  REPORT_TYPE_CAREER: () => m.esports_report_type_career(),
  REPORT_TYPE_SCHOOL: () => m.esports_report_type_school(),
};

// The tenant policy stays off the dial — the server refuses it until the
// tenant plane exists, so the page never offers a value it would reject.
const policyOptions: { label: () => string; value: string }[] = [
  { label: () => m.esports_report_policies_policy_open(), value: "REPORT_ACCESS_POLICY_OPEN" },
  {
    label: () => m.esports_report_policies_policy_membership(),
    value: "REPORT_ACCESS_POLICY_MEMBERSHIP",
  },
  { label: () => m.esports_report_policies_policy_closed(), value: "REPORT_ACCESS_POLICY_CLOSED" },
];

// Wire names travel the URL and the Select; the generated enums travel the
// request — the two translations live here, the only place both appear.
const familyFromWire = (wire: string): ReportType =>
  wire === "REPORT_TYPE_COGNITIVE"
    ? ReportType.COGNITIVE
    : wire === "REPORT_TYPE_CAREER"
      ? ReportType.CAREER
      : ReportType.SCHOOL;

const policyFromWire = (wire: string): ReportAccessPolicy =>
  wire === "REPORT_ACCESS_POLICY_OPEN"
    ? ReportAccessPolicy.OPEN
    : wire === "REPORT_ACCESS_POLICY_MEMBERSHIP"
      ? ReportAccessPolicy.MEMBERSHIP
      : ReportAccessPolicy.CLOSED;

const policyColors: Record<string, string> = {
  REPORT_ACCESS_POLICY_OPEN: "success",
  REPORT_ACCESS_POLICY_MEMBERSHIP: "processing",
  REPORT_ACCESS_POLICY_CLOSED: "error",
};

export function PoliciesTable({
  rows,
  loading,
  pending,
  setting,
  onRefresh,
  onSet,
  onSubmit,
}: Readonly<PoliciesTableProps>) {
  const { fieldAction, fieldTimestamp } = useColumns();
  const [pick, setPick] = useState("");

  useEffect(() => {
    if (setting) setPick(setting.current);
  }, [setting]);

  const columns: TableProps<ReportPolicyRow>["columns"] = [
    {
      title: () => m.esports_report_policies_col_family(),
      dataIndex: "reportType",
      render: (_, row) => familyLabels[row.reportType ?? ""]?.() ?? row.reportType,
    },
    {
      title: () => m.esports_report_policies_col_policy(),
      dataIndex: "policy",
      width: 130,
      render: (_, row) => (
        <Tag className="m-0" color={policyColors[String(row.policy ?? "")]}>
          {policyOptions.find((o) => o.value === String(row.policy))?.label() ?? String(row.policy)}
        </Tag>
      ),
    },
    {
      title: () => m.esports_report_policies_col_state(),
      dataIndex: "overridden",
      width: 110,
      render: (_, row) =>
        row.overridden ? (
          <Tag className="m-0">{m.esports_report_policies_state_overridden()}</Tag>
        ) : (
          <Typography.Text type="secondary">
            {m.esports_report_policies_state_default()}
          </Typography.Text>
        ),
    },
    {
      title: () => m.esports_report_policies_col_updated_by(),
      dataIndex: "updatedBy",
      width: 160,
      render: (_, row) => row.updatedBy ?? <Typography.Text type="secondary">-</Typography.Text>,
    },
    {
      ...fieldTimestamp<ReportPolicyRow>("updatedAt", {
        title: m.esports_report_policies_col_updated_at(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      ...fieldAction<ReportPolicyRow>(
        (_, row) => (
          <ActionIconButton
            icon={KeyRound}
            label={m.esports_report_policies_act_set()}
            onClick={() =>
              onSet({ reportType: String(row.reportType ?? ""), current: String(row.policy ?? "") })
            }
          />
        ),
        { title: m.esports_report_policies_col_actions() },
      ),
      width: 90,
    },
  ];

  return (
    <>
      <DataTable<ReportPolicyRow>
        columns={columns}
        dataSource={rows}
        emptyDescription={m.esports_report_policies_empty()}
        loading={loading}
        pagination={false}
        rowKey="reportType"
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
      <Modal
        cancelText={m.common_cancel()}
        okButtonProps={{ loading: pending, disabled: !setting || pick === setting.current }}
        okText={m.common_confirm()}
        onCancel={() => onSet(null)}
        onOk={() => setting && onSubmit(familyFromWire(setting.reportType), policyFromWire(pick))}
        open={!!setting}
        title={m.esports_report_policies_set_title()}
      >
        <Select
          className="w-full"
          onChange={setPick}
          options={policyOptions.map((o) => ({ label: o.label(), value: o.value }))}
          value={pick}
        />
      </Modal>
    </>
  );
}

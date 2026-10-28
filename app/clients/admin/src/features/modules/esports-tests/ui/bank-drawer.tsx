import { useQuery } from "@connectrpc/connect-query";
import type { QuestionBankRow } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin_pb";
import { listQuestionBank } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin-TestAdminService_connectquery";
import type { Test } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Drawer, Tag, Typography } from "antd";
import { useState } from "react";
import { m } from "#/paraglide/messages";

interface BankDrawerProps {
  test: Test;
  onClose: () => void;
}

// The prototype's readable columns — scenario, description, optimal
// choice — read straight off the bank's known payload/answer keys; other
// shapes fall back to the verbatim JSON rather than a guess.
export function BankDrawer({ test, onClose }: Readonly<BankDrawerProps>) {
  const [pageNo, setPageNo] = useState(1);
  const bankQuery = useQuery(listQuestionBank, {
    testCode: test.code ?? "",
    page: { pageNo, pageSize: 10 },
  });
  const { fieldTimestamp } = useColumns();

  const columns: TableProps<QuestionBankRow>["columns"] = [
    {
      title: () => m.esports_tests_bank_col_scenario(),
      dataIndex: "scenarioCode",
      width: 120,
    },
    {
      title: () => m.esports_tests_bank_col_description(),
      dataIndex: "payload",
      render: (_, row) => {
        const desc = (row.payload as Record<string, unknown> | null | undefined)?.description;
        return typeof desc === "string" ? (
          desc
        ) : (
          <pre className="m-0 max-h-40 overflow-auto text-xs">
            {JSON.stringify(row.payload ?? {}, null, 2)}
          </pre>
        );
      },
    },
    {
      title: () => m.esports_tests_bank_col_optimal(),
      dataIndex: "answer",
      width: 130,
      render: (_, row) => {
        const opt = (row.answer as Record<string, unknown> | null | undefined)?.optimal_choice;
        return typeof opt === "string" ? (
          <Tag className="m-0" color="green">
            {opt}
          </Tag>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        );
      },
    },
    {
      ...fieldTimestamp<QuestionBankRow>("createdAt", {
        title: m.esports_tests_bank_col_created(),
        format: { time: "second" },
      }),
      width: 170,
    },
  ];

  return (
    <Drawer
      onClose={onClose}
      open
      placement="right"
      title={`${m.esports_tests_bank_title()} · ${test.name ?? test.code ?? ""}`}
      width={720}
    >
      <DataTable<QuestionBankRow>
        columns={columns}
        dataSource={bankQuery.data?.questions ?? []}
        emptyDescription={m.esports_tests_bank_empty()}
        loading={bankQuery.isFetching}
        pagination={{
          current: pageNo,
          onChange: setPageNo,
          pageSize: 10,
          showSizeChanger: false,
          total: Number(bankQuery.data?.page?.total ?? 0),
        }}
        rowKey="scenarioCode"
      />
    </Drawer>
  );
}

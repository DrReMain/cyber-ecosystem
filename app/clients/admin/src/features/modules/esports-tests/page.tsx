import { useQuery } from "@connectrpc/connect-query";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin-TestAdminService_connectquery";
import type { Test } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { Card, Select } from "antd";
import { useMemo, useState } from "react";
import { m } from "#/paraglide/messages";

import { BankDrawer } from "./ui/bank-drawer";
import { ConfigModal } from "./ui/config-modal";
import { TestsTable } from "./ui/tests-table";

// Thirteen rows at seed, flat by design; the category dial filters locally —
// the tuning and bank faces hang off each row's actions.
export function EsportsTestsPage() {
  const listQuery = useQuery(listTests, {});
  const [config, setConfig] = useState<Test | null>(null);
  const [bank, setBank] = useState<Test | null>(null);
  const [category, setCategory] = useState<string>();
  const rows = useMemo(
    () =>
      category
        ? (listQuery.data?.tests ?? []).filter((t) => t.category === category)
        : (listQuery.data?.tests ?? []),
    [listQuery.data, category],
  );

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Select
          allowClear
          className="min-w-36"
          onChange={setCategory}
          options={[
            { label: m.esports_tests_category_fps(), value: "fps" },
            { label: m.esports_tests_category_moba(), value: "moba" },
            { label: m.esports_tests_category_universal(), value: "universal" },
          ]}
          placeholder={m.esports_tests_filter_category()}
          value={category}
        />
      </Card>
      <TestsTable
        loading={listQuery.isFetching}
        onOpenBank={setBank}
        onOpenConfig={setConfig}
        onRefresh={() => listQuery.refetch()}
        rows={rows}
      />
      {config ? <ConfigModal onClose={() => setConfig(null)} test={config} /> : null}
      {bank ? <BankDrawer onClose={() => setBank(null)} test={bank} /> : null}
    </div>
  );
}

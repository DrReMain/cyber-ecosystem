import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  listReportAccessPolicies,
  setReportAccessPolicy,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { useState } from "react";
import { PoliciesTable, type PolicyRowKey } from "./ui/policies-table";

// Three rows, no filters, no pagination — one settled table where a default
// and an override read the same; the set action mutates and invalidates.
export function ReportPoliciesPage() {
  const listQuery = useQuery(listReportAccessPolicies, {});
  const [setting, setSetting] = useState<PolicyRowKey | null>(null);
  const setMut = useMutation(setReportAccessPolicy, {
    onSuccess: () => {
      setSetting(null);
      void listQuery.refetch();
    },
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      <PoliciesTable
        loading={listQuery.isFetching}
        onRefresh={() => listQuery.refetch()}
        onSet={(row) => setSetting(row)}
        onSubmit={(reportType, policy) => setMut.mutate({ reportType, policy })}
        pending={setMut.isPending}
        rows={listQuery.data?.policies ?? []}
        setting={setting}
      />
    </div>
  );
}

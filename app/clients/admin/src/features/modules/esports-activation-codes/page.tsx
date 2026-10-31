import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  generateActivationCodes,
  listActivationCodes,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Button, Card, Select } from "antd";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { codesSearchSchema, parseCodesSearch, reportTypeFromWire } from "./search";
import { CodesTable } from "./ui/codes-table";
import { GenerateModal } from "./ui/generate-modal";

const AREA = "/dashboard/channels/activation-codes/";

const reportTypeOptions = [
  { label: () => m.esports_report_type_cognitive(), value: "REPORT_TYPE_COGNITIVE" },
  { label: () => m.esports_report_type_career(), value: "REPORT_TYPE_CAREER" },
  { label: () => m.esports_report_type_school(), value: "REPORT_TYPE_SCHOOL" },
];

export function ActivationCodesPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseCodesSearch);
  const [generateOpen, setGenerateOpen] = useState(false);
  const listQuery = useQuery(listActivationCodes, {
    reportType: reportTypeFromWire(search.reportType),
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, codesSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.esports_activation_codes_total({ n: t }),
  });
  const generateMut = useMutation(generateActivationCodes, {
    onSuccess: () => listQuery.refetch(),
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <div className="flex items-center justify-between gap-2">
          <Select
            allowClear
            className="min-w-44"
            onChange={(v) => store.patch({ reportType: v, pageNo: 1 })}
            options={reportTypeOptions.map((o) => ({ label: o.label(), value: o.value }))}
            placeholder={m.esports_activation_codes_filter_type()}
            value={search.reportType}
          />
          <Button icon={<Plus size={14} />} onClick={() => setGenerateOpen(true)} type="primary">
            {m.esports_activation_codes_generate()}
          </Button>
        </div>
      </Card>
      <CodesTable
        loading={listQuery.isFetching}
        onRefresh={() => listQuery.refetch()}
        pagination={pagination}
        rows={listQuery.data?.codes ?? []}
      />
      <GenerateModal
        onClose={() => setGenerateOpen(false)}
        onSubmit={(req) => generateMut.mutate(req)}
        open={generateOpen}
        result={generateMut.data ?? null}
        submitting={generateMut.isPending}
      />
    </div>
  );
}

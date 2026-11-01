import { useMutation, useQuery } from "@connectrpc/connect-query";
import { ActivationCodeStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import {
  exportActivationCodes,
  generateActivationCodes,
  listActivationBatches,
  setActivationBatchStatus,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Button, Card, Select } from "antd";
import { Download, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { fullDate } from "#/features/modules/esports-members/ui/dates";
import { downloadCsv, useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { codesSearchSchema, parseCodesSearch, reportTypeFromWire } from "./search";
import { BatchesTable } from "./ui/batches-table";
import { GenerateModal } from "./ui/generate-modal";
import { reportTypeOptions } from "./vocab";

const AREA = "/dashboard/channels/activation-codes/";

const stateOptions = [
  { label: () => m.esports_activation_codes_filter_live(), value: "live" },
  { label: () => m.esports_activation_codes_filter_disabled(), value: "disabled" },
];

export function ActivationCodesPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseCodesSearch);
  const [generateOpen, setGenerateOpen] = useState(false);
  const listQuery = useQuery(listActivationBatches, {
    reportType: reportTypeFromWire(search.reportType),
    disabled: search.state ? search.state === "disabled" : undefined,
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, codesSearchSchema, listQuery.data?.page, {
    showTotal: (t) =>
      m.esports_activation_codes_footer({
        batches: listQuery.data?.totalBatches ?? 0,
        codes: (listQuery.data?.totalCodes ?? 0).toLocaleString("en-US"),
        used: (listQuery.data?.usedCodes ?? 0).toLocaleString("en-US"),
        pct:
          (listQuery.data?.totalCodes ?? 0) > 0
            ? (
                ((listQuery.data?.usedCodes ?? 0) / (listQuery.data?.totalCodes ?? 1)) *
                100
              ).toFixed(1)
            : "0",
        t,
      }),
  });
  const generateMut = useMutation(generateActivationCodes, {
    onSuccess: () => listQuery.refetch(),
  });
  const flipMut = useMutation(setActivationBatchStatus, {
    onSuccess: (_d, vars) => {
      toast.success(
        vars.disabled
          ? m.esports_activation_codes_disable_ok({ no: vars.batchNo ?? "" })
          : m.esports_activation_codes_enable_ok({ no: vars.batchNo ?? "" }),
      );
      void listQuery.refetch();
    },
  });
  const exportMut = useMutation(exportActivationCodes, {
    onSuccess: (resp) => {
      // The export hands out usable code values — the CSV lands in the
      // operator's hands through the audited RPC, never a client-side
      // page scrape.
      downloadCsv(
        "activation_codes_unused.csv",
        [
          m.esports_activation_codes_csv_code(),
          m.esports_activation_codes_csv_batch(),
          m.esports_activation_codes_csv_type(),
          m.esports_activation_codes_csv_expires(),
        ],
        (resp.codes ?? []).map((c) => [
          c.code ?? "",
          c.batchNo ?? "",
          c.reportType ?? "",
          c.expiresAt ? fullDate(c.expiresAt) : "",
        ]),
      );
      toast.success(m.esports_activation_codes_export_ok({ n: resp.codes?.length ?? 0 }));
    },
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <Select
              allowClear
              className="min-w-44"
              onChange={(v) => store.patch({ reportType: v, pageNo: 1 })}
              options={reportTypeOptions.map((o) => ({ label: o.label(), value: o.value }))}
              placeholder={m.esports_activation_codes_filter_type()}
              value={search.reportType}
            />
            <Select
              allowClear
              className="min-w-32"
              onChange={(v) => store.patch({ state: v, pageNo: 1 })}
              options={stateOptions.map((o) => ({ label: o.label(), value: o.value }))}
              placeholder={m.esports_activation_codes_filter_state()}
              value={search.state}
            />
          </div>
          <div className="flex items-center gap-2">
            <Button
              icon={<Download size={14} />}
              loading={exportMut.isPending}
              onClick={() => exportMut.mutate({ codeStatus: ActivationCodeStatus.UNUSED })}
            >
              {m.esports_activation_codes_export()}
            </Button>
            <Button icon={<Plus size={14} />} onClick={() => setGenerateOpen(true)} type="primary">
              {m.esports_activation_codes_generate()}
            </Button>
          </div>
        </div>
      </Card>
      <BatchesTable
        loading={listQuery.isFetching}
        onFlip={(row, disabled) => flipMut.mutate({ batchNo: row.batchNo ?? "", disabled })}
        onRefresh={() => listQuery.refetch()}
        pagination={pagination}
        rows={listQuery.data?.batches ?? []}
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

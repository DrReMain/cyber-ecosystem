import { useMutation, useQuery } from "@connectrpc/connect-query";
import { ActivationCodeStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import {
  exportActivationCodes,
  listActivationBatches,
  listActivationCodes,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { Link, useSearch } from "@tanstack/react-router";
import { Button, Card, Input, Progress, Segmented, Tag } from "antd";
import { ArrowLeft, Copy, Download } from "lucide-react";
import { toast } from "sonner";

import { fullDate } from "#/features/modules/esports-members/ui/dates";
import { downloadCsv, useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { batchCodesSearchSchema, parseBatchCodesSearch } from "./search";
import { CodesTable } from "./ui/codes-table";

const AREA = "/dashboard/channels/activation-codes/$batchNo";

// The prototype's restock bar threshold, shared with the list page.
const RESTOCK_AT = 95;

const csvHeader = () => [
  m.esports_activation_batch_csv_code(),
  m.esports_activation_batch_csv_state(),
  m.esports_activation_batch_csv_user(),
  m.esports_activation_batch_csv_redeemed(),
  m.esports_activation_batch_csv_expires(),
  m.esports_activation_batch_csv_batch(),
];

export function ActivationBatchPage({ batchNo }: { batchNo: string }) {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseBatchCodesSearch);
  const batchQuery = useQuery(listActivationBatches, {
    batchNo,
    // The exact-match lookup still rides the paged shape — the paginate
    // floor refuses a pageless ask.
    page: { pageNo: 1, pageSize: 1 },
  });
  const batch = batchQuery.data?.batches?.[0] ?? null;
  const codesQuery = useQuery(listActivationCodes, {
    batchNo,
    code: search.code?.trim() || undefined,
    codeStatus: search.status ? statusToWire(search.status) : undefined,
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, batchCodesSearchSchema, codesQuery.data?.page, {
    pageSizeOptions: [50, 100, 200],
  });

  const exportMut = useMutation(exportActivationCodes, {
    onSuccess: (resp) => {
      downloadCsv(
        `activation_codes_${batchNo}.csv`,
        csvHeader(),
        (resp.codes ?? []).map((c) => [
          c.code ?? "",
          Number(c.usedCount ?? 0) > 0
            ? m.esports_activation_batch_state_redeemed()
            : m.esports_activation_batch_state_unused(),
          c.nickname ?? c.userId ?? "",
          c.redeemedAt ? fullDate(c.redeemedAt) : "",
          c.expiresAt ? fullDate(c.expiresAt) : "",
          c.batchNo ?? batchNo,
        ]),
      );
      toast.success(m.esports_activation_batch_export_ok({ n: resp.codes?.length ?? 0 }));
    },
  });

  const copyUnused = async () => {
    // Copying the unspent remainder rides the audited export — the plain
    // clipboard write has no audit line of its own.
    const resp = await exportMut
      .mutateAsync({ batchNo, codeStatus: ActivationCodeStatus.UNUSED })
      .catch(() => null);
    if (!resp) {
      return; // the error cache already surfaced the reason
    }
    const codes = (resp.codes ?? [])
      .map((c) => c.code ?? "")
      .filter(Boolean)
      .join("\n");
    await navigator.clipboard.writeText(codes);
    toast.success(m.esports_activation_batch_copy_ok({ n: resp.codes?.length ?? 0 }));
  };

  const count = Number(batch?.count ?? 0);
  const used = Number(batch?.usedCount ?? 0);
  const pct = count > 0 ? (used / count) * 100 : 0;

  return (
    <div className="flex flex-col gap-4 p-4">
      <Link
        className="text-muted-foreground hover:text-foreground flex items-center gap-1 text-sm"
        search={{ pageNo: 1, pageSize: 20 }}
        to="/dashboard/channels/activation-codes"
      >
        <ArrowLeft size={14} />
        {m.esports_activation_batch_back()}
      </Link>
      <Card>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <h2 className="font-mono text-lg font-bold">{batchNo}</h2>
            {batch?.disabledAt ? (
              <Tag className="m-0">{m.esports_activation_codes_state_disabled()}</Tag>
            ) : (
              <Tag className="m-0" color="success">
                {m.esports_activation_codes_state_live()}
              </Tag>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button
              icon={<Copy size={14} />}
              loading={exportMut.isPending}
              onClick={() => void copyUnused()}
            >
              {m.esports_activation_batch_copy()}
            </Button>
            <Button
              icon={<Download size={14} />}
              loading={exportMut.isPending}
              onClick={() =>
                exportMut.mutate({
                  batchNo,
                  code: search.code?.trim() || undefined,
                  codeStatus: search.status ? statusToWire(search.status) : undefined,
                })
              }
              type="primary"
            >
              {m.esports_activation_batch_export({ n: codesQuery.data?.page?.total ?? 0 })}
            </Button>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-2 gap-4 md:grid-cols-4">
          <div>
            <div className="text-muted-foreground text-xs">
              {m.esports_activation_batch_kv_gen()}
            </div>
            <div className="text-lg font-bold tabular-nums">
              {count.toLocaleString("en-US")}
              <span className="text-muted-foreground text-xs font-normal">
                {" / "}
                {m.esports_activation_batch_kv_used({ n: used.toLocaleString("en-US") })}
              </span>
            </div>
          </div>
          <div>
            <div className="text-muted-foreground text-xs">
              {m.esports_activation_batch_kv_progress()}
            </div>
            <div className="flex items-center gap-2">
              <span className="text-lg font-bold tabular-nums">{pct.toFixed(1)}%</span>
              <Progress
                className="w-16"
                percent={Math.round(pct)}
                showInfo={false}
                size="small"
                strokeColor={pct >= RESTOCK_AT ? "#eab308" : undefined}
              />
              {pct >= RESTOCK_AT ? (
                <Tag className="m-0" color="warning">
                  {m.esports_activation_codes_restock()}
                </Tag>
              ) : null}
            </div>
          </div>
          <div>
            <div className="text-muted-foreground text-xs">
              {m.esports_activation_batch_kv_unused()}
            </div>
            <div className="text-lg font-bold text-emerald-600 tabular-nums">
              {(count - used).toLocaleString("en-US")}
            </div>
          </div>
          <div>
            <div className="text-muted-foreground text-xs">
              {m.esports_activation_batch_kv_expires()}
            </div>
            <div className="text-lg font-bold tabular-nums">
              {batch?.expiresAt ? fullDate(batch.expiresAt) : "-"}
            </div>
          </div>
        </div>
      </Card>
      <Card>
        <div className="flex flex-wrap items-center gap-3">
          <Segmented
            onChange={(v) =>
              store.patch({
                status: v === "all" ? undefined : (v as "unused" | "redeemed"),
                pageNo: 1,
              })
            }
            options={[
              { label: m.esports_activation_batch_filter_all(), value: "all" },
              { label: m.esports_activation_batch_state_unused(), value: "unused" },
              { label: m.esports_activation_batch_state_redeemed(), value: "redeemed" },
            ]}
            value={search.status ?? "all"}
          />
          <Input
            allowClear
            className="w-55"
            onChange={(e) => store.patch({ code: e.target.value.trim() || undefined, pageNo: 1 })}
            placeholder={m.esports_activation_batch_filter_code()}
            value={search.code ?? ""}
          />
        </div>
      </Card>
      <CodesTable
        loading={codesQuery.isFetching}
        onRefresh={() => {
          void codesQuery.refetch();
          void batchQuery.refetch();
        }}
        pagination={pagination}
        rows={codesQuery.data?.codes ?? []}
      />
    </div>
  );
}

// The URL carries the wire words; the request carries the generated enum.
const statusToWire = (status: string) =>
  status === "redeemed" ? ActivationCodeStatus.REDEEMED : ActivationCodeStatus.UNUSED;

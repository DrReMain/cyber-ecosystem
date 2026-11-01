import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  exportAgreementConsents,
  listAgreementConsents,
  listAgreementVersions,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin-AgreementAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Button, Card, DatePicker, Input, Select, Typography } from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { Download, RotateCcw } from "lucide-react";
import { toast } from "sonner";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { downloadCsv, useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { docFromWire, docOptions, loginMethodLabel } from "../esports-agreement-versions/vocab";
import { consentsSearchSchema, parseConsentsSearch } from "./search";
import { ConsentsTable } from "./ui/consents-table";

const AREA = "/dashboard/compliance/agreement-consents/";

export function AgreementConsentsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseConsentsSearch);

  // The version dropdown reads the version ledger, not the consent rows'
  // distinct values: deleted versions stay queryable (their evidence
  // rows keep the string), so the list is the narrower source.
  const versionsQuery = useQuery(listAgreementVersions, {});
  const versionOptions = [...new Set((versionsQuery.data?.versions ?? []).map((v) => v.no ?? ""))]
    .sort()
    .reverse()
    .map((no) => ({ label: no, value: no }));

  // A date filter is a whole-day window: from starts at midnight, to
  // ends at the last instant of its day.
  const rangeFrom = (from?: string) =>
    from ? timestampFromMs(dayjs(`${from}T00:00:00`).valueOf()) : undefined;
  const rangeTo = (to?: string) =>
    to ? timestampFromMs(dayjs(`${to}T23:59:59.999`).valueOf()) : undefined;

  const filters = {
    doc: docFromWire(search.doc),
    keyword: search.keyword,
    page: {
      createdAtA: rangeFrom(search.from),
      createdAtZ: rangeTo(search.to),
      pageNo: search.pageNo,
      pageSize: search.pageSize,
    },
    versionNo: search.ver,
  };
  const listQuery = useQuery(listAgreementConsents, filters);
  const pagination = useServerPagination(store, consentsSearchSchema, listQuery.data?.page, {
    showTotal: (t) => m.esports_agreement_consents_footer({ n: t }),
  });

  // The export answers the visible filter, never a wider set — the same
  // argument object the list reads, so the two can never disagree.
  const exportMut = useMutation(exportAgreementConsents, {
    onSuccess: (resp) => {
      downloadCsv(
        "agreement_consents.csv",
        [
          m.esports_agreement_consents_csv_user(),
          m.esports_agreement_consents_csv_phone(),
          m.esports_agreement_consents_csv_doc(),
          m.esports_agreement_consents_csv_ver(),
          m.esports_agreement_consents_csv_time(),
          m.esports_agreement_consents_csv_method(),
        ],
        (resp.consents ?? []).map((r) => [
          r.nickname ?? r.userId ?? "",
          r.phoneMasked ?? "",
          docOptions.find((o) => o.value === String(r.doc))?.label() ?? String(r.doc),
          r.versionNo ?? "",
          r.consentedAt ? fullTime(r.consentedAt) : "",
          loginMethodLabel(String(r.loginMethod)),
        ]),
      );
      toast.success(m.esports_agreement_consents_export_ok({ n: resp.consents?.length ?? 0 }));
    },
  });

  const range: [Dayjs | null, Dayjs | null] = [
    search.from ? dayjs(search.from) : null,
    search.to ? dayjs(search.to) : null,
  ];

  return (
    <div className="flex flex-col gap-3 p-4">
      <div>
        <Typography.Title className="!mb-0" level={4}>
          {m.esports_agreement_consents_title()}
        </Typography.Title>
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_agreement_consents_desc()}
        </Typography.Text>
      </div>

      <Card>
        <div className="flex flex-wrap items-center gap-2">
          <Select
            allowClear
            className="min-w-36"
            onChange={(v) => store.patch({ doc: v, pageNo: 1 })}
            options={docOptions.map((o) => ({ label: o.label(), value: o.value }))}
            placeholder={m.esports_agreement_consents_filter_doc()}
            value={search.doc}
          />
          <Select
            allowClear
            className="min-w-28"
            onChange={(v) => store.patch({ ver: v, pageNo: 1 })}
            options={versionOptions}
            placeholder={m.esports_agreement_consents_filter_ver()}
            value={search.ver}
          />
          <Input
            className="w-44"
            onChange={(e) => store.patch({ keyword: e.target.value, pageNo: 1 })}
            placeholder={
              search.doc || search.ver
                ? m.esports_agreement_consents_filter_keyword_narrowed()
                : m.esports_agreement_consents_filter_keyword()
            }
            value={search.keyword}
          />
          <DatePicker.RangePicker
            onChange={(dates) =>
              store.patch({
                from: dates?.[0]?.format("YYYY-MM-DD"),
                pageNo: 1,
                to: dates?.[1]?.format("YYYY-MM-DD"),
              })
            }
            value={range}
          />
          <Button
            icon={<RotateCcw size={14} />}
            onClick={() =>
              store.patch({
                doc: undefined,
                from: undefined,
                keyword: undefined,
                pageNo: 1,
                to: undefined,
                ver: undefined,
              })
            }
          >
            {m.esports_agreement_consents_filter_reset()}
          </Button>
          <Button
            className="ml-auto"
            icon={<Download size={14} />}
            loading={exportMut.isPending}
            onClick={() =>
              exportMut.mutate({
                createdAtA: rangeFrom(search.from),
                createdAtZ: rangeTo(search.to),
                doc: docFromWire(search.doc),
                keyword: search.keyword,
                versionNo: search.ver,
              })
            }
          >
            {m.esports_agreement_consents_export()}
          </Button>
        </div>
      </Card>

      <ConsentsTable
        loading={listQuery.isFetching}
        onRefresh={() => void listQuery.refetch()}
        pagination={pagination}
        rows={listQuery.data?.consents ?? []}
      />
    </div>
  );
}

import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Tenant } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant_pb";
import {
  listTenants,
  updateTenantStatus,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant-TenantService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useMemo } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parseTenantsSearch, tenantsSearchSchema } from "./search";
import { TenantsTable } from "./ui/tenants-table";

const TENANTS_INPUT = { orderBy: ["code:asc"], page: { all: true } };

function matchKeyword(tenant: Tenant, keyword: string): boolean {
  if (!keyword) return true;
  const inName = tenant.name?.toLowerCase().includes(keyword) ?? false;
  const inCode = tenant.code?.toLowerCase().includes(keyword) ?? false;
  return inName || inCode;
}

function matchStatus(tenant: Tenant, status: string | undefined): boolean {
  if (status === "on") return tenant.enabled ?? false;
  if (status === "off") return !(tenant.enabled ?? false);
  return true;
}

function applySearch(tenants: Tenant[], values: { keyword?: string; status?: string }): Tenant[] {
  const keyword = values.keyword?.trim().toLowerCase();
  return tenants.filter(
    (tenant) => matchKeyword(tenant, keyword ?? "") && matchStatus(tenant, values.status),
  );
}

const AREA = "/dashboard/platform/tenants";

export function TenantsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseTenantsSearch);
  const { values, onFilter, onReset } = useFilter(store, tenantsSearchSchema);

  const tenantsQuery = useQuery(listTenants, TENANTS_INPUT);
  const toggleMut = useMutation(updateTenantStatus, {
    onSuccess: () => void tenantsQuery.refetch(),
  });
  const pendingId = toggleMut.isPending ? (toggleMut.variables?.id ?? null) : null;

  const tenants = useMemo(
    () => applySearch(tenantsQuery.data?.list ?? [], values),
    [tenantsQuery.data, values],
  );

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Filter
          columns={3}
          initialValues={values}
          labels={{
            search: m.common_filter_search(),
            reset: m.common_filter_reset(),
            fold: m.common_filter_fold(),
            expand: m.common_filter_expand(),
          }}
          onFilter={onFilter}
          onReset={onReset}
          options={[
            {
              label: m.platform_tenants_filter_keyword(),
              name: "keyword",
              placeholder: m.platform_tenants_filter_keyword_placeholder(),
            },
            {
              label: m.platform_tenants_filter_status(),
              name: "status",
              type: "select",
              placeholder: m.platform_tenants_filter_any(),
              options: [
                { label: m.platform_tenants_status_on(), value: "on" },
                { label: m.platform_tenants_status_off(), value: "off" },
              ],
            },
          ]}
        />
      </Card>
      <TenantsTable
        loading={tenantsQuery.isFetching}
        onRefresh={() => void tenantsQuery.refetch()}
        onToggle={(tenant, next) => toggleMut.mutate({ id: tenant.id ?? "", enabled: next })}
        pendingId={pendingId}
        tenants={tenants}
      />
    </div>
  );
}

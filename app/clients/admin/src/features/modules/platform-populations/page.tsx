import { useQuery } from "@connectrpc/connect-query";
import type { Population } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/population_pb";
import { listPopulations } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/population-PopulationService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useMemo } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parsePopulationsSearch, populationsSearchSchema } from "./search";
import { PopulationsTable } from "./ui/populations-table";

const POPULATIONS_INPUT = { orderBy: ["code:asc"], page: { all: true } };

function matchKeyword(population: Population, keyword: string): boolean {
  if (!keyword) return true;
  const inName = population.name?.toLowerCase().includes(keyword) ?? false;
  const inCode = population.code?.toLowerCase().includes(keyword) ?? false;
  return inName || inCode;
}

const AREA = "/dashboard/platform/populations";

export function PopulationsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parsePopulationsSearch);
  const { values, onFilter, onReset } = useFilter(store, populationsSearchSchema);

  const populationsQuery = useQuery(listPopulations, POPULATIONS_INPUT);

  const populations = useMemo(
    () =>
      (populationsQuery.data?.list ?? []).filter((population) =>
        matchKeyword(population, values.keyword?.trim().toLowerCase() ?? ""),
      ),
    [populationsQuery.data, values],
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
              label: m.platform_populations_filter_keyword(),
              name: "keyword",
              placeholder: m.platform_populations_filter_keyword_placeholder(),
            },
          ]}
        />
      </Card>
      <PopulationsTable
        loading={populationsQuery.isFetching}
        onRefresh={() => void populationsQuery.refetch()}
        populations={populations}
      />
    </div>
  );
}

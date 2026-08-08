import { deptServiceCreateDept, deptServiceListDepts } from "@cyber-ecosystem/gen-openapi-ts";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { httpClient } from "#/services/http";

const LIST_QUERY = { "page.pageNo": 1, "page.pageSize": 10, "page.all": true } as const;

export const Route = createFileRoute("/_app/playground-http")({
  component: HttpPlayground,
});

function HttpPlayground() {
  const { data, isFetching, refetch } = useQuery({
    queryKey: ["depts"],
    queryFn: async () =>
      (await deptServiceListDepts({ client: httpClient, query: LIST_QUERY })).data,
  });
  const createMut = useMutation({
    mutationFn: async (args: { name: string }) =>
      (await deptServiceCreateDept({ body: args, client: httpClient })).data,
    onSuccess: () => refetch(),
  });
  const [name, setName] = useState("");

  return (
    <div className="space-y-3">
      <h2 className="font-semibold text-lg">Dept (http)</h2>
      <div className="flex gap-2">
        <input
          className="border px-2 py-1"
          onChange={(e) => setName(e.target.value)}
          value={name}
        />
        <button
          className="border px-2 py-1"
          onClick={() => {
            if (!name) return;
            createMut.mutate({ name });
            setName("");
          }}
          type="submit"
        >
          create
        </button>
      </div>

      {data ? (
        <pre className="overflow-auto text-xs">{JSON.stringify(data, null, 2)}</pre>
      ) : (
        <p className="text-sm opacity-70">loading…</p>
      )}
      {isFetching ? <p className="text-sm opacity-70">refreshing…</p> : null}
    </div>
  );
}

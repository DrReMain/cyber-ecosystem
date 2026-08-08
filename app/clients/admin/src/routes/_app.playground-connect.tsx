import { toJson } from "@bufbuild/protobuf";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { ListDeptsResponseSchema } from "@cyber-ecosystem/gen-ts/cyber/system/v1/dept_pb";
import {
  createDept,
  listDepts,
} from "@cyber-ecosystem/gen-ts/cyber/system/v1/dept-DeptService_connectquery";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

const LIST_QUERY = { page: { pageNo: 1, pageSize: 10, all: true } } as const;

export const Route = createFileRoute("/_app/playground-connect")({
  component: ConnectPlayground,
});

function ConnectPlayground() {
  const { data, isFetching, refetch } = useQuery(listDepts, LIST_QUERY);
  const createMut = useMutation(createDept, { onSuccess: () => refetch() });
  const [name, setName] = useState("");

  return (
    <div className="space-y-3">
      <h2 className="font-semibold text-lg">Dept (connect)</h2>
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
        <pre className="overflow-auto text-xs">
          {JSON.stringify(
            toJson(ListDeptsResponseSchema, data, { alwaysEmitImplicit: true }),
            null,
            2,
          )}
        </pre>
      ) : (
        <p className="text-sm opacity-70">loading…</p>
      )}
      {isFetching ? <p className="text-sm opacity-70">refreshing…</p> : null}
    </div>
  );
}

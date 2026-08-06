import { useQuery } from "@connectrpc/connect-query";
import { listResource } from "@cyber-ecosystem/gen-ts/cyber/system/v1/resource-ResourceService_connectquery";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_app/playground-connect")({
  component: ConnectDemo,
});

function ConnectDemo() {
  const { data, isLoading, error } = useQuery(listResource, {});
  if (isLoading) return <p className="text-sm opacity-70">loading…</p>;
  if (error) return <p className="text-red-500 text-sm">error: {String(error)}</p>;
  return (
    <div className="space-y-3">
      <h2 className="font-semibold text-lg">ListResource (connect)</h2>
      <pre className="overflow-auto text-xs">{JSON.stringify(data, null, 2)}</pre>
    </div>
  );
}

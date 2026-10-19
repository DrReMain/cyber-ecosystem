import { Skeleton, Typography } from "antd";
import clsx from "clsx";
import { m } from "#/paraglide/messages";
import { formatGrantsSummary, summarizePatterns } from "../../system-roles/grants";
import type { OperationCatalog } from "../types";

interface PreviewGroup {
  key: string;
  // A "svc/*" wildcard covers every method of this service.
  wildcard: boolean;
  methods: string[];
  folded: number;
}

interface PreviewGroups {
  global: boolean;
  groups: PreviewGroup[];
  foldedServices: number;
}

function groupOperations(operations: string[]): PreviewGroups {
  const out: PreviewGroups = { global: false, groups: [], foldedServices: 0 };
  const byService = new Map<string, PreviewGroup>();
  for (const op of operations) {
    const rest = op.replace(/^\//, "");
    if (rest === "*") {
      out.global = true;
      continue;
    }
    const slash = rest.lastIndexOf("/");
    if (slash <= 0) continue;
    const svc = rest.slice(0, slash);
    const method = rest.slice(slash + 1);
    let g = byService.get(svc);
    if (g === undefined) {
      g = { key: svc, wildcard: false, methods: [], folded: 0 };
      byService.set(svc, g);
    }
    if (method === "*") g.wildcard = true;
    else g.methods.push(method);
  }
  const groups = [...byService.values()].sort((a, b) => (a.key < b.key ? -1 : 1));
  for (const g of groups) {
    g.methods.sort();
    if (g.wildcard) {
      g.folded += g.methods.length;
      g.methods = [];
    }
  }
  if (out.global) {
    out.foldedServices = groups.length;
  } else {
    out.groups = groups;
  }
  return out;
}

interface GrantPreviewProps {
  operations: string[];
  catalog: OperationCatalog;
  loading: boolean;
}

export function GrantPreview({ operations, catalog, loading }: Readonly<GrantPreviewProps>) {
  if (loading && operations.length === 0) {
    return <Skeleton active paragraph={{ rows: 2 }} />;
  }
  if (operations.length === 0) {
    return (
      <div className="rounded-lg bg-fill-quaternary px-3 py-3">
        <Typography.Text className="text-[12px]" type="secondary">
          {m.system_users_preview_empty()}
        </Typography.Text>
      </div>
    );
  }
  const { global, groups, foldedServices } = groupOperations(operations);
  return (
    <div
      className={clsx(
        "flex flex-col gap-2 rounded-lg bg-fill-quaternary px-3 py-2 transition-opacity",
        loading && "opacity-60",
      )}
    >
      <Typography.Text className="text-[12px]" type="secondary">
        {formatGrantsSummary(summarizePatterns(operations))}
      </Typography.Text>
      {global ? (
        <>
          <span className="flex items-baseline gap-2">
            <span className="font-medium text-[13px]">{m.system_users_preview_all_services()}</span>
            <span className="font-mono text-[11px] text-ink-tertiary">/*</span>
          </span>
          {foldedServices > 0 && (
            <Typography.Text className="text-[12px]" type="secondary">
              {m.system_users_preview_folded_services({ n: foldedServices })}
            </Typography.Text>
          )}
        </>
      ) : (
        groups.map((g) => {
          const meta = catalog.get(g.key);
          return (
            <div className="flex flex-col gap-1" key={g.key}>
              <span className="flex items-baseline gap-2">
                <span className="font-medium text-[13px]">
                  {meta?.comment || g.key.split(".").pop()}
                </span>
                <Typography.Text className="font-mono text-[11px]" type="secondary">
                  {g.key}
                </Typography.Text>
              </span>
              {g.wildcard ? (
                <Typography.Text className="ps-1 text-[13px]" type="secondary">
                  {m.system_users_preview_wildcard()}
                </Typography.Text>
              ) : (
                <div className="flex flex-col gap-0.5 ps-1">
                  {g.methods.map((mt) => (
                    <span className="flex min-h-6 items-center gap-2" key={mt}>
                      <span className="text-[13px]">{meta?.methods.get(mt) || mt}</span>
                      <Typography.Text className="font-mono text-[12px]" type="secondary">
                        {mt}
                      </Typography.Text>
                    </span>
                  ))}
                </div>
              )}
              {g.wildcard && g.folded > 0 && (
                <Typography.Text className="ps-1 text-[12px]" type="secondary">
                  {m.system_users_preview_folded_methods({ n: g.folded })}
                </Typography.Text>
              )}
            </div>
          );
        })
      )}
      <Typography.Text className="text-[11px]" type="secondary">
        {m.system_users_preview_footnote()}
      </Typography.Text>
    </div>
  );
}

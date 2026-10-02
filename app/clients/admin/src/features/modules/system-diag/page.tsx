import { useMutation, useQuery } from "@connectrpc/connect-query";
import { listResource } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource-ResourceService_connectquery";
import {
  explainOperation,
  listRolesByPrincipal,
  previewGrants,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role-RoleService_connectquery";
import { listUsers } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { keepPreviousData, useQuery as useRqQuery } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Button, Card, Select, Typography } from "antd";
import { Play } from "lucide-react";
import { useMemo, useState } from "react";
import { m } from "#/paraglide/messages";
import { sessionQuery } from "../../layout-dashboard/auth/session.fn";
import { buildOperationCatalog } from "../system-users/types";
import { GrantPreview } from "../system-users/ui/grant-preview";
import { diagSearchSchema } from "./search";
import { OperationSelect } from "./ui/operation-select";
import { SessionPanel } from "./ui/session-panel";
import { type ExplainView, VerdictCard } from "./ui/verdict-card";

export function DiagPage() {
  const search = useSearch({ from: "/dashboard/system/diag" });
  const navigate = useNavigate();
  const [subjectId, setSubjectId] = useState<string | undefined>(search.userId);
  const [operation, setOperation] = useState<string | undefined>();

  const usersQuery = useQuery(listUsers, { orderBy: ["createdAt:asc"], page: { all: true } });
  const catalogQuery = useQuery(listResource, {});
  const principalRolesQuery = useQuery(
    listRolesByPrincipal,
    { principalType: "user", principalId: subjectId ?? "" },
    { enabled: subjectId !== undefined },
  );
  const roleCodes = useMemo(
    () => (principalRolesQuery.data?.list ?? []).map((r) => r.code ?? ""),
    [principalRolesQuery.data],
  );
  const previewQuery = useQuery(
    previewGrants,
    { roleCodes },
    {
      enabled: subjectId !== undefined && principalRolesQuery.isSuccess,
      placeholderData: keepPreviousData,
    },
  );
  const session = useRqQuery(sessionQuery);

  const explainMut = useMutation(explainOperation);

  const userOptions = useMemo(
    () =>
      (usersQuery.data?.list ?? []).map((u) => ({
        value: u.id ?? "",
        label: (
          <span className="flex items-center gap-2">
            <span>{u.email?.split("@")[0]}</span>
            <span className="font-mono text-[12px] text-ink-tertiary">{u.email}</span>
            {(u.roles ?? []).length === 0 && (
              <Typography.Text className="text-[12px]" type="secondary">
                {m.system_users_unassigned()}
              </Typography.Text>
            )}
          </span>
        ),
        search: u.email ?? "",
      })),
    [usersQuery.data],
  );
  const catalog = useMemo(
    () => buildOperationCatalog(catalogQuery.data?.list ?? []),
    [catalogQuery.data],
  );
  const roleByCode = useMemo(
    () =>
      new Map(
        (principalRolesQuery.data?.list ?? []).map((r) => [
          r.code ?? "",
          { name: r.name ?? "", enabled: r.enabled ?? false },
        ]),
      ),
    [principalRolesQuery.data],
  );

  const onSubject = (next: string) => {
    setSubjectId(next);
    explainMut.reset();
    navigate({
      to: "/dashboard/system/diag",
      search: (prev) => diagSearchSchema.parse({ ...prev, userId: next }),
      replace: true,
    });
  };

  const runExplain = () => {
    if (subjectId === undefined || operation === undefined) return;
    explainMut.mutate({
      principalType: "user",
      principalId: subjectId,
      operation,
    });
  };

  const explainView: ExplainView | null =
    explainMut.data === undefined
      ? null
      : {
          verdict: explainMut.data.verdict ?? "DENY",
          roleCodes: explainMut.data.roleCodes ?? [],
          hits: (explainMut.data.hits ?? []).map((h) => ({
            pattern: h.pattern ?? "",
            roleCode: h.roleCode ?? "",
            scopeKind: String(h.scopeKind ?? ""),
            policies: (h.policies ?? []).map((p) => ({
              kind: p.kind ?? "",
              name: p.name ?? "",
              state: p.state ?? "",
            })),
          })),
        };

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <div className="flex flex-col gap-4 lg:flex-row">
          <div className="flex flex-1 flex-col gap-1">
            <span className="font-medium text-[13px]">{m.system_diag_subject()}</span>
            <Select
              aria-label={m.system_diag_subject()}
              className="w-full"
              onChange={onSubject}
              options={userOptions}
              placeholder={m.system_diag_ph_subject()}
              showSearch={{
                filterOption: (input, option) =>
                  (option?.search ?? "").toLowerCase().includes(input.toLowerCase()),
              }}
              value={subjectId}
            />
          </div>
          <div className="flex flex-1 flex-col gap-1">
            <span className="font-medium text-[13px]">{m.system_diag_operation()}</span>
            <OperationSelect
              catalog={catalogQuery.data?.list ?? []}
              onChange={(next) => {
                setOperation(next);
                explainMut.reset();
              }}
              value={operation}
            />
          </div>
          <div className="flex flex-col justify-end gap-1">
            <Button
              block
              color="primary"
              disabled={subjectId === undefined || operation === undefined}
              icon={<Play size={14} />}
              loading={explainMut.isPending}
              onClick={runExplain}
              variant="filled"
            >
              {m.system_diag_run()}
            </Button>
          </div>
        </div>
      </Card>
      <div className="flex flex-col gap-4">
        <Card title={m.system_diag_verdict_title()}>
          {explainView === null ? (
            <Typography.Text type="secondary">{m.system_diag_no_result()}</Typography.Text>
          ) : (
            <VerdictCard result={explainView} roleByCode={roleByCode} />
          )}
          <Typography.Text className="text-[12px]" type="secondary">
            {m.system_diag_note_realtime()}
          </Typography.Text>
        </Card>
        {subjectId !== undefined && (
          <Card title={m.system_diag_effective()}>
            <GrantPreview
              catalog={catalog}
              loading={previewQuery.isFetching || principalRolesQuery.isFetching}
              operations={previewQuery.data?.operations ?? []}
            />
            <Typography.Text className="text-[12px]" type="secondary">
              {m.system_diag_note_preview()}
            </Typography.Text>
          </Card>
        )}
        {session.data?.status === "authed" && (
          <Card>
            <SessionPanel email={session.data.user.email} permissions={session.data.permissions} />
          </Card>
        )}
      </div>
    </div>
  );
}

import { Result, Steps, Tag } from "antd";
import { m } from "#/paraglide/messages";
import { scopeKindLabel, toScopeKind } from "../../system-roles/grants";
import { type ExplainPolicyView, PolicyChip } from "./policy-chip";

interface ExplainHitView {
  pattern: string;
  roleCode: string;
  scopeKind: string;
  policies: ExplainPolicyView[];
}

export interface ExplainView {
  verdict: string;
  roleCodes: string[];
  hits: ExplainHitView[];
}

interface RoleMeta {
  name: string;
  enabled: boolean;
}

interface VerdictCardProps {
  result: ExplainView;
  roleByCode: Map<string, RoleMeta>;
}

export function VerdictCard({ result, roleByCode }: Readonly<VerdictCardProps>) {
  const allowed = result.verdict === "ALLOW";
  const roleName = (code: string) => roleByCode.get(code)?.name ?? code;
  const allPolicies = result.hits.flatMap((hit) => hit.policies);
  const attached = allPolicies.length;
  const failed = allPolicies.filter((p) => p.state === "failed");
  return (
    <div className="flex flex-col gap-2">
      <Result
        status={allowed ? "success" : "error"}
        subTitle={
          allowed
            ? m.system_diag_step_verdict_allow()
            : failed.length > 0
              ? m.system_diag_step_verdict_blocked({ name: failed[0]?.name ?? "" })
              : m.system_diag_step_verdict_deny()
        }
        title={allowed ? m.system_diag_verdict_allow() : m.system_diag_verdict_deny()}
      />
      <Steps
        current={3}
        items={[
          {
            title: m.system_diag_step_auth(),
            content: m.system_diag_step_auth_desc(),
          },
          {
            title: m.system_diag_step_bindings(),
            content:
              result.roleCodes.length === 0 ? (
                m.system_diag_step_bindings_none()
              ) : (
                <span className="flex flex-wrap gap-1">
                  {result.roleCodes.map((code) => {
                    const meta = roleByCode.get(code);
                    return (
                      <Tag className={meta && !meta.enabled ? "opacity-50" : ""} key={code}>
                        {roleName(code)}
                        {meta && !meta.enabled ? ` · ${m.system_diag_role_disabled()}` : ""}
                      </Tag>
                    );
                  })}
                </span>
              ),
          },
          {
            title: m.system_diag_step_match(),
            content:
              result.hits.length === 0 ? (
                m.system_diag_step_match_none()
              ) : (
                <span className="flex flex-col gap-1 pt-1">
                  {result.hits.map((hit) => (
                    <span
                      className="flex flex-wrap items-center gap-2"
                      key={`${hit.roleCode}${hit.pattern}`}
                    >
                      <span className="rounded bg-black/5 px-1.5 py-0.5 font-mono text-[12px] dark:bg-black/40">
                        {hit.pattern}
                      </span>
                      <Tag>
                        ← {roleName(hit.roleCode)}
                        {` · ${scopeKindLabel(toScopeKind(hit.scopeKind))}`}
                      </Tag>
                      {hit.policies.map((policy) => (
                        <PolicyChip key={policy.name} policy={policy} />
                      ))}
                    </span>
                  ))}
                </span>
              ),
          },
          {
            status: allowed ? "finish" : "error",
            title: m.system_diag_step_verdict(),
            content:
              allowed && attached > 0
                ? m.system_diag_step_verdict_all_passed({ n: attached })
                : undefined,
          },
        ]}
        orientation="vertical"
        size="small"
      />
    </div>
  );
}

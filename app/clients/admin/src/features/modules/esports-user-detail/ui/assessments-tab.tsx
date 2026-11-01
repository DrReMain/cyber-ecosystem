import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { useQuery } from "@connectrpc/connect-query";
import type {
  ActivationCode,
  ReportPolicyRow,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { ActivationCodeStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import {
  getMembership,
  listActivationCodes,
  listReportAccessPolicies,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import type { TestAttemptRow } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin_pb";
import {
  listTestAttempts,
  listTests,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin-TestAdminService_connectquery";
import { DataTable } from "@cyber-ecosystem/shared-antd/table";
import { Card, Table, Tag, Typography } from "antd";
import { useMemo } from "react";
import { dateOf, fullDate, fullTime } from "#/features/modules/esports-members/ui/dates";
import { isLive } from "#/features/modules/esports-members/ui/members-table";
import { m } from "#/paraglide/messages";

// The six-word ability vocabulary the catalog carries — labels live
// client-side, the words themselves are seed data.
const ABILITY_LABELS: Record<string, () => string> = {
  reactionSpeed: m.esports_tests_ability_reaction,
  aimingAccuracy: m.esports_tests_ability_aiming,
  trackingStability: m.esports_tests_ability_tracking,
  skillReleaseAccuracy: m.esports_tests_ability_skill,
  resourceAllocation: m.esports_tests_ability_resource,
  decisionEfficiency: m.esports_tests_ability_decision,
};

const abilityLabel = (ability?: string) =>
  ability ? (ABILITY_LABELS[ability]?.() ?? ability) : "-";

const CATEGORY_LABELS: Record<string, () => string> = {
  fps: m.esports_tests_category_fps,
  moba: m.esports_tests_category_moba,
  universal: m.esports_tests_category_universal,
};

const categoryLabel = (category: string) => CATEGORY_LABELS[category]?.() ?? category;

const gradeOf = (score: number) =>
  score >= 90
    ? { label: m.esports_user_detail_grade_s(), color: "gold" as const }
    : score >= 80
      ? { label: m.esports_user_detail_grade_a(), color: "success" as const }
      : score >= 70
        ? { label: m.esports_user_detail_grade_b(), color: "processing" as const }
        : { label: m.esports_user_detail_grade_c(), color: "error" as const };

// The report families in display order, keyed by the transport layer's
// wire names — the row fields carry those strings, not the enum members
// (the vocab table's own precedent).
const REPORT_FAMILIES = [
  {
    key: "REPORT_TYPE_COGNITIVE",
    label: () => m.esports_report_type_cognitive(),
  },
  {
    key: "REPORT_TYPE_CAREER",
    label: () => m.esports_report_type_career(),
  },
  {
    key: "REPORT_TYPE_SCHOOL",
    label: () => m.esports_report_type_school(),
  },
];

const POLICY_LABELS: Record<string, () => string> = {
  REPORT_ACCESS_POLICY_OPEN: m.esports_user_detail_policy_open,
  REPORT_ACCESS_POLICY_MEMBERSHIP: m.esports_user_detail_policy_membership,
  REPORT_ACCESS_POLICY_TENANT: m.esports_user_detail_policy_tenant,
  REPORT_ACCESS_POLICY_CLOSED: m.esports_user_detail_policy_closed,
};

const mean = (xs: number[]) => (xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : 0);

// Catalog join: the attempt row stores the code, the name and ability ride
// the catalog — the row's own comment says so.
interface Catalog {
  name: (code: string) => string;
  ability: (code: string) => string | undefined;
  category: (code: string) => string | undefined;
}

export function AssessmentsTab({ userId }: Readonly<{ userId: string }>) {
  // A member's attempt volume is bounded by the seven-day window and the
  // retest budget — dozens of rows, never pages. One wide page fetches
  // the whole set; 100 is the paginate floor's own ceiling, so this rides
  // the family law without stretching it.
  const attemptsQuery = useQuery(listTestAttempts, {
    userId,
    page: { pageNo: 1, pageSize: 100 },
  });
  const testsQuery = useQuery(listTests, {});
  const attempts = attemptsQuery.data?.attempts ?? [];

  const catalog = useMemo<Catalog>(() => {
    const byCode = new Map<string, { name?: string; ability?: string; category?: string }>();
    for (const t of testsQuery.data?.tests ?? []) {
      if (t.code) byCode.set(t.code, { name: t.name, ability: t.ability, category: t.category });
    }
    return {
      name: (code: string) => byCode.get(code)?.name ?? code,
      ability: (code: string) => byCode.get(code)?.ability,
      category: (code: string) => byCode.get(code)?.category,
    };
  }, [testsQuery.data]);

  const sessions = useMemo(() => buildSessions(attempts, catalog), [attempts, catalog]);
  const abilities = useMemo(() => buildAbilities(attempts, catalog), [attempts, catalog]);
  const detail = useMemo(() => buildDetail(attempts), [attempts]);

  return (
    <div className="flex flex-col gap-4">
      <SessionCard loading={attemptsQuery.isFetching} rows={sessions} />
      <AbilityCard rows={abilities} />
      <DetailCard
        catalog={catalog}
        loading={attemptsQuery.isFetching}
        onRefresh={() => void attemptsQuery.refetch()}
        rows={detail}
      />
      <UnlockCard userId={userId} />
    </div>
  );
}

// ── derived views ───────────────────────────────────────────────────────

interface SessionRow {
  id: string;
  categories: string[];
  // The earliest attempt's raw timestamp — the presenter renders it with
  // the shared day formatter.
  startedAt: Timestamp | undefined;
  items: number;
  tries: number;
  abilityScores: { ability: string; score: number }[];
  // Null while no attempt in the session carries an authoritative score —
  // a zero or a grade cut on nothing would read as a verdict.
  composite: number | null;
}

// A session here is the attempt rows' own orchestration trace — rows
// sharing a match_session_id. Local-practice rows (no session) never form
// one; they stay visible in the detail table. The prototype's done/total
// progress and "in progress" state need the session's planned steps, a
// member-face fact this plane cannot see, so the honest columns are items
// and tries.
function buildSessions(attempts: TestAttemptRow[], catalog: Catalog): SessionRow[] {
  const groups = new Map<string, TestAttemptRow[]>();
  for (const a of attempts) {
    if (a.matchSessionId == null) continue;
    const rows = groups.get(a.matchSessionId) ?? [];
    rows.push(a);
    groups.set(a.matchSessionId, rows);
  }
  const rows: SessionRow[] = [];
  for (const [id, group] of groups) {
    const scored = group.filter((a) => a.valid && a.authoritativeScore != null);
    const codes = new Set(group.map((a) => a.testCode));
    const categories = new Set(
      group
        .map((a) => catalog.category(a.testCode))
        .filter(Boolean)
        .map((c) => c as string),
    );
    const perAbility = new Map<string, number[]>();
    for (const a of scored) {
      const ability = catalog.ability(a.testCode);
      if (!ability) continue;
      const arr = perAbility.get(ability) ?? [];
      arr.push(a.authoritativeScore as number);
      perAbility.set(ability, arr);
    }
    const starts = group
      .map((a) => dateOf(a.startedAt))
      .filter(Boolean)
      .map((d) => d as Date)
      .sort((x, y) => x.getTime() - y.getTime());
    const earliest = starts.at(0);
    const earliestRow = group.find(
      (a) => dateOf(a.startedAt) != null && dateOf(a.startedAt)?.getTime() === earliest?.getTime(),
    );
    rows.push({
      id,
      categories: [...categories],
      startedAt: earliestRow?.startedAt,
      items: codes.size,
      tries: group.length,
      abilityScores: [...perAbility.entries()]
        .map(([ability, xs]) => ({ ability, score: mean(xs) }))
        .sort((x, y) => y.score - x.score),
      composite: scored.length > 0 ? mean(scored.map((a) => a.authoritativeScore as number)) : null,
    });
  }
  return rows.sort(
    (x, y) => (dateOf(y.startedAt)?.getTime() ?? 0) - (dateOf(x.startedAt)?.getTime() ?? 0),
  );
}

interface AbilityRow {
  ability: string;
  count: number;
  meanScore: number;
  max: number;
  lift: number | null;
  lastAt: Timestamp | undefined;
}

// Per-ability fold over the scored rows in time order — the count/mean/max
// the prototype shows plus a lift (latest minus first), the honest
// progress number derivable from the rows themselves.
function buildAbilities(attempts: TestAttemptRow[], catalog: Catalog): AbilityRow[] {
  const scored = attempts
    .filter((a) => a.valid && a.authoritativeScore != null)
    .map((a) => ({
      ability: catalog.ability(a.testCode),
      score: a.authoritativeScore as number,
      at: dateOf(a.endedAt) ?? dateOf(a.startedAt),
      endedAt: a.endedAt,
    }))
    .sort((x, y) => (x.at?.getTime() ?? 0) - (y.at?.getTime() ?? 0));
  const groups = new Map<string, { score: number; endedAt: TestAttemptRow["endedAt"] }[]>();
  for (const row of scored) {
    if (!row.ability) continue;
    const arr = groups.get(row.ability) ?? [];
    arr.push({ score: row.score, endedAt: row.endedAt });
    groups.set(row.ability, arr);
  }
  return [...groups.entries()]
    .map(([ability, xs]) => {
      const first = xs.at(0);
      const last = xs.at(-1);
      return {
        ability,
        count: xs.length,
        meanScore: mean(xs.map((x) => x.score)),
        max: Math.max(...xs.map((x) => x.score)),
        lift: first && last && xs.length > 1 ? last.score - first.score : null,
        lastAt: last?.endedAt,
      };
    })
    .sort((x, y) => y.count - x.count);
}

interface DetailRow {
  attempt: TestAttemptRow;
  lift: number | null;
}

// The flat ledger: every attempt row, newest first, each carrying its lift
// against the previous attempt of the same test.
function buildDetail(attempts: TestAttemptRow[]): DetailRow[] {
  // Fold oldest first so "previous" is the earlier attempt of the pair —
  // the newer-then-older walk would diff against the future.
  const ordered = [...attempts].sort(
    (x, y) => (dateOf(x.endedAt)?.getTime() ?? 0) - (dateOf(y.endedAt)?.getTime() ?? 0),
  );
  const prevScore = new Map<string, number>();
  const rows: DetailRow[] = [];
  for (const a of ordered) {
    if (a.valid && a.authoritativeScore != null) {
      const prev = prevScore.get(a.testCode);
      rows.push({ attempt: a, lift: prev == null ? null : a.authoritativeScore - prev });
      prevScore.set(a.testCode, a.authoritativeScore);
    } else {
      rows.push({ attempt: a, lift: null });
    }
  }
  return rows.reverse();
}

// ── presenters ──────────────────────────────────────────────────────────

function SessionCard({ loading, rows }: Readonly<{ loading: boolean; rows: SessionRow[] }>) {
  return (
    <Card
      extra={
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_user_detail_sess_note()}
        </Typography.Text>
      }
      size="small"
      title={m.esports_user_detail_sess_title()}
    >
      <Table<SessionRow>
        columns={[
          {
            title: m.esports_user_detail_sess_col_session(),
            dataIndex: "id",
            render: (_, r) => <span className="font-mono text-xs">{r.id.slice(0, 8)}</span>,
          },
          {
            title: m.esports_user_detail_sess_col_type(),
            render: (_, r) =>
              r.categories.map((c) => (
                <Tag className="m-0" key={c}>
                  {categoryLabel(c)}
                </Tag>
              )),
          },
          {
            title: m.esports_user_detail_sess_col_time(),
            render: (_, r) =>
              r.startedAt ? <span className="tabular-nums">{fullDate(r.startedAt)}</span> : "-",
          },
          {
            title: m.esports_user_detail_sess_col_items(),
            render: (_, r) => (
              <span className="tabular-nums">
                {r.items}
                {r.tries > r.items
                  ? ` · ${m.esports_user_detail_sess_retests({ n: r.tries - r.items })}`
                  : ""}
              </span>
            ),
          },
          {
            title: m.esports_user_detail_sess_col_abilities(),
            render: (_, r) => (
              <span className="text-xs">
                {r.abilityScores
                  .map((a) => `${abilityLabel(a.ability)} ${a.score.toFixed(1)}`)
                  .join(" · ")}
              </span>
            ),
          },
          {
            title: m.esports_user_detail_sess_col_score(),
            render: (_, r) =>
              r.composite == null ? (
                <Typography.Text type="secondary">-</Typography.Text>
              ) : (
                <span className="font-semibold tabular-nums">{r.composite.toFixed(1)}</span>
              ),
          },
          {
            title: m.esports_user_detail_sess_col_grade(),
            render: (_, r) => {
              if (r.composite == null) {
                return <Typography.Text type="secondary">-</Typography.Text>;
              }
              const g = gradeOf(r.composite);
              return (
                <Tag className="m-0" color={g.color}>
                  {g.label}
                </Tag>
              );
            },
          },
        ]}
        dataSource={rows}
        loading={loading}
        locale={{ emptyText: m.esports_user_detail_empty_attempts() }}
        pagination={false}
        rowKey={(r) => r.id}
        size="small"
      />
    </Card>
  );
}

function AbilityCard({ rows }: Readonly<{ rows: AbilityRow[] }>) {
  return (
    <Card size="small" title={m.esports_user_detail_sum_title()}>
      <Table<AbilityRow>
        columns={[
          {
            title: m.esports_user_detail_sum_col_ability(),
            render: (_, r) => abilityLabel(r.ability),
          },
          {
            title: m.esports_user_detail_sum_col_count(),
            dataIndex: "count",
            render: (_, r) => <span className="tabular-nums">{r.count}</span>,
          },
          {
            title: m.esports_user_detail_sum_col_mean(),
            render: (_, r) => <span className="tabular-nums">{r.meanScore.toFixed(1)}</span>,
          },
          {
            title: m.esports_user_detail_sum_col_max(),
            render: (_, r) => <span className="tabular-nums">{r.max.toFixed(1)}</span>,
          },
          {
            title: m.esports_user_detail_sum_col_progress(),
            render: (_, r) =>
              r.lift == null ? (
                <Typography.Text type="secondary">-</Typography.Text>
              ) : (
                <span
                  className={`tabular-nums ${r.lift >= 0 ? "text-emerald-600" : "text-red-500"}`}
                >
                  {r.lift >= 0 ? "+" : ""}
                  {r.lift.toFixed(1)}
                </span>
              ),
          },
          {
            title: m.esports_user_detail_sum_col_recent(),
            render: (_, r) =>
              r.lastAt ? <span className="tabular-nums">{fullDate(r.lastAt)}</span> : "-",
          },
        ]}
        dataSource={rows}
        locale={{ emptyText: m.esports_user_detail_empty_attempts() }}
        pagination={false}
        rowKey={(r) => r.ability}
        size="small"
      />
    </Card>
  );
}

function DetailCard({
  catalog,
  loading,
  onRefresh,
  rows,
}: Readonly<{
  catalog: Catalog;
  loading: boolean;
  onRefresh: () => void;
  rows: DetailRow[];
}>) {
  return (
    <Card size="small" title={m.esports_user_detail_det_title()}>
      <DataTable<DetailRow>
        columns={[
          {
            title: m.esports_user_detail_det_col_time(),
            render: (_, r) => (
              <span className={`tabular-nums ${r.attempt.valid ? "" : "opacity-50"}`}>
                {fullTime(r.attempt.endedAt)}
              </span>
            ),
          },
          {
            title: m.esports_user_detail_det_col_test(),
            render: (_, r) => (
              <span className={r.attempt.valid ? "" : "opacity-50"}>
                {catalog.name(r.attempt.testCode)}
              </span>
            ),
          },
          {
            title: m.esports_user_detail_det_col_ability(),
            render: (_, r) => abilityLabel(catalog.ability(r.attempt.testCode)),
          },
          {
            title: m.esports_user_detail_det_col_score(),
            render: (_, r) => {
              if (!r.attempt.valid) {
                return (
                  <Tag className="m-0" title={r.attempt.invalidReason ?? undefined}>
                    {m.esports_user_detail_det_invalid()}
                  </Tag>
                );
              }
              if (r.attempt.authoritativeScore == null) {
                return (
                  <Tag className="m-0" color="warning">
                    {m.esports_user_detail_det_pending()}
                  </Tag>
                );
              }
              return (
                <span className="tabular-nums">{r.attempt.authoritativeScore.toFixed(1)}</span>
              );
            },
          },
          {
            title: m.esports_user_detail_det_col_raw(),
            render: (_, r) =>
              r.attempt.rawScore == null ? (
                <Typography.Text type="secondary">-</Typography.Text>
              ) : (
                <span className="tabular-nums">{r.attempt.rawScore.toFixed(0)}</span>
              ),
          },
          {
            title: m.esports_user_detail_det_col_grade(),
            render: (_, r) =>
              r.attempt.grade ? (
                <Tag className="m-0">{r.attempt.grade}</Tag>
              ) : (
                <Typography.Text type="secondary">-</Typography.Text>
              ),
          },
          {
            title: m.esports_user_detail_det_col_lift(),
            render: (_, r) =>
              r.lift == null ? (
                <Typography.Text type="secondary">-</Typography.Text>
              ) : (
                <span
                  className={`tabular-nums ${r.lift >= 0 ? "text-emerald-600" : "text-red-500"}`}
                >
                  {r.lift >= 0 ? "+" : ""}
                  {r.lift.toFixed(1)}
                </span>
              ),
          },
        ]}
        dataSource={rows}
        emptyDescription={m.esports_user_detail_empty_attempts()}
        loading={loading}
        pagination={false}
        rowKey={(r) => r.attempt.recordId}
        toolbar={{
          labels: {
            density: m.common_toolbar_density(),
            densityLarge: m.common_toolbar_density_large(),
            densityMiddle: m.common_toolbar_density_middle(),
            densitySmall: m.common_toolbar_density_small(),
            refresh: m.common_toolbar_refresh(),
          },
          loading,
          onRefresh,
        }}
      />
    </Card>
  );
}

// Three facts side by side — the family's policy dial, the codes this
// member redeemed, the live membership — without re-folding them into a
// verdict: the member-facing access judgement owns that fold, and the
// operator reads the facts it was cut from.
function UnlockCard({ userId }: Readonly<{ userId: string }>) {
  const policiesQuery = useQuery(listReportAccessPolicies, {});
  const redeemedQuery = useQuery(listActivationCodes, {
    userId,
    codeStatus: ActivationCodeStatus.REDEEMED,
    page: { pageNo: 1, pageSize: 50 },
  });
  const mshipQuery = useQuery(getMembership, { userId }, { throwOnError: false });
  const redeemed = redeemedQuery.data?.codes ?? [];
  const membership = mshipQuery.data?.membership;

  return (
    <Card size="small" title={m.esports_user_detail_unlock_title()}>
      {REPORT_FAMILIES.map((family) => (
        <UnlockRow
          codes={redeemed.filter((c) => String(c.reportType) === family.key)}
          key={family.key}
          label={family.label()}
          membershipLive={membership != null && isLive(membership)}
          policy={(policiesQuery.data?.policies ?? []).find(
            (p) => String(p.reportType) === family.key,
          )}
        />
      ))}
    </Card>
  );
}

function UnlockRow({
  codes,
  label,
  membershipLive,
  policy,
}: Readonly<{
  codes: ActivationCode[];
  label: string;
  membershipLive: boolean;
  policy?: ReportPolicyRow;
}>) {
  const mine = [...codes].sort(
    (x, y) => (dateOf(y.redeemedAt)?.getTime() ?? 0) - (dateOf(x.redeemedAt)?.getTime() ?? 0),
  );
  const latest = mine.at(0);
  return (
    <div className="flex flex-wrap items-center gap-2 py-1.5">
      <Typography.Text className="w-24 shrink-0" type="secondary">
        {label}
      </Typography.Text>
      {latest ? (
        <>
          <Tag className="m-0 font-mono" color="purple">
            {latest.code}
          </Tag>
          <span className="text-xs tabular-nums">
            {latest.redeemedAt ? fullDate(latest.redeemedAt) : ""}
          </span>
          <Typography.Text className="text-xs" type="secondary">
            {latest.expiresAt
              ? m.esports_user_detail_unlock_until({ date: fullDate(latest.expiresAt) })
              : m.esports_user_detail_unlock_permanent()}
            {mine.length > 1
              ? ` · ${m.esports_user_detail_unlock_codes_total({ n: mine.length })}`
              : ""}
          </Typography.Text>
        </>
      ) : (
        <Typography.Text type="secondary">{m.esports_user_detail_unlock_none()}</Typography.Text>
      )}
      {membershipLive ? (
        <Tag className="m-0" color="success">
          {m.esports_user_detail_unlock_with_member()}
        </Tag>
      ) : null}
      {policy ? (
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_user_detail_unlock_policy({
            policy: POLICY_LABELS[String(policy.policy)]?.() ?? String(policy.policy),
          })}
        </Typography.Text>
      ) : null}
    </div>
  );
}

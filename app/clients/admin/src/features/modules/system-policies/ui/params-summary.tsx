import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { Typography } from "antd";
import { m } from "#/paraglide/messages";
import { type DaysSummary, daysSummary, draftOf } from "../params";
import { dayLabels } from "./day-labels";

function daysText(summary: DaysSummary): string {
  if (summary.kind === "every") return m.system_policies_days_every();
  if (summary.kind === "weekdays") return m.system_policies_days_weekdays();
  if (summary.kind === "weekend") return m.system_policies_days_weekend();
  return summary.days.map((d) => dayLabels()[d]).join(" ");
}

export function PolicyParamsSummary({ policy }: Readonly<{ policy: Policy }>) {
  const draft = draftOf(policy);
  if (draft === null) {
    return <Typography.Text type="secondary">-</Typography.Text>;
  }
  if (draft.kind === "time_window") {
    const spans = draft.timeWindow.windows.map((w) => `${w.start}–${w.end}`);
    return (
      <span className="font-mono text-[12.5px]">
        {spans.join(` ${m.system_policies_or()} `)}
        {` · ${daysText(daysSummary(draft.timeWindow.days))}`}
      </span>
    );
  }
  const base =
    draft.calendar.base === "ALL"
      ? m.system_policies_cal_base_all()
      : draft.calendar.base === "WEEKDAYS"
        ? m.system_policies_cal_base_weekdays()
        : m.system_policies_cal_base_none();
  const n = draft.calendar.overrides.length;
  return (
    <span>
      {base}
      {n > 0 && ` + ${m.system_policies_cal_overrides({ n })}`}
    </span>
  );
}

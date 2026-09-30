import { Button, DatePicker, Segmented, Typography } from "antd";
import dayjs from "dayjs";
import { Plus, X } from "lucide-react";
import { m } from "#/paraglide/messages";
import { type CalendarDraft, dateKey, duplicateDates, isDateValid, passesNow } from "../params";

interface PolicyCalendarFormProps {
  draft: CalendarDraft;
  onChange: (next: CalendarDraft) => void;
}

const baseOptions = () => [
  { label: m.system_policies_cal_base_all(), value: "ALL" },
  { label: m.system_policies_cal_base_weekdays(), value: "WEEKDAYS" },
  { label: m.system_policies_cal_base_none(), value: "NONE" },
];

function baseHint(base: CalendarDraft["base"]): string {
  if (base === "ALL") return m.system_policies_cal_base_all_hint();
  if (base === "WEEKDAYS") return m.system_policies_cal_base_weekdays_hint();
  return m.system_policies_cal_base_none_hint();
}

export function PolicyCalendarForm({ draft, onChange }: Readonly<PolicyCalendarFormProps>) {
  const dups = duplicateDates(draft.overrides);
  const valid = dups.size === 0 && draft.overrides.every((o) => isDateValid(o.date));
  const active = valid && passesNow({ kind: "calendar", calendar: draft }, new Date());

  const setOverride = (index: number, date: string, allowed: boolean) => {
    const overrides = draft.overrides.map((o, i) => (i === index ? { date, allowed } : o));
    onChange({ ...draft, overrides });
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="font-medium text-[13px]">{m.system_policies_cal_base()}</span>
        {valid ? (
          <Typography.Text className="text-[12px]" type={active ? "success" : "secondary"}>
            {active ? m.system_policies_now_active() : m.system_policies_now_inactive()}
          </Typography.Text>
        ) : null}
      </div>
      <Segmented
        block
        onChange={(v) => onChange({ ...draft, base: v as CalendarDraft["base"] })}
        options={baseOptions()}
        value={draft.base}
      />
      <Typography.Text className="text-[12px]" type="secondary">
        {baseHint(draft.base)}
      </Typography.Text>
      <div className="flex flex-col gap-2 border-black/8 border-t pt-3 dark:border-white/8">
        <span className="font-medium text-[13px]">{m.system_policies_cal_overrides_label()}</span>
        {draft.overrides.map((o, i) => {
          const bad = dups.has(o.date) || !isDateValid(o.date);
          return (
            // biome-ignore lint/suspicious/noArrayIndexKey: rows hold no local state and the draft array is replaced wholesale, so index reuse cannot leak; date keys would collide before the row is filled in
            <div className="flex flex-col gap-1" key={i}>
              <div className="flex items-center gap-2">
                <DatePicker
                  allowClear={false}
                  aria-label={m.system_policies_cal_overrides_label()}
                  onChange={(d) => setOverride(i, d?.format("YYYY-MM-DD") ?? o.date, o.allowed)}
                  status={bad ? "error" : undefined}
                  value={o.date ? dayjs(o.date) : null}
                />
                <Segmented
                  onChange={(v) => setOverride(i, o.date, v === "allow")}
                  options={[
                    { label: m.system_policies_cal_override_allow(), value: "allow" },
                    { label: m.system_policies_cal_override_deny(), value: "deny" },
                  ]}
                  value={o.allowed ? "allow" : "deny"}
                />
                <Button
                  aria-label={m.system_policies_tw_remove()}
                  icon={<X size={14} />}
                  onClick={() =>
                    onChange({ ...draft, overrides: draft.overrides.filter((_, j) => j !== i) })
                  }
                  size="small"
                  type="text"
                />
              </div>
              {dups.has(o.date) ? (
                <Typography.Text className="text-[12px]" type="danger">
                  {m.system_policies_cal_date_duplicate()}
                </Typography.Text>
              ) : null}
            </div>
          );
        })}
        <div>
          <Button
            color="default"
            icon={<Plus size={14} />}
            onClick={() =>
              onChange({
                ...draft,
                overrides: [...draft.overrides, { date: dateKey(new Date()), allowed: false }],
              })
            }
            size="small"
            variant="filled"
          >
            {m.system_policies_cal_add()}
          </Button>
        </div>
      </div>
    </div>
  );
}

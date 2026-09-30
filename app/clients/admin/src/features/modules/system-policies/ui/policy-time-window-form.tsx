import { Button, Checkbox, Typography } from "antd";
import { Plus } from "lucide-react";
import { m } from "#/paraglide/messages";
import { DAY_ORDER, isSpanValid, passesNow, type TimeWindowDraft } from "../params";
import { dayLabels } from "./day-labels";
import { WindowRow } from "./window-row";

interface PolicyTimeWindowFormProps {
  draft: TimeWindowDraft;
  onChange: (next: TimeWindowDraft) => void;
}

export function PolicyTimeWindowForm({ draft, onChange }: Readonly<PolicyTimeWindowFormProps>) {
  const setWindow = (index: number, start: string, end: string) => {
    const windows = draft.windows.map((w, i) => (i === index ? { start, end } : w));
    onChange({ ...draft, windows });
  };
  const removeWindow = (index: number) => {
    onChange({ ...draft, windows: draft.windows.filter((_, i) => i !== index) });
  };
  const addWindow = () => {
    onChange({ ...draft, windows: [...draft.windows, { start: "09:00", end: "18:00" }] });
  };
  // Plain checkboxes, zero side effects: empty selection means "every day"
  // (the wire contract) and the static hint says so - no auto-fill, no mode
  // jump. Save-side normalization folds an all-seven selection back to [].
  const toggleDay = (day: number, on: boolean) => {
    const days = on ? [...new Set([...draft.days, day])] : draft.days.filter((d) => d !== day);
    onChange({ ...draft, days });
  };

  const valid = draft.windows.length > 0 && draft.windows.every(isSpanValid);
  const active = valid && passesNow({ kind: "time_window", timeWindow: draft }, new Date());

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="font-medium text-[13px]">{m.system_policies_tw_windows()}</span>
        {valid ? (
          <Typography.Text className="text-[12px]" type={active ? "success" : "secondary"}>
            {active ? m.system_policies_now_active() : m.system_policies_now_inactive()}
          </Typography.Text>
        ) : null}
      </div>
      <div className="flex flex-col gap-0.5">
        <Typography.Text className="text-[12px]" type="secondary">
          {m.system_policies_tw_or_hint()}
        </Typography.Text>
        <Typography.Text className="text-[12px]" type="secondary">
          {m.system_policies_tw_midnight_hint()}
        </Typography.Text>
      </div>
      {draft.windows.map((w, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: rows hold no local state and the draft array is replaced wholesale, so index reuse cannot leak; content keys would remount the picker on every keystroke
        <WindowRow index={i} key={i} onChange={setWindow} onRemove={removeWindow} span={w} />
      ))}
      <div className="flex flex-col gap-1">
        <div>
          <Button
            color="default"
            icon={<Plus size={14} />}
            onClick={addWindow}
            size="small"
            variant="filled"
          >
            {m.system_policies_tw_add()}
          </Button>
        </div>
        {draft.windows.length === 0 ? (
          <Typography.Text className="text-[12px]" type="danger">
            {m.system_policies_tw_windows_required()}
          </Typography.Text>
        ) : null}
      </div>
      <div className="flex flex-col gap-2 border-black/8 border-t pt-3 dark:border-white/8">
        <div className="flex items-center justify-center gap-1">
          {DAY_ORDER.map((d) => (
            <Checkbox
              checked={draft.days.includes(d)}
              key={d}
              onChange={(e) => toggleDay(d, e.target.checked)}
            >
              {dayLabels()[d]}
            </Checkbox>
          ))}
        </div>
        <Typography.Text className="text-[12px]" type="secondary">
          {m.system_policies_tw_days_hint()}
        </Typography.Text>
      </div>
    </div>
  );
}

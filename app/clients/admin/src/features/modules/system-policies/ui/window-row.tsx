import { Button, Tag, TimePicker, Typography } from "antd";
import clsx from "clsx";
import dayjs from "dayjs";
import { X } from "lucide-react";
import { m } from "#/paraglide/messages";
import { isSpanValid, type WindowDraft } from "../params";

interface WindowRowProps {
  index: number;
  span: WindowDraft;
  onChange: (index: number, start: string, end: string) => void;
  onRemove: (index: number) => void;
}

// One same-day [start, end) span; "24:00" is the end-of-day sentinel a wall
// clock cannot express.
export function WindowRow({ index, span, onChange, onRemove }: Readonly<WindowRowProps>) {
  const bad = !isSpanValid(span);
  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={clsx(
            "text-[12px]",
            index > 0 ? "text-ink-tertiary" : "select-none text-transparent",
          )}
        >
          {m.system_policies_or()}
        </span>
        <TimePicker
          allowClear={false}
          aria-label={m.system_policies_tw_start()}
          format="HH:mm"
          needConfirm={false}
          onChange={(t) => onChange(index, t?.format("HH:mm") ?? span.start, span.end)}
          status={bad ? "error" : undefined}
          value={span.start ? dayjs(span.start, "HH:mm") : null}
        />
        <span>–</span>
        {span.end === "24:00" ? (
          <Tag closable onClose={() => onChange(index, span.start, "23:59")}>
            {`24:00 · ${m.system_policies_tw_end_of_day()}`}
          </Tag>
        ) : (
          <TimePicker
            allowClear={false}
            aria-label={m.system_policies_tw_end()}
            format="HH:mm"
            needConfirm={false}
            onChange={(t) => onChange(index, span.start, t?.format("HH:mm") ?? span.end)}
            status={bad ? "error" : undefined}
            value={span.end ? dayjs(span.end, "HH:mm") : null}
          />
        )}
        {span.end !== "24:00" ? (
          <Button onClick={() => onChange(index, span.start, "24:00")} size="small" type="link">
            {m.system_policies_tw_end_of_day()}
          </Button>
        ) : null}
        <Button
          aria-label={m.system_policies_tw_remove()}
          icon={<X size={14} />}
          onClick={() => onRemove(index)}
          size="small"
          type="text"
        />
      </div>
      {bad ? (
        <Typography.Text className="text-[12px]" type="danger">
          {m.system_policies_tw_span_invalid()}
        </Typography.Text>
      ) : null}
    </div>
  );
}
